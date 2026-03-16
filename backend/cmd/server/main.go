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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"

	"github.com/jibei/scouter/internal/admin"
	"github.com/jibei/scouter/internal/imagefetch"
	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/coach"
	"github.com/jibei/scouter/internal/collaborator"
	"github.com/jibei/scouter/internal/config"
	"github.com/jibei/scouter/internal/db"
	"github.com/jibei/scouter/internal/dealcalendar"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/export"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/metrics"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/pricealertdigest"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/purchase"
	"github.com/jibei/scouter/internal/research"
	"github.com/jibei/scouter/internal/researchjob"
	"github.com/jibei/scouter/internal/scheduler"
	"github.com/jibei/scouter/internal/search"
	"github.com/jibei/scouter/internal/settings"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/template"
	"github.com/jibei/scouter/internal/translation"
	"github.com/jibei/scouter/internal/usage"
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

	// Image fetch worker
	minioUploader, err := imagefetch.NewUploader(ctx, imagefetch.UploaderConfig{
		Endpoint:  cfg.MinioEndpoint,
		PublicURL: cfg.MinioPublicURL,
		AccessKey: cfg.MinioAccessKey,
		SecretKey: cfg.MinioSecretKey,
		Bucket:    cfg.MinioBucket,
	})
	if err != nil {
		log.Error("minio init", "err", err)
		os.Exit(1)
	}
	imageRepo := imagefetch.NewRepository(pool)
	imageScraper := imagefetch.NewScraper()
	imageWorker := imagefetch.NewWorker(imageRepo, minioUploader, imageScraper)
	imageWorker.Start(ctx)

	imageHandler := imagefetch.NewHandler(imageRepo, minioUploader)

	// Purge cron: enforce MinIO quota daily (always runs, independent of price-check scheduler)
	purgeCron := cron.New(cron.WithLocation(time.UTC))
	purgeCron.AddFunc("@daily", imagefetch.PurgeJob(ctx, minioUploader, imageRepo, cfg.MinioQuotaMB)) //nolint:errcheck
	purgeCron.Start()

	// Agents
	researchAgent := research.NewAgent(provider, optionRepo, agentRunRepo, usageSvc)
	researchAgent.SetEmbedChannel(embedWorker.Jobs())
	researchAgent.SetImageChannel(imageWorker.Jobs())
	researchAgent.SetRecorder(rec)

	// Translation worker — only when a SmartRouter is available (nil when LLM_PROVIDER=anthropic).
	var translateWorker *translation.Worker
	var translateHandler *translation.Handler
	if smartRouter != nil {
		rawLocales := strings.Split(cfg.SupportedLocales, ",")
		var supportedLocales []string
		for _, l := range rawLocales {
			if t := strings.TrimSpace(l); t != "" {
				supportedLocales = append(supportedLocales, t)
			}
		}
		if len(supportedLocales) > 0 {
			translator := translation.NewTranslator(smartRouter)
			translateWorker = translation.NewWorker(translator, optionRepo, supportedLocales, nil)
			translateWorker.Start(ctx)
			translateHandler = translation.NewHandler(translateWorker.Submit)
			researchAgent.SetTranslateChannel(translateWorker.Jobs())
		}
	}

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
	optionHandler.WithImageHandler(imageHandler)
	shoppingHandler := shopping.NewHandler(shoppingSvc)
	notifHandler := notification.NewHandler(notifRepo)
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

	// Price Alert Digest (Phase 166)
	priceAlertDigestHandler := pricealertdigest.NewHandler(pool)

	// Async research jobs
	researchJobRepo := researchjob.NewRepository(pool)
	if err := researchJobRepo.FailStaleJobs(context.Background()); err != nil {
		log.Warn("failed to clean stale research jobs", "err", err)
	}
	researchJobHandler := researchjob.NewHandler(
		researchJobRepo,
		missionSvc,
		&researchAgentAdapter{agent: researchAgent},
	)

	deps := routeDeps{
		pool:        pool,
		provider:    provider,
		smartRouter: smartRouter,
		rec:         rec,
		cfg:         cfg,
		log:         log,

		missionRepo:  missionRepo,
		optionRepo:   optionRepo,
		shoppingRepo: shoppingRepo,
		notifRepo:    notifRepo,
		embedder:         embedder,
		embedRepo:        embedRepo,
		embedWorker:      embedWorker,
		translateWorker:  translateWorker,
		translateHandler: translateHandler,

		missionSvc:  missionSvc,
		optionSvc:   optionSvc,
		shoppingSvc: shoppingSvc,

		missionHandler:          missionHandler,
		optionHandler:           optionHandler,
		shoppingHandler:         shoppingHandler,
		notifHandler:            notifHandler,
		pricingHandler:          pricingHandler,
		researchJobRepo:         researchJobRepo,
		researchJobHandler:      researchJobHandler,
		usageHandler:            usageHandler,
		decisionHandler:         decisionHandler,
		agentRunHandler:         agentRunHandler,
		coachHandler:            coachHandler,
		purchaseHandler:         purchaseHandler,
		statsHandler:            statsHandler,
		searchHandler:           searchHandler,
		exportHandler:           exportHandler,
		settingsHandler:         settingsHandler,
		collaboratorHandler:     collaboratorHandler,
		voteHandler:             voteHandler,
		adminHandler:            adminHandler,
		templateHandler:         templateHandler,
		dealCalHandler:          dealCalHandler,
		priceAlertDigestHandler: priceAlertDigestHandler,
		wishlistRepo:            wishlistRepo,
		metricsHandler:          metricsHandler,
	}

	// Router
	r := chi.NewRouter()
	registerRoutes(r, &deps)

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
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-purgeCron.Stop().Done()
	}()
	// Wait for embedding worker to drain in-flight jobs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		embedWorker.Wait()
	}()
	// Wait for image fetch worker to drain in-flight jobs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		imageWorker.Wait()
	}()
	// Wait for translation worker to drain in-flight jobs (nil when SmartRouter unavailable).
	if translateWorker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			translateWorker.Wait()
		}()
	}
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
