package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jibei/scouter/internal/admin"
	"github.com/jibei/scouter/internal/autotag"
	"github.com/jibei/scouter/internal/benchmark"
	"github.com/jibei/scouter/internal/dealexplain"
	"github.com/jibei/scouter/internal/optimizer"
	"github.com/jibei/scouter/internal/digest"
	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/coach"
	"github.com/jibei/scouter/internal/collaborator"
	"github.com/jibei/scouter/internal/comment"
	"github.com/jibei/scouter/internal/comparison"
	"github.com/jibei/scouter/internal/config"
	"github.com/jibei/scouter/internal/db"
	"github.com/jibei/scouter/internal/dealcalendar"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/export"
	"github.com/jibei/scouter/internal/forecast"
	"github.com/jibei/scouter/internal/health"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/metrics"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/negotiation"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/persona"
	"github.com/jibei/scouter/internal/prediction"
	"github.com/jibei/scouter/internal/pricecomp"
	"github.com/jibei/scouter/internal/product"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/purchase"
	"github.com/jibei/scouter/internal/research"
	"github.com/jibei/scouter/internal/reviews"
	"github.com/jibei/scouter/internal/timing"
	"github.com/jibei/scouter/internal/scheduler"
	"github.com/jibei/scouter/internal/search"
	"github.com/jibei/scouter/internal/settings"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/stockcheck"
	"github.com/jibei/scouter/internal/substitute"
	"github.com/jibei/scouter/internal/summary"
	"github.com/jibei/scouter/internal/template"
	"github.com/jibei/scouter/internal/travel"
	"github.com/jibei/scouter/internal/usage"
	"github.com/jibei/scouter/internal/receipt"
	"github.com/jibei/scouter/internal/vote"
	"github.com/jibei/scouter/internal/wishlist"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config error", "err", err)
		os.Exit(1)
	}

	// Run database migrations
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Error("migration failed", "err", err)
		os.Exit(1)
	}
	log.Info("migrations applied")

	// Set up signal-aware context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to database
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	log.Info("database connected")

	// Build metrics recorder (Phase 14)
	var rec metrics.Recorder = metrics.NoopRecorder{}
	var metricsHandler http.Handler
	if cfg.MetricsEnabled {
		promRec, promReg := metrics.NewPrometheusRecorder(pool)
		rec = promRec
		metricsHandler = promhttp.HandlerFor(promReg, promhttp.HandlerOpts{})
		log.Info("metrics enabled", "endpoint", "/metrics")
	}

	// Build LLM provider
	provider, smartRouter := buildSmartRouter(cfg, log, rec)

	// Repositories
	missionRepo := mission.NewRepository(pool)
	optionRepo := option.NewRepository(pool)
	shoppingRepo := shopping.NewRepository(pool)
	notifRepo := notification.NewRepository(pool)
	usageRepo := usage.NewRepository(pool)
	agentRunRepo := agentrun.NewRepository(pool)

	// Services
	missionSvc := mission.NewService(missionRepo)
	optionSvc := option.NewService(optionRepo)
	shoppingSvc := shopping.NewService(shoppingRepo)
	usageSvc := usage.NewService(usageRepo)

	// Embedding worker (Phase 11)
	embedder := llm.NewOllamaEmbedder(cfg.OllamaBaseURL, cfg.OllamaEmbedModel, "", cfg.OllamaFastTimeout)
	embedRepo := embedding.NewRepository(pool)
	embedWorker := embedding.NewWorker(embedder, optionRepo, embedRepo)
	embedWorker.Start(ctx)
	// Wire embed channel into option service and research agent.
	optionSvc.WithEmbedChannel(embedWorker.Jobs())

	// Agents
	researchAgent := research.NewAgent(provider, optionRepo, agentRunRepo, usageSvc)
	researchAgent.SetEmbedChannel(embedWorker.Jobs())
	researchAgent.SetRecorder(rec)
	pricingAgent := pricing.NewAgent(provider, shoppingRepo, agentRunRepo, usageSvc)
	pricingAgent.SetRecorder(rec)
	decisionAgent := decision.NewAgent(provider, usageSvc)
	decisionAgent.SetRecorder(rec)
	coachAgent := coach.NewAgent(provider)

	// Decision
	decisionRepo := decision.NewRepository(pool)
	decisionSvc := decision.NewService(decisionRepo, decisionAgent, missionRepo, optionRepo)

	// Handlers
	missionHandler := mission.NewHandler(missionSvc)
	optionHandler := option.NewHandler(optionSvc)
	shoppingHandler := shopping.NewHandler(shoppingSvc)
	notifHandler := notification.NewHandler(notifRepo)
	researchHandler := research.NewHandler(researchAgent, missionSvc)
	pricingHandler := pricing.NewHandler(pricingAgent, missionSvc, optionRepo)
	usageHandler := usage.NewHandler(usageSvc)
	decisionHandler := decision.NewHandler(decisionSvc)
	agentRunSvc := agentrun.NewService(agentRunRepo)
	agentRunHandler := agentrun.NewHandler(agentRunSvc)
	coachHandler := coach.NewHandler(coachAgent, missionSvc, optionRepo, shoppingRepo)

	// Wish List (Phase 21) + Price Alerts (Phase 24) — declared here so
	// wishlistPriceChecker is available to the scheduler block below.
	wishlistRepo := wishlist.NewRepository(pool)
	wishlistPriceChecker := wishlist.NewPriceChecker(wishlistRepo, notifRepo, provider)

	// Price check scheduler (opt-in via PRICE_CHECK_ENABLED=true)
	var sched *scheduler.Scheduler
	if cfg.PriceCheckEnabled {
		orch := scheduler.NewOrchestrator(
			missionRepo, optionRepo, shoppingRepo, notifRepo,
			pricingAgent, cfg.PriceCheckMaxMissions, log,
		)
		orch.SetRecorder(rec)
		var schedErr error
		sched, schedErr = scheduler.New(cfg.PriceCheckCron, orch, log, wishlistPriceChecker)
		if schedErr != nil {
			log.Error("scheduler init failed", "err", schedErr)
			os.Exit(1)
		}
		sched.Start()
	}

	// Search (Phase 11)
	searchRepo := search.NewRepository(pool)
	searchHandler := search.NewHandler(searchRepo, embedder, embedRepo, embedWorker)

	// Purchase lifecycle (Phase 12)
	purchaseRepo := purchase.NewRepository(pool)
	purchaseSvc := purchase.NewService(purchaseRepo, missionRepo)
	purchaseHandler := purchase.NewHandler(purchaseSvc)
	statsHandler := purchase.NewStatsHandler(pool)

	// Settings (Phase 13)
	settingsRepo := settings.NewRepository(pool)
	settingsHandler := settings.NewHandler(settingsRepo)

	// Collaborative missions (Phase 16)
	collaboratorRepo := collaborator.NewRepository(pool)
	collaboratorSvc := collaborator.NewService(collaboratorRepo)
	collaboratorHandler := collaborator.NewHandler(collaboratorSvc)
	voteRepo := vote.NewRepository(pool)
	voteSvc := vote.NewService(voteRepo)
	voteHandler := vote.NewHandler(voteSvc)

	// Admin data management (Phase 13)
	adminHandler := admin.NewHandler(pool)

	// Export
	exportGatherer := export.NewGatherer(missionRepo, optionRepo, shoppingRepo, decisionRepo)
	exportHandler := export.NewHandler(exportGatherer)

	// Templates (no DB dependency — compiled into binary)
	templateReg := template.NewRegistry()
	templateHandler := template.NewHandler(templateReg)

	// Deal Calendar (no DB dependency — hardcoded French events)
	dealCalHandler := dealcalendar.NewHandler()

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(1 << 20)) // 1 MiB max request body
	r.Use(metrics.Middleware(rec))

	if cfg.Env == "development" {
		r.Use(corsMiddleware)
	}

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(pingCtx); err != nil {
			dbStatus = "error"
		}
		overall := "ok"
		if dbStatus != "ok" {
			overall = "degraded"
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]string{
			"status": overall,
			"db":     dbStatus,
		})
	})

	r.Get("/api/usage", usageHandler.GetSummary)

	// Mission CRUD: /api/missions and /api/missions/{slug}
	r.Get("/api/missions", missionHandler.List)
	r.Post("/api/missions", missionHandler.Create)
	r.Get("/api/missions/{slug}", missionHandler.Get)
	r.Patch("/api/missions/{slug}", missionHandler.Update)
	r.Delete("/api/missions/{slug}", missionHandler.Delete)
	r.Post("/api/missions/{slug}/duplicate", missionHandler.Duplicate)
	r.Post("/api/missions/{slug}/clone", missionHandler.Clone)

	// Mission sub-resources
	r.Mount("/api/missions/{missionID}/options", optionHandler.Routes())
	r.Mount("/api/missions/{missionID}/shopping", shoppingHandler.Routes())
	r.Mount("/api/missions/{missionID}/research", researchHandler.Routes())
	r.Mount("/api/missions/{missionID}/pricing", pricingHandler.Routes())
	r.Mount("/api/missions/{missionID}/decision", decisionHandler.Routes())
	r.Mount("/api/missions/{missionID}/agent-runs", agentRunHandler.Routes())
	r.Mount("/api/missions/{slug}/coach", coachHandler.Routes())

	// Purchase lifecycle (Phase 12)
	r.Mount("/api/missions/{missionID}/purchase", purchaseHandler.Routes())
	r.Get("/api/stats", statsHandler.Stats)
	r.Get("/api/stats/monthly", statsHandler.MonthlyStats)

	// Export, share, archive
	r.Get("/api/missions/{missionID}/export", exportHandler.Export)
	r.Post("/api/missions/{missionID}/share", missionHandler.Share)
	r.Delete("/api/missions/{missionID}/share", missionHandler.RevokeShare)
	r.Post("/api/missions/{missionID}/archive", missionHandler.Archive)
	r.Post("/api/missions/{missionID}/unarchive", missionHandler.Unarchive)

	// Public shared mission (accessible without auth, always CORS-open)
	r.With(corsMiddleware).Get("/api/shared/{token}", func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		m, err := missionSvc.GetByShareToken(r.Context(), token)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if m == nil {
			httputil.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, m)
	})

	// Semantic search (Phase 11)
	r.Get("/api/search", searchHandler.Search)
	r.Get("/api/options/{optionID}/similar", searchHandler.Similar)
	// Reindex is a heavy admin operation (up to 500 embed jobs). It is only
	// exposed in development; in production it should be triggered via CLI or a
	// protected admin route. No general auth layer exists yet.
	if cfg.Env == "development" {
		r.Post("/api/search/reindex", searchHandler.Reindex)
	}

	// Notifications
	r.Mount("/api/notifications", notifHandler.Routes())

	// Templates
	r.Get("/api/templates", templateHandler.List)
	r.Get("/api/templates/{slug}", templateHandler.Get)

	// Deal Calendar
	r.Get("/api/deal-calendar", dealCalHandler.List)

	// Settings (Phase 13)
	r.Get("/api/settings", settingsHandler.GetAll)
	r.Patch("/api/settings", settingsHandler.Update)

	// Data management (Phase 13)
	r.Delete("/api/data", adminHandler.DeleteAllData)

	// Collaborative missions (Phase 16)
	r.Mount("/api/missions/{missionID}/invites", collaboratorHandler.InviteRoutes())
	r.Mount("/api/missions/{missionID}/collaborators", collaboratorHandler.CollaboratorRoutes())
	r.Get("/api/invites/{token}", collaboratorHandler.GetInvite)
	r.Post("/api/invites/{token}/join", collaboratorHandler.JoinMission)
	r.Mount("/api/options/{optionID}/votes", voteHandler.Routes())

	// Smart Budget Forecaster (Phase 23)
	forecastRepo := forecast.NewRepository(pool)
	forecastAgent := forecast.NewForecastAgent(provider, forecastRepo, pool)
	forecastHandler := forecast.NewHandler(forecastAgent, forecastRepo)
	r.Route("/api/missions/{missionID}/forecast", forecastHandler.Routes())

	// Wish List route (repo+checker declared above, before the scheduler block)
	r.Route("/api/wishlist", wishlist.Routes(wishlistRepo))

	// Envelope Budgeting (Phase 35)
	envelopeRepo := envelope.NewRepository(pool)
	envelopeSvc := envelope.NewService(envelopeRepo)
	envelopeHandler := envelope.NewHandler(envelopeSvc)
	r.Route("/api/envelopes", envelope.Routes(envelopeHandler))

	// Spending Persona (Phase 26)
	personaRepo := persona.NewRepository(pool)
	personaAgent := persona.NewPersonaAgent(provider, personaRepo, pool)
	personaHandler := persona.NewHandler(personaAgent, personaRepo)
	r.Route("/api/persona", personaHandler.Routes())

	// Product barcode lookup (Phase 25)
	productLooker := product.NewLooker()
	r.Route("/api/products", product.Routes(productLooker))

	// Social Proof & Review Aggregation (Phase 30)
	reviewsCache := reviews.NewCache(6 * time.Hour)
	reviewsAgent := reviews.NewAgent(provider)
	reviewsGetter := reviews.NewOptionRepoGetter(optionRepo)
	reviewsHandler := reviews.NewHandler(reviewsGetter, reviewsCache, reviewsAgent)

	// Travel integrations (Phase 29)
	avClient := travel.NewAviationStackClient(cfg.AviationStackAPIKey, cfg.AviationStackBaseURL)
	sncfClient := travel.NewSNCFClient(cfg.SNCFAPIKey, cfg.SNCFBaseURL)
	travelCache := travel.NewCache(time.Hour)
	travelHandler := travel.NewHandler(avClient, sncfClient, travelCache)
	r.Mount("/api/travel", travelHandler.Routes())

	// Weighted Comparison Matrix (Phase 22)
	comparisonRepo := comparison.NewRepository(pool)
	r.Route("/api/missions/{missionID}/comparison-weights", comparison.WeightRoutes(comparisonRepo, pool))
	r.Get("/api/missions/{missionID}/matrix", comparison.NewHandler(comparisonRepo, pool).Matrix)

	// Social Proof & Review Aggregation (Phase 30)
	r.Get("/api/options/{id}/reviews", reviewsHandler.GetReviews)
	r.Post("/api/options/{id}/reviews/refresh", reviewsHandler.RefreshReviews)

	// Product Price Comparison (Phase 31)
	pricecompCache := pricecomp.NewCache(30 * time.Minute)
	pricecompAgent := pricecomp.NewAgent(provider)
	pricecompGetter := pricecomp.NewOptionRepoGetter(optionRepo)
	pricecompHandler := pricecomp.NewHandler(pricecompGetter, pricecompCache, pricecompAgent)
	r.Get("/api/options/{id}/price-comparison", pricecompHandler.GetComparison)
	r.Post("/api/options/{id}/price-comparison/refresh", pricecompHandler.RefreshComparison)

	// AI Purchase Timing Advisor (Phase 32)
	timingCache := timing.NewCache(24 * time.Hour)
	timingAgent := timing.NewAgent(provider)
	timingGetter := timing.NewMissionRepoGetter(pool)
	timingHandler := timing.NewHandler(timingGetter, timingCache, timingAgent)
	r.Post("/api/missions/{id}/timing-advice", timingHandler.GetTimingAdvice)

	// AI Negotiation Coach (Phase 17)
	negotiationHandler := negotiation.NewHandler(pool, provider)
	r.Mount("/api/options/{optionID}", negotiationHandler.Routes())

	// Smart Price Prediction (Phase 52)
	predictionHandler := prediction.NewHandler(shoppingRepo)
	r.Get("/api/shopping-items/{id}/prediction", predictionHandler.GetPrediction)

	// Smart Category Auto-Tagging (Phase 65)
	autotagAgent := autotag.NewAgent(provider)
	autotagHandler := autotag.NewHandler(missionRepo, autotagAgent)
	r.Post("/api/missions/{slug}/suggest-category", autotagHandler.Suggest)

	// AI Shopping List Optimizer (Phase 68)
	optimizerHandler := optimizer.NewHandler(pool, provider)
	r.Post("/api/missions/{slug}/optimize", optimizerHandler.Optimize)

	// Smart Receipt Analyzer (Phase 69)
	receiptRepo := receipt.NewRepository(pool)
	receiptAgent := receipt.NewAgent(provider)
	receiptHandler := receipt.NewHandler(receiptRepo, receiptAgent)
	r.Post("/api/missions/{slug}/receipts", receiptHandler.Analyze)
	r.Get("/api/missions/{slug}/receipts", receiptHandler.List)

	// AI Deal Explainer (Phase 60)
	dealExplainCache := dealexplain.NewCache(1 * time.Hour)
	dealExplainAgent := dealexplain.NewAgent(provider)
	dealExplainHandler := dealexplain.NewHandler(shoppingRepo, dealExplainAgent, dealExplainCache)
	r.Get("/api/shopping-items/{id}/explain", dealExplainHandler.Explain)

	// Price Benchmark vs Market Average (Phase 70)
	benchmarkCache := benchmark.NewCache(2 * time.Hour)
	benchmarkAgent := benchmark.NewAgent(provider)
	benchmarkHandler := benchmark.NewHandler(shoppingRepo, benchmarkAgent, benchmarkCache)
	r.Get("/api/shopping-items/{id}/benchmark", benchmarkHandler.Get)

	// Mission Collaboration Threads (Phase 61)
	commentRepo := comment.NewRepository(pool)
	commentResolver := comment.NewBridgeResolver(func(ctx context.Context, slug string) (string, bool, error) {
		m, err := missionSvc.GetBySlug(ctx, slug)
		if err != nil {
			return "", false, err
		}
		if m == nil {
			return "", false, nil
		}
		return m.ID.String(), true, nil
	})
	commentHandler := comment.NewHandler(commentRepo, commentResolver)
	r.Mount("/api/missions/{slug}/comments", commentHandler.Routes())

	// Stock availability heuristic (Phase 42)
	stockCache := stockcheck.NewCache(5 * time.Minute)
	stockHandler := stockcheck.NewHandler(stockCache)
	r.Get("/api/stock-check", stockHandler.Check)

	// AI Product Substitute Finder — French market (Phase 54)
	substituteCache := substitute.NewCache(2 * time.Hour)
	substituteAgent := substitute.NewAgent(provider)
	substituteGetter := substitute.NewOptionRepoGetter(optionRepo)
	substituteHandler := substitute.NewHandler(substituteGetter, substituteCache, substituteAgent)
	r.Get("/api/options/{id}/substitutes", substituteHandler.GetSubstitutes)

	// AI Shopping Summary Report (Phase 43)
	summaryCache := summary.NewCache(24 * time.Hour)
	summaryAgent := summary.NewAgent(provider)
	summaryHandler := summary.NewHandler(summaryAgent, missionRepo, optionRepo, shoppingRepo, summaryCache)
	r.Post("/api/missions/{missionID}/summary", summaryHandler.Generate)
	r.Get("/api/missions/{missionID}/summary", summaryHandler.Get)

	// Mission Health Score (Phase 57)
	healthCache := health.NewCache(30 * time.Minute)
	healthAgent := health.NewAgent(provider)
	healthHandler := health.NewHandler(missionRepo, optionRepo, shoppingRepo, healthAgent, healthCache)
	r.Get("/api/missions/{slug}/health", healthHandler.GetHealth)

	// Price Drop Weekly Digest (Phase 58)
	digestRepo := digest.NewRepository(pool)
	digestHandler := digest.NewHandler(digestRepo)
	r.Get("/api/digest/weekly", digestHandler.Weekly)

	// LLM pool health (nil when provider is Anthropic-only)
	if smartRouter != nil {
		checker := llm.NewHealthChecker(smartRouter.Pool(), smartRouter.Breakers())
		r.Get("/api/health/llm", llm.HealthHandler(checker))
	}

	// Prometheus metrics endpoint (Phase 14)
	if metricsHandler != nil {
		r.Get("/metrics", metricsHandler.ServeHTTP)
	}

	addr := ":" + cfg.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	// Drain scheduler and HTTP server concurrently so in-flight LLM calls (up to
	// 60 s) are not cut short by a sequential scheduler drain (also up to 60 s).
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	if sched != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sched.Stop()
		}()
	}
	// Wait for embedding worker to drain in-flight jobs.
	// The 65 s deadline is shared across HTTP drain, scheduler drain, and embed
	// drain. HTTP in-flight LLM calls (≤60 s) and embed calls run concurrently
	// so they don't compound. pool.Close is called only after wg.Wait() ensures
	// all worker goroutines have exited and released pool connections.
	wg.Add(1)
	go func() {
		defer wg.Done()
		embedWorker.Wait()
	}()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}
	wg.Wait()
	pool.Close()
	log.Info("shutdown complete")
}

// buildSmartRouter constructs the LLM provider based on config.
//
// Modes:
//   - "anthropic"  → single AnthropicProvider (no SmartRouter)
//   - "ollama"     → SmartRouter with heavy + fast Ollama models
//   - "routing"    → SmartRouter: heavy Ollama → fast Ollama → Anthropic fallback
//   - default      → same as "routing" when ANTHROPIC_API_KEY is set, else "ollama"
func buildSmartRouter(cfg *config.Config, log *slog.Logger, rec llm.LLMRecorder) (llm.Provider, *llm.SmartRouter) {
	mode := strings.ToLower(cfg.LLMProvider)

	if mode == "anthropic" {
		log.Info("llm provider: anthropic", "model", "claude-sonnet-4-6")
		return llm.NewAnthropicProvider(cfg.AnthropicAPIKey), nil
	}

	// Build Ollama entries.
	heavyOpts := []llm.OllamaOption{llm.WithTimeout(cfg.OllamaHeavyTimeout)}
	fastOpts := []llm.OllamaOption{llm.WithTimeout(cfg.OllamaFastTimeout)}

	heavyProvider := llm.NewOllamaProvider(cfg.OllamaBaseURL, cfg.OllamaHeavyModel, heavyOpts...)
	fastProvider := llm.NewOllamaProvider(cfg.OllamaBaseURL, cfg.OllamaFastModel, fastOpts...)

	entries := []llm.ModelEntry{
		{
			Name:               cfg.OllamaHeavyModel,
			Provider:           heavyProvider,
			Capabilities:       llm.CapText | llm.CapToolUse | llm.CapLongCtx,
			Priority:           1,
			Timeout:            cfg.OllamaHeavyTimeout,
			SystemPromptPrefix: "/no_think",
		},
		{
			Name:               cfg.OllamaFastModel,
			Provider:           fastProvider,
			Capabilities:       llm.CapText | llm.CapToolUse,
			Priority:           2,
			Timeout:            cfg.OllamaFastTimeout,
			SystemPromptPrefix: "/no_think",
		},
	}

	// Optional cloud Ollama model.
	routerOpts := []llm.SmartRouterOption{llm.WithRecorder(rec)}
	if cfg.OllamaCloudURL != "" && cfg.OllamaCloudModel != "" {
		cloudOpts := []llm.OllamaOption{
			llm.WithAPIKey(cfg.OllamaCloudAPIKey),
			llm.WithTimeout(cfg.OllamaHeavyTimeout),
		}
		cloudProvider := llm.NewOllamaProvider(cfg.OllamaCloudURL, cfg.OllamaCloudModel, cloudOpts...)
		entries = append(entries, llm.ModelEntry{
			Name:         cfg.OllamaCloudModel,
			Provider:     cloudProvider,
			Capabilities: llm.CapText | llm.CapToolUse | llm.CapLongCtx,
			Priority:     3,
			Timeout:      cfg.OllamaHeavyTimeout,
		})
		routerOpts = append(routerOpts, llm.WithModelRateLimit(cfg.OllamaCloudModel, cfg.OllamaCloudRPM))
		log.Info("llm cloud model enabled", "url", cfg.OllamaCloudURL, "model", cfg.OllamaCloudModel, "rpm", cfg.OllamaCloudRPM)
	}

	// Optional Anthropic fallback (routing mode).
	if mode != "ollama" && cfg.AnthropicAPIKey != "" {
		anthropic := llm.NewAnthropicProvider(cfg.AnthropicAPIKey)
		entries = append(entries, llm.ModelEntry{
			Name:         "claude-sonnet-4-6",
			Provider:     anthropic,
			Capabilities: llm.CapText | llm.CapToolUse | llm.CapLongCtx,
			Priority:     10,
			Timeout:      60 * time.Second,
		})
		log.Info("llm anthropic fallback enabled")
	} else if mode != "ollama" && cfg.AnthropicAPIKey == "" {
		log.Warn("ANTHROPIC_API_KEY not set — Anthropic fallback disabled, using Ollama only")
	}

	pool := llm.NewModelPool(entries...)
	sr := llm.NewSmartRouter(pool, log, routerOpts...)

	log.Info("llm provider: smart-router",
		"heavy", cfg.OllamaHeavyModel,
		"fast", cfg.OllamaFastModel,
		"models", len(entries))
	return sr, sr
}

// corsMiddleware adds permissive CORS headers for local development only.
// It is only registered when ENV=development.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
