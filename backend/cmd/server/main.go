package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/config"
	"github.com/jibei/scouter/internal/db"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/research"
	"github.com/jibei/scouter/internal/scheduler"
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
	provider := buildProvider(cfg, log)

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

	// Agents
	researchAgent := research.NewAgent(provider, optionRepo, agentRunRepo, usageSvc)
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

	// Notifications
	r.Mount("/api/notifications", notifHandler.Routes())

	// Templates
	r.Get("/api/templates", templateHandler.List)
	r.Get("/api/templates/{slug}", templateHandler.Get)

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

	// Drain scheduler first so any running price check can finish.
	if sched != nil {
		sched.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}
	pool.Close()
	log.Info("shutdown complete")
}

// buildProvider constructs the LLM provider based on config.
// "routing" (default when both keys are available) uses Ollama as primary
// with Anthropic fallback. "ollama" uses Ollama only. "anthropic" uses Anthropic only.
func buildProvider(cfg *config.Config, log *slog.Logger) llm.Provider {
	ollamaProvider := llm.NewOllamaProvider(cfg.OllamaBaseURL, cfg.OllamaModel)
	mode := strings.ToLower(cfg.LLMProvider)

	switch mode {
	case "anthropic":
		log.Info("llm provider: anthropic", "model", "claude-sonnet-4-6")
		return llm.NewAnthropicProvider(cfg.AnthropicAPIKey)
	case "ollama":
		log.Info("llm provider: ollama", "base_url", cfg.OllamaBaseURL, "model", cfg.OllamaModel)
		return ollamaProvider
	default: // "routing" or anything else → routing mode
		if cfg.AnthropicAPIKey == "" {
			log.Warn("ANTHROPIC_API_KEY not set — RoutingProvider fallback disabled, using Ollama only")
			return ollamaProvider
		}
		anthropicProvider := llm.NewAnthropicProvider(cfg.AnthropicAPIKey)
		log.Info("llm provider: routing (ollama primary → anthropic fallback)",
			"base_url", cfg.OllamaBaseURL, "model", cfg.OllamaModel)
		return llm.NewRoutingProvider(ollamaProvider, anthropicProvider)
	}
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
