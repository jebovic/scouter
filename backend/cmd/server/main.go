package main

import (
	"context"
	"fmt"
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

	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/config"
	"github.com/jibei/scouter/internal/db"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/export"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/research"
	"github.com/jibei/scouter/internal/scheduler"
	"github.com/jibei/scouter/internal/search"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/template"
	"github.com/jibei/scouter/internal/usage"
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

	// Build LLM provider
	provider, smartRouter := buildSmartRouter(cfg, log)

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
	pricingAgent := pricing.NewAgent(provider, shoppingRepo, agentRunRepo, usageSvc)
	decisionAgent := decision.NewAgent(provider, usageSvc)

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

	// Price check scheduler (opt-in via PRICE_CHECK_ENABLED=true)
	var sched *scheduler.Scheduler
	if cfg.PriceCheckEnabled {
		orch := scheduler.NewOrchestrator(
			missionRepo, optionRepo, shoppingRepo, notifRepo,
			pricingAgent, cfg.PriceCheckMaxMissions, log,
		)
		var schedErr error
		sched, schedErr = scheduler.New(cfg.PriceCheckCron, orch, log)
		if schedErr != nil {
			log.Error("scheduler init failed", "err", schedErr)
			os.Exit(1)
		}
		sched.Start()
	}

	// Search (Phase 11)
	searchRepo := search.NewRepository(pool)
	searchHandler := search.NewHandler(searchRepo, embedder, embedRepo, embedWorker)

	// Export
	exportGatherer := export.NewGatherer(missionRepo, optionRepo, shoppingRepo, decisionRepo)
	exportHandler := export.NewHandler(exportGatherer)

	// Templates (no DB dependency — compiled into binary)
	templateReg := template.NewRegistry()
	templateHandler := template.NewHandler(templateReg)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(1 << 20)) // 1 MiB max request body

	if cfg.Env == "development" {
		r.Use(corsMiddleware)
	}

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	r.Get("/api/usage", usageHandler.GetSummary)

	// Mission CRUD: /api/missions and /api/missions/{slug}
	r.Get("/api/missions", missionHandler.List)
	r.Post("/api/missions", missionHandler.Create)
	r.Get("/api/missions/{slug}", missionHandler.Get)
	r.Patch("/api/missions/{slug}", missionHandler.Update)
	r.Delete("/api/missions/{slug}", missionHandler.Delete)

	// Mission sub-resources
	r.Mount("/api/missions/{missionID}/options", optionHandler.Routes())
	r.Mount("/api/missions/{missionID}/shopping", shoppingHandler.Routes())
	r.Mount("/api/missions/{missionID}/research", researchHandler.Routes())
	r.Mount("/api/missions/{missionID}/pricing", pricingHandler.Routes())
	r.Mount("/api/missions/{missionID}/decision", decisionHandler.Routes())
	r.Mount("/api/missions/{missionID}/agent-runs", agentRunHandler.Routes())

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
	r.Post("/api/search/reindex", searchHandler.Reindex)

	// Notifications
	r.Mount("/api/notifications", notifHandler.Routes())

	// Templates
	r.Get("/api/templates", templateHandler.List)
	r.Get("/api/templates/{slug}", templateHandler.Get)

	// LLM pool health (nil when provider is Anthropic-only)
	if smartRouter != nil {
		checker := llm.NewHealthChecker(smartRouter.Pool(), smartRouter.Breakers())
		r.Get("/api/health/llm", llm.HealthHandler(checker))
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
func buildSmartRouter(cfg *config.Config, log *slog.Logger) (llm.Provider, *llm.SmartRouter) {
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
	var routerOpts []llm.SmartRouterOption
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
