package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jibei/scouter/internal/activityfeed"
	"github.com/jibei/scouter/internal/admin"
	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/autotag"
	"github.com/jibei/scouter/internal/benchmark"
	"github.com/jibei/scouter/internal/budgetadvisor"
	"github.com/jibei/scouter/internal/budgetalert"
	"github.com/jibei/scouter/internal/budgetheatmap"
	"github.com/jibei/scouter/internal/budgetplanner"
	"github.com/jibei/scouter/internal/budgetrec"
	"github.com/jibei/scouter/internal/bundledetector"
	"github.com/jibei/scouter/internal/burnrate"
	"github.com/jibei/scouter/internal/carbon"
	"github.com/jibei/scouter/internal/cashback"
	"github.com/jibei/scouter/internal/cashbacktracker"
	"github.com/jibei/scouter/internal/coach"
	"github.com/jibei/scouter/internal/collaborator"
	"github.com/jibei/scouter/internal/comment"
	"github.com/jibei/scouter/internal/comparison"
	"github.com/jibei/scouter/internal/comparisonscore"
	"github.com/jibei/scouter/internal/competitorprice"
	"github.com/jibei/scouter/internal/conditionpricing"
	"github.com/jibei/scouter/internal/config"
	"github.com/jibei/scouter/internal/couponfinder"
	"github.com/jibei/scouter/internal/crossmission"
	"github.com/jibei/scouter/internal/csvexport"
	"github.com/jibei/scouter/internal/currency"
	"github.com/jibei/scouter/internal/dealaggregator"
	"github.com/jibei/scouter/internal/dealcalendar"
	"github.com/jibei/scouter/internal/dealexplain"
	"github.com/jibei/scouter/internal/dealfeed"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/decisionmatrix"
	"github.com/jibei/scouter/internal/digest"
	"github.com/jibei/scouter/internal/dropreporter"
	"github.com/jibei/scouter/internal/duplicatedetector"
	"github.com/jibei/scouter/internal/ecoscore"
	"github.com/jibei/scouter/internal/elasticity"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/expensecategorizer"
	"github.com/jibei/scouter/internal/export"
	"github.com/jibei/scouter/internal/forecast"
	"github.com/jibei/scouter/internal/frenchbenchmark"
	"github.com/jibei/scouter/internal/frenchmarket"
	"github.com/jibei/scouter/internal/giftfinder"
	"github.com/jibei/scouter/internal/health"
	"github.com/jibei/scouter/internal/holidayalert"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/inflationtracker"
	"github.com/jibei/scouter/internal/itemtagger"
	"github.com/jibei/scouter/internal/listoptimizer"
	"github.com/jibei/scouter/internal/listsorter"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/loyaltypoints"
	"github.com/jibei/scouter/internal/loyaltytracker"
	"github.com/jibei/scouter/internal/marketplace"
	"github.com/jibei/scouter/internal/merchantrecommender"
	"github.com/jibei/scouter/internal/metrics"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/missionprogress"
	"github.com/jibei/scouter/internal/missionreport"
	"github.com/jibei/scouter/internal/negotiate"
	"github.com/jibei/scouter/internal/negotiation"
	"github.com/jibei/scouter/internal/negotiationoutcome"
	"github.com/jibei/scouter/internal/negotiationscript"
	"github.com/jibei/scouter/internal/negotiationsim"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/notificationrules"
	"github.com/jibei/scouter/internal/optimizer"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/persona"
	"github.com/jibei/scouter/internal/prediction"
	"github.com/jibei/scouter/internal/pricealertdigest"
	"github.com/jibei/scouter/internal/pricealertrule"
	"github.com/jibei/scouter/internal/priceanalytics"
	"github.com/jibei/scouter/internal/priceannotation"
	"github.com/jibei/scouter/internal/pricecomp"
	"github.com/jibei/scouter/internal/pricecomparison"
	"github.com/jibei/scouter/internal/pricedigest"
	"github.com/jibei/scouter/internal/pricedropwatch"
	"github.com/jibei/scouter/internal/pricefloor"
	"github.com/jibei/scouter/internal/priceforecast"
	"github.com/jibei/scouter/internal/priceinsights"
	"github.com/jibei/scouter/internal/pricestreak"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/product"
	"github.com/jibei/scouter/internal/purchase"
	"github.com/jibei/scouter/internal/purchaseadvisor"
	"github.com/jibei/scouter/internal/quantityoptimizer"
	"github.com/jibei/scouter/internal/rebalancer"
	"github.com/jibei/scouter/internal/receipt"
	"github.com/jibei/scouter/internal/regretanalyzer"
	"github.com/jibei/scouter/internal/reordersuggestion"
	"github.com/jibei/scouter/internal/research"
	"github.com/jibei/scouter/internal/researchjob"
	"github.com/jibei/scouter/internal/reviews"
	"github.com/jibei/scouter/internal/reviewsummary"
	"github.com/jibei/scouter/internal/roicalculator"
	"github.com/jibei/scouter/internal/salecalendar"
	"github.com/jibei/scouter/internal/scorecard"
	"github.com/jibei/scouter/internal/search"
	"github.com/jibei/scouter/internal/seasonal"
	"github.com/jibei/scouter/internal/seasonalcalendar"
	"github.com/jibei/scouter/internal/settings"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/shoppingoptimizer"
	"github.com/jibei/scouter/internal/spendinganalytics"
	"github.com/jibei/scouter/internal/spendingvelocity"
	"github.com/jibei/scouter/internal/stockalert"
	"github.com/jibei/scouter/internal/stockcheck"
	"github.com/jibei/scouter/internal/substitute"
	"github.com/jibei/scouter/internal/summary"
	"github.com/jibei/scouter/internal/targetsuggestion"
	"github.com/jibei/scouter/internal/template"
	"github.com/jibei/scouter/internal/timelineplanner"
	"github.com/jibei/scouter/internal/timing"
	"github.com/jibei/scouter/internal/timingrecommender"
	"github.com/jibei/scouter/internal/translation"
	"github.com/jibei/scouter/internal/travel"
	"github.com/jibei/scouter/internal/usage"
	"github.com/jibei/scouter/internal/volatilitycalendar"
	"github.com/jibei/scouter/internal/vote"
	"github.com/jibei/scouter/internal/watchlist"
	"github.com/jibei/scouter/internal/weeklydigest"
	"github.com/jibei/scouter/internal/wishlist"
	"github.com/jibei/scouter/internal/wishlistprioritizer"
	"github.com/jibei/scouter/internal/wishlistshare"
	"github.com/jibei/scouter/internal/wishlistvote"

	"github.com/jackc/pgx/v5/pgxpool"
)

// routeDeps holds all pre-initialized dependencies for route registration.
type routeDeps struct {
	pool        *pgxpool.Pool
	provider    llm.Provider
	smartRouter *llm.SmartRouter
	rec         metrics.Recorder
	cfg         *config.Config
	log         *slog.Logger

	// repos
	missionRepo      mission.Repository
	optionRepo       option.Repository
	shoppingRepo     shopping.Repository
	notifRepo        notification.Repository
	embedder         *llm.OllamaEmbedder
	embedRepo        embedding.Repository
	embedWorker      *embedding.Worker
	translateWorker  *translation.Worker
	translateHandler *translation.Handler

	// services
	missionSvc  *mission.Service
	optionSvc   *option.Service
	shoppingSvc *shopping.Service

	// researchjob
	researchJobRepo    researchjob.Repository
	researchJobHandler *researchjob.Handler

	// core handlers
	missionHandler          *mission.Handler
	optionHandler           *option.Handler
	shoppingHandler         *shopping.Handler
	notifHandler            *notification.Handler
	pricingHandler          *pricing.Handler
	usageHandler            *usage.Handler
	decisionHandler         *decision.Handler
	agentRunHandler         *agentrun.Handler
	coachHandler            *coach.Handler
	purchaseHandler         *purchase.Handler
	statsHandler            *purchase.StatsHandler
	searchHandler           *search.Handler
	exportHandler           *export.Handler
	settingsHandler         *settings.Handler
	collaboratorHandler     *collaborator.Handler
	voteHandler             *vote.Handler
	adminHandler            *admin.Handler
	templateHandler         *template.Handler
	dealCalHandler          *dealcalendar.Handler
	priceAlertDigestHandler *pricealertdigest.Handler
	wishlistRepo            *wishlist.Repository
	metricsHandler          http.Handler
}

// researchAgentAdapter bridges research.Agent to researchjob.AgentRunner.
type researchAgentAdapter struct {
	agent *research.Agent
}

func (a *researchAgentAdapter) Run(ctx context.Context, ms mission.Mission, fb *researchjob.FeedbackInput) ([]option.Option, error) {
	var rFb *research.FeedbackInput
	if fb != nil {
		rFb = &research.FeedbackInput{Feedback: fb.Feedback}
	}
	return a.agent.Run(ctx, ms, rFb)
}

// registerRoutes mounts all HTTP routes onto r using dependencies from d.
func registerRoutes(r chi.Router, d *routeDeps) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(1 << 20)) // 1 MiB max request body
	r.Use(metrics.Middleware(d.rec))

	if d.cfg.Env == "development" {
		r.Use(corsMiddleware)
	}

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := d.pool.Ping(pingCtx); err != nil {
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

	r.Get("/api/usage", d.usageHandler.GetSummary)

	// Mission CRUD: /api/missions and /api/missions/{slug}
	r.Get("/api/missions", d.missionHandler.List)
	r.Post("/api/missions", d.missionHandler.Create)
	r.Get("/api/missions/{slug}", d.missionHandler.Get)
	r.Patch("/api/missions/{slug}", d.missionHandler.Update)
	r.Delete("/api/missions/{slug}", d.missionHandler.Delete)
	r.Post("/api/missions/{slug}/duplicate", d.missionHandler.Duplicate)
	r.Post("/api/missions/{slug}/clone", d.missionHandler.Clone)

	// Wire translation handler into option handler (nil-safe)
	if d.translateHandler != nil {
		d.optionHandler.WithTranslationHandler(d.translateHandler.Retranslate)
	}

	// Mission sub-resources
	r.Mount("/api/missions/{missionID}/options", d.optionHandler.Routes())
	r.Mount("/api/missions/{missionID}/shopping", d.shoppingHandler.Routes())
	r.Mount("/api/missions/{missionID}/research", d.researchJobHandler.Routes())
	r.Mount("/api/missions/{missionID}/pricing", d.pricingHandler.Routes())
	r.Mount("/api/missions/{missionID}/decision", d.decisionHandler.Routes())
	r.Mount("/api/missions/{missionID}/agent-runs", d.agentRunHandler.Routes())
	r.Mount("/api/missions/{slug}/coach", d.coachHandler.Routes())

	// Option export by slug (Phase 75)
	r.Get("/api/missions/{slug}/options/export.csv", func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		m, err := d.missionSvc.GetBySlug(r.Context(), slug)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if m == nil {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		// Convert to missionID context for the option handler
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("missionID", m.ID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		d.optionHandler.ExportCSV(w, r)
	})

	// Purchase lifecycle (Phase 12)
	r.Mount("/api/missions/{missionID}/purchase", d.purchaseHandler.Routes())
	r.Get("/api/stats", d.statsHandler.Stats)
	r.Get("/api/stats/monthly", d.statsHandler.MonthlyStats)

	// Export, share, archive
	csvExportHandler := csvexport.NewHandler(d.pool)
	r.Get("/api/missions/{missionID}/export", d.exportHandler.Export)
	r.Get("/api/missions/{missionID}/export/csv", csvExportHandler.ExportMissionCSV)
	r.Post("/api/missions/{missionID}/share", d.missionHandler.Share)
	r.Delete("/api/missions/{missionID}/share", d.missionHandler.RevokeShare)
	r.Post("/api/missions/{missionID}/archive", d.missionHandler.Archive)
	r.Post("/api/missions/{missionID}/unarchive", d.missionHandler.Unarchive)

	// Regret Analyzer (Phase 158)
	regretHandler := regretanalyzer.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/regret-analysis", regretHandler.GetAnalysis)

	// Budget Allocation Advisor (Phase 163)
	budgetAdvisorHandler := budgetadvisor.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/budget-advice", budgetAdvisorHandler.GetAdvice)

	// Price Alert Digest (Phase 166)
	r.Get("/api/missions/{missionId}/price-alert-digest", d.priceAlertDigestHandler.GetDigest)

	// Smart Expense Categorizer (Phase 164)
	expenseCatHandler := expensecategorizer.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/expense-summary", expenseCatHandler.GetSummary)

	// Public shared mission (accessible without auth, always CORS-open)
	r.With(corsMiddleware).Get("/api/shared/{token}", func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		m, err := d.missionSvc.GetByShareToken(r.Context(), token)
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
	r.Get("/api/search", d.searchHandler.Search)
	r.Get("/api/options/{optionID}/similar", d.searchHandler.Similar)
	if d.cfg.Env == "development" {
		r.Post("/api/search/reindex", d.searchHandler.Reindex)
	}

	// Notifications
	r.Mount("/api/notifications", d.notifHandler.Routes())

	// Templates
	r.Get("/api/templates", d.templateHandler.List)
	r.Get("/api/templates/{slug}", d.templateHandler.Get)

	// Deal Calendar
	r.Get("/api/deal-calendar", d.dealCalHandler.List)

	// Smart Cashback Tracker (Phase 87)
	cashbackRepo := cashback.NewRepository()
	cashbackHandler := cashback.NewHandler(cashbackRepo)
	r.Mount("/api/cashback", cashbackHandler.Routes())

	// French Promo Feed (Phase 83)
	dealFeedHandler := dealfeed.NewHandler(d.smartRouter, d.log)
	r.Get("/api/promo-feed", dealFeedHandler.Get)

	// Settings (Phase 13)
	r.Get("/api/settings", d.settingsHandler.GetAll)
	r.Patch("/api/settings", d.settingsHandler.Update)

	// Data management (Phase 13)
	r.Delete("/api/data", d.adminHandler.DeleteAllData)

	// Collaborative missions (Phase 16)
	r.Mount("/api/missions/{missionID}/invites", d.collaboratorHandler.InviteRoutes())
	r.Mount("/api/missions/{missionID}/collaborators", d.collaboratorHandler.CollaboratorRoutes())
	r.Get("/api/invites/{token}", d.collaboratorHandler.GetInvite)
	r.Post("/api/invites/{token}/join", d.collaboratorHandler.JoinMission)
	r.Mount("/api/options/{optionID}/votes", d.voteHandler.Routes())

	// Smart Budget Forecaster (Phase 23)
	forecastRepo := forecast.NewRepository(d.pool)
	forecastAgent := forecast.NewForecastAgent(d.provider, forecastRepo, d.pool)
	forecastHandler := forecast.NewHandler(forecastAgent, forecastRepo)
	r.Route("/api/missions/{missionID}/forecast", forecastHandler.Routes())

	// Wish List route (repo declared in main, before the scheduler block)
	r.Route("/api/wishlist", wishlist.Routes(d.wishlistRepo))

	// Envelope Budgeting (Phase 35)
	envelopeRepo := envelope.NewRepository(d.pool)
	envelopeSvc := envelope.NewService(envelopeRepo)
	envelopeHandler := envelope.NewHandler(envelopeSvc)
	r.Route("/api/envelopes", envelope.Routes(envelopeHandler))

	// Spending Persona (Phase 26) + AI Shopping Persona Insights (Phase 82)
	personaRepo := persona.NewRepository(d.pool)
	personaAgent := persona.NewPersonaAgent(d.provider, personaRepo, d.pool)
	personaHandler := persona.NewHandler(personaAgent, personaRepo)
	shoppingPersonaAgent := persona.NewShoppingPersonaAgent(d.provider)
	shoppingPersonaHandler := persona.NewShoppingPersonaHandler(d.missionRepo, d.optionRepo, d.shoppingRepo, shoppingPersonaAgent)
	r.Route("/api/persona", func(r chi.Router) {
		personaHandler.Routes()(r)
		r.Get("/shopping", shoppingPersonaHandler.GetShoppingPersona)
	})

	// Product barcode lookup (Phase 25)
	productLooker := product.NewLooker()
	r.Route("/api/products", product.Routes(productLooker))

	// Social Proof & Review Aggregation (Phase 30)
	reviewsCache := reviews.NewCache(6 * time.Hour)
	reviewsAgent := reviews.NewAgent(d.provider)
	reviewsGetter := reviews.NewOptionRepoGetter(d.optionRepo)
	reviewsHandler := reviews.NewHandler(reviewsGetter, reviewsCache, reviewsAgent)

	// Travel integrations (Phase 29)
	avClient := travel.NewAviationStackClient(d.cfg.AviationStackAPIKey, d.cfg.AviationStackBaseURL)
	sncfClient := travel.NewSNCFClient(d.cfg.SNCFAPIKey, d.cfg.SNCFBaseURL)
	travelCache := travel.NewCache(time.Hour)
	travelHandler := travel.NewHandler(avClient, sncfClient, travelCache)
	r.Mount("/api/travel", travelHandler.Routes())

	// Weighted Comparison Matrix (Phase 22)
	comparisonRepo := comparison.NewRepository(d.pool)
	r.Route("/api/missions/{missionID}/comparison-weights", comparison.WeightRoutes(comparisonRepo, d.pool))
	r.Get("/api/missions/{missionID}/matrix", comparison.NewHandler(comparisonRepo, d.pool).Matrix)

	// Social Proof & Review Aggregation (Phase 30)
	r.Get("/api/options/{id}/reviews", reviewsHandler.GetReviews)
	r.Post("/api/options/{id}/reviews/refresh", reviewsHandler.RefreshReviews)

	// Product Price Comparison (Phase 31)
	pricecompCache := pricecomp.NewCache(30 * time.Minute)
	pricecompAgent := pricecomp.NewAgent(d.provider)
	pricecompGetter := pricecomp.NewOptionRepoGetter(d.optionRepo)
	pricecompHandler := pricecomp.NewHandler(pricecompGetter, pricecompCache, pricecompAgent)
	r.Get("/api/options/{id}/price-comparison", pricecompHandler.GetComparison)
	r.Post("/api/options/{id}/price-comparison/refresh", pricecompHandler.RefreshComparison)

	// AI Purchase Timing Advisor (Phase 32)
	timingCache := timing.NewCache(24 * time.Hour)
	timingAgent := timing.NewAgent(d.provider)
	timingGetter := timing.NewMissionRepoGetter(d.pool)
	timingHandler := timing.NewHandler(timingGetter, timingCache, timingAgent)
	r.Post("/api/missions/{id}/timing-advice", timingHandler.GetTimingAdvice)

	// AI Negotiation Coach (Phase 17)
	negotiationHandler := negotiation.NewHandler(d.pool, d.provider)
	r.Mount("/api/options/{optionID}", negotiationHandler.Routes())

	// Smart Price Prediction (Phase 52)
	predictionHandler := prediction.NewHandler(d.shoppingRepo)
	r.Get("/api/shopping-items/{id}/prediction", predictionHandler.GetPrediction)

	// Smart Category Auto-Tagging (Phase 65)
	autotagAgent := autotag.NewAgent(d.provider)
	autotagHandler := autotag.NewHandler(d.missionRepo, autotagAgent)
	r.Post("/api/missions/{slug}/suggest-category", autotagHandler.Suggest)

	// AI Shopping List Optimizer (Phase 68)
	optimizerHandler := optimizer.NewHandler(d.pool, d.provider)
	r.Post("/api/missions/{slug}/optimize", optimizerHandler.Optimize)

	// Smart Receipt Analyzer (Phase 69)
	receiptRepo := receipt.NewRepository(d.pool)
	receiptAgent := receipt.NewAgent(d.provider)
	receiptHandler := receipt.NewHandler(receiptRepo, receiptAgent)
	r.Post("/api/missions/{slug}/receipts", receiptHandler.Analyze)
	r.Get("/api/missions/{slug}/receipts", receiptHandler.List)

	// AI Deal Explainer (Phase 60)
	dealExplainCache := dealexplain.NewCache(1 * time.Hour)
	dealExplainAgent := dealexplain.NewAgent(d.provider)
	dealExplainHandler := dealexplain.NewHandler(d.shoppingRepo, dealExplainAgent, dealExplainCache)
	r.Get("/api/shopping-items/{id}/explain", dealExplainHandler.Explain)

	// AI Negotiation Coach (Phase 76)
	negotiateCache := negotiate.NewCache(24 * time.Hour)
	negotiateAgent := negotiate.NewAgent(d.provider)
	negotiateHandler := negotiate.NewHandler(d.shoppingRepo, negotiateAgent, negotiateCache)
	r.Get("/api/shopping-items/{id}/negotiate", negotiateHandler.Get)

	// Price Benchmark vs Market Average (Phase 70)
	benchmarkCache := benchmark.NewCache(2 * time.Hour)
	benchmarkAgent := benchmark.NewAgent(d.provider)
	benchmarkHandler := benchmark.NewHandler(d.shoppingRepo, benchmarkAgent, benchmarkCache)
	r.Get("/api/shopping-items/{id}/benchmark", benchmarkHandler.Get)

	// AI Price Forecast with Confidence Intervals (Phase 85)
	priceForecastCache := priceforecast.NewCache(2 * time.Hour)
	priceForecastAgent := priceforecast.NewAgent(d.provider)
	priceForecastHandler := priceforecast.NewHandler(priceForecastAgent, d.shoppingRepo, d.shoppingRepo, priceForecastCache)
	r.Get("/api/missions/{missionID}/shopping/{itemID}/forecast", priceForecastHandler.GetForecast)

	// AI Carbon Footprint Estimator (Phase 88)
	carbonCache := carbon.NewCache(24 * time.Hour)
	carbonAgent := carbon.NewAgent(d.provider)
	carbonHandler := carbon.NewHandler(carbonAgent, carbonCache)
	r.Get("/api/carbon", carbonHandler.GetEstimate)

	// AI Seasonal Discount Predictor (Phase 90)
	seasonalCache := seasonal.NewCache(24 * time.Hour)
	seasonalAgent := seasonal.NewAgent(d.provider)
	seasonalHandler := seasonal.NewHandler(seasonalAgent, seasonalCache)
	r.Get("/api/seasonal", seasonalHandler.GetPrediction)

	// Marketplace
	mktHandler := marketplace.NewHandler()
	mktHandler.RegisterRoutes(r)

	// Price Drop Watchlist (Phase 94)
	watchlistRepo := watchlist.NewRepository()
	watchlistHandler := watchlist.NewHandler(watchlistRepo)
	watchlistHandler.RegisterRoutes(r)

	// Mission Collaboration Threads (Phase 61)
	commentRepo := comment.NewRepository(d.pool)
	commentResolver := comment.NewBridgeResolver(func(ctx context.Context, slug string) (string, bool, error) {
		m, err := d.missionSvc.GetBySlug(ctx, slug)
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
	substituteAgent := substitute.NewAgent(d.provider)
	substituteGetter := substitute.NewOptionRepoGetter(d.optionRepo)
	substituteHandler := substitute.NewHandler(substituteGetter, substituteCache, substituteAgent)
	r.Get("/api/options/{id}/substitutes", substituteHandler.GetSubstitutes)

	// AI Mission Summary Card (Phase 72)
	summaryCache := summary.NewCache(time.Hour)
	summaryAgent := summary.NewAgent(d.provider)
	summaryHandler := summary.NewHandler(d.missionRepo, d.optionRepo, d.shoppingRepo, summaryAgent, summaryCache)
	r.Get("/api/missions/{slug}/summary", summaryHandler.Get)

	// Mission Health Score (Phase 57)
	healthCache := health.NewCache(30 * time.Minute)
	healthAgent := health.NewAgent(d.provider)
	healthHandler := health.NewHandler(d.missionRepo, d.optionRepo, d.shoppingRepo, healthAgent, healthCache)
	r.Get("/api/missions/{slug}/health", healthHandler.GetHealth)

	// Smart Budget Rebalancer (Phase 74)
	rebalancerAgent := rebalancer.NewAgent(d.provider)
	rebalancerHandler := rebalancer.NewHandler(envelopeRepo, d.missionRepo, rebalancerAgent)
	r.Get("/api/budget/rebalance", rebalancerHandler.Get)

	// Price Drop Weekly Digest (Phase 58)
	digestRepo := digest.NewRepository(d.pool)
	digestHandler := digest.NewHandler(digestRepo)
	r.Get("/api/digest/weekly", digestHandler.Weekly)

	// Real-time Currency Converter (Phase 98)
	currencySvc := currency.NewService()
	currencyHandler := currency.NewHandler(currencySvc)
	r.Get("/api/currency/rates", currencyHandler.GetRates)
	r.Get("/api/currency/convert", currencyHandler.Convert)

	// Price History Analytics (Phase 97)
	priceAnalyticsHandler := priceanalytics.NewHandler(d.pool)
	r.Get("/api/missions/{missionID}/items/{itemID}/price-stats", priceAnalyticsHandler.GetStats)

	// Spending Analytics Dashboard (Phase 123)
	spendingAnalyticsHandler := spendinganalytics.NewHandler(d.pool)
	r.Get("/api/analytics/spending", spendingAnalyticsHandler.GetAnalytics)

	// Budget Heatmap (Phase 132)
	budgetHeatmapHandler := budgetheatmap.NewHandler(d.pool)
	r.Get("/api/analytics/budget-heatmap", budgetHeatmapHandler.GetHeatmap)

	// Smart Budget Recommendations (Phase 96)
	budgetrecHandler := budgetrec.NewHandler(d.shoppingSvc)
	r.Get("/api/missions/{missionID}/budget-analysis", budgetrecHandler.GetAnalysis)

	// Smart Budget Planner with Purchase Sequencing (Phase 115)
	budgetPlannerHandler := budgetplanner.NewHandler(d.pool)
	r.Get("/api/missions/{id}/budget-plan", budgetPlannerHandler.GetPlan)

	// Smart Price Alert Rules (Phase 100) — in-memory, no DB dependency
	alertRuleRepo := pricealertrule.NewRepository()
	alertRuleHandler := pricealertrule.NewHandler(alertRuleRepo)
	r.Post("/api/items/{itemID}/alert-rules", alertRuleHandler.CreateRule)
	r.Get("/api/items/{itemID}/alert-rules", alertRuleHandler.ListRules)
	r.Delete("/api/items/{itemID}/alert-rules/{ruleID}", alertRuleHandler.DeleteRule)
	r.Post("/api/items/{itemID}/alert-rules/check", alertRuleHandler.CheckRules)

	// Product Review Aggregator (Phase 102) — deterministic, in-memory cache
	reviewSummaryHandler := reviewsummary.NewHandler()
	r.Get("/api/items/{itemID}/review-summary", reviewSummaryHandler.GetReviewSummary)

	// Loyalty Points & Cashback Tracker (Phase 103) — hardcoded French registry, no DB
	loyaltySvc := loyaltypoints.NewService()
	loyaltyHandler := loyaltypoints.NewHandler(loyaltySvc)
	r.Get("/api/loyalty/programs", loyaltyHandler.ListPrograms)
	r.Post("/api/loyalty/calculate", loyaltyHandler.Calculate)

	// Smart Shopping List Optimizer (Phase 104) — deterministic, no LLM
	shopOptimizerHandler := shoppingoptimizer.NewHandler(d.pool)
	r.Post("/api/missions/{missionID}/shopping/optimize", shopOptimizerHandler.Optimize)

	// French Market Price Intelligence (Phase 106) — hardcoded knowledge base, no LLM
	frenchMarketHandler := frenchmarket.NewHandler()
	r.Get("/api/market/insight", frenchMarketHandler.GetInsight)

	// Smart Budget Alerts (Phase 108) — direct pool query, no service layer
	budgetAlertHandler := budgetalert.NewHandler(d.pool)
	r.Get("/api/budget-alerts", budgetAlertHandler.GetAlerts)

	// Eco-Score & Sustainability Dashboard (Phase 109) — deterministic, no LLM
	ecoScoreHandler := ecoscore.NewHandler(d.shoppingSvc)
	r.Get("/api/missions/{missionID}/eco-score", ecoScoreHandler.GetEcoScore)

	// AI Purchase Advisor (Phase 110) — LLM tool-use, 2h in-memory cache
	purchaseAdvisorHandler := purchaseadvisor.NewHandler(d.shoppingRepo, d.provider)
	r.Post("/api/items/{itemID}/purchase-advice", purchaseAdvisorHandler.GetAdvice)

	// Price Drop Alert Digest (Phase 113) — 24h price change summary, direct pool query
	priceDigestHandler := pricedigest.NewHandler(d.pool)
	r.Get("/api/price-digest", priceDigestHandler.GetDigest)

	// Wishlist Social Sharing (Phase 114) — in-memory cache, 5 min TTL
	wishlistShareHandler := wishlistshare.NewHandler(d.pool)
	r.Get("/api/missions/{id}/wishlist-card", wishlistShareHandler.GetWishlistCard)

	// Price Negotiation Simulator (Phase 116) — deterministic scripts, 1h in-memory cache
	negotiationSimHandler := negotiationsim.NewHandler(d.shoppingRepo)
	r.Get("/api/missions/{missionId}/items/{itemId}/negotiation-sim", negotiationSimHandler.GetScript)

	// Contextual Negotiation Script Generator (Phase 150) — context-aware French scripts, 1h cache
	negotiationScriptHandler := negotiationscript.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/negotiation-script", negotiationScriptHandler.GetScript)

	// French Price Comparison Widget (Phase 153)
	priceComparisonHandler := pricecomparison.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/price-comparison", priceComparisonHandler.GetComparison)

	// Mission Progress Dashboard Widget (Phase 117) — 5min in-memory cache
	missionProgressHandler := missionprogress.NewHandler(d.pool)
	r.Get("/api/missions/{id}/progress", missionProgressHandler.GetProgress)

	// Activity Feed (Phase 121)
	activityFeedHandler := activityfeed.NewHandler(d.pool)
	r.Get("/api/activity-feed", activityFeedHandler.GetFeed)

	// Gift Finder Assistant (Phase 122) — French market catalog, 30min in-memory cache
	giftFinderHandler := giftfinder.NewHandler(d.pool)
	r.Get("/api/missions/{id}/gift-suggestions", giftFinderHandler.GetGiftSuggestions)

	// Weekly Price Alert Digest (Phase 124) — 1h in-memory cache
	weeklyDigestHandler := weeklydigest.NewHandler(d.pool)
	r.Get("/api/weekly-digest", weeklyDigestHandler.GetWeeklyDigest)

	// Price History Insights (Phase 118) — 2h in-memory cache
	priceInsightsHandler := priceinsights.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/price-insights", priceInsightsHandler.GetInsights)

	// Competitor Price Monitor (Phase 119) — deterministic FNV hash, 30min in-memory cache
	competitorPriceHandler := competitorprice.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/competitor-prices", competitorPriceHandler.GetCompetitorPrices)

	// Smart Coupon & Promo Finder (Phase 120) — deterministic FNV hash, 1h in-memory cache
	couponFinderHandler := couponfinder.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/coupons", couponFinderHandler.FindCoupons)

	// Smart Category Auto-tagger & Item Enrichment (Phase 125) — pure Go, 24h in-memory cache
	itemTaggerHandler := itemtagger.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/tags", itemTaggerHandler.GetTags)

	// Price Drop Streak Detector (Phase 126) — 1h in-memory cache
	priceStreakHandler := pricestreak.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/price-streak", priceStreakHandler.GetStreak)

	// Price Floor Predictor (Phase 129) — 2h in-memory cache
	priceFloorHandler := pricefloor.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/price-floor", priceFloorHandler.GetFloor)

	// Smart Duplicate Item Detector (Phase 128) — 30min in-memory cache
	duplicateDetectorHandler := duplicatedetector.NewHandler(d.pool)
	r.Get("/api/missions/{id}/duplicates", duplicateDetectorHandler.GetDuplicates)

	// Item Condition & Grade Tracker (Phase 131) — 2h in-memory cache
	conditionPricingHandler := conditionpricing.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/condition-pricing", conditionPricingHandler.GetConditionPricing)

	// Flash Sale Calendar — French market (Phase 130) — hardcoded, no DB
	saleCalendarHandler := salecalendar.NewHandler()
	r.Get("/api/sale-calendar", saleCalendarHandler.GetCalendar)

	// Smart Price Thresholds — Auto-Suggested Target Prices (Phase 133)
	targetSuggestionHandler := targetsuggestion.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/target-suggestion", targetSuggestionHandler.GetSuggestion)
	r.Post("/api/missions/{missionId}/items/{itemId}/target-suggestion/apply", targetSuggestionHandler.ApplySuggestion)

	// Price Drop Predictor ML-Style (Phase 134)
	dropReporterHandler := dropreporter.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/drop-prediction", dropReporterHandler.GetPrediction)

	// Smart Shopping List Sorter (Phase 136)
	listSorterHandler := listsorter.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/sorted-items", listSorterHandler.GetSortedItems)

	// Merchant Loyalty Tracker (Phase 135)
	loyaltyTrackerHandler := loyaltytracker.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/loyalty-summary", loyaltyTrackerHandler.GetLoyaltySummary)

	// Price Timeline Annotations (Phase 137)
	priceAnnotationHandler := priceannotation.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/price-annotations", priceAnnotationHandler.GetAnnotations)

	// French Public Holidays & Shopping Events (Phase 138) — static, no DB
	holidayAlertHandler := holidayalert.NewHandler()
	r.Get("/api/holidays-and-events", holidayAlertHandler.List)

	// Price Volatility Heatmap Calendar (Phase 139) — 2h cache
	volatilityCalendarHandler := volatilitycalendar.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/volatility-calendar", volatilityCalendarHandler.GetCalendar)

	// Mission ROI Calculator (Phase 140) — 30min in-memory cache
	roiHandler := roicalculator.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/roi", roiHandler.GetROI)

	// French Inflation Impact Tracker (Phase 141) — 30min in-memory cache
	inflationHandler := inflationtracker.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/inflation-impact", inflationHandler.GetInflationImpact)

	// Purchase Decision Matrix (Phase 142) — 15min in-memory cache
	decisionMatrixHandler := decisionmatrix.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/decision-matrix", decisionMatrixHandler.GetMatrix)

	// Smart Notification Rules Engine (Phase 143) — 5min in-memory cache
	smartAlertsHandler := notificationrules.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/smart-alerts", smartAlertsHandler.GetAlerts)

	// Collaborative Wishlist Voting Summary (Phase 144) — 10min in-memory cache
	voteSummaryHandler := wishlistvote.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/vote-summary", voteSummaryHandler.GetVoteSummary)

	// Price Elasticity & Demand Estimator (Phase 145) — 1h in-memory cache
	elasticityHandler := elasticity.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/elasticity", elasticityHandler.GetElasticity)

	// Mission Summary Text Report (Phase 146)
	missionReportHandler := missionreport.NewHandler(d.pool)
	r.Get("/api/missions/{missionID}/report", missionReportHandler.GetReport)

	// Multi-Mission Price Comparison Dashboard (Phase 147)
	crossMissionHandler := crossmission.NewHandler(d.pool)
	r.Get("/api/analytics/cross-mission", crossMissionHandler.GetComparison)

	// Smart Reorder & Repurchase Suggestions (Phase 148) — deterministic, 10min in-memory cache
	reorderHandler := reordersuggestion.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/reorder-suggestions", reorderHandler.GetSuggestions)

	// Real-Time Stock Availability Alert Simulator (Phase 149) — deterministic FNV hash, 30min in-memory cache
	stockAlertHandler := stockalert.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/stock-status", stockAlertHandler.GetStockStatus)

	// Negotiation Outcome Tracker (Phase 151)
	negotiationOutcomeHandler := negotiationoutcome.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/negotiation-outcomes", negotiationOutcomeHandler.GetOutcomes)

	// Smart Bundle Deal Detector (Phase 152)
	bundleDetectorHandler := bundledetector.NewHandler(d.pool)
	r.Get("/api/missions/{id}/bundle-deals", bundleDetectorHandler.GetBundles)

	// Smart Purchase Timing Score (Phase 154)
	timingRecommenderHandler := timingrecommender.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/timing-score", timingRecommenderHandler.GetScore)

	// Smart Deal Aggregator (Phase 155)
	dealAggregatorHandler := dealaggregator.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/deals", dealAggregatorHandler.GetDeals)

	// Budget Burn Rate Tracker (Phase 156)
	burnRateHandler := burnrate.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/burn-rate", burnRateHandler.GetBurnRate)

	// Smart Spending Velocity (Phase 167)
	spendingVelocityHandler := spendingvelocity.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/spending-velocity", spendingVelocityHandler.GetReport)

	// Smart Merchant Recommender (Phase 157)
	merchantRecHandler := merchantrecommender.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/merchant-recommendations", merchantRecHandler.GetRecommendations)

	// Smart Shopping List Optimizer (Phase 159)
	listOptimizerHandler := listoptimizer.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/optimized-list", listOptimizerHandler.GetOptimizedList)

	// French Cashback & Reward Tracker (Phase 160)
	cashbackTrackerHandler := cashbacktracker.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/cashback", cashbackTrackerHandler.GetCashbackSummary)

	// Smart Price Drop Watchlist (Phase 161)
	priceDropWatchlistHandler := pricedropwatch.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/price-drop-watchlist", priceDropWatchlistHandler.GetWatchlist)

	// Seasonal Buying Calendar (Phase 162)
	seasonalCalHandler := seasonalcalendar.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/seasonal-calendar", seasonalCalHandler.GetCalendar)

	// Smart Comparison Score (Phase 165)
	compScoreHandler := comparisonscore.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/comparison-score", compScoreHandler.GetReport)

	// Smart Wishlist Prioritizer (Phase 168)
	wishlistPrioritizerHandler := wishlistprioritizer.NewHandler(d.pool)
	r.Get("/api/wishlist/prioritized", wishlistPrioritizerHandler.GetPrioritized)

	// French Market Price Benchmark (Phase 169)
	frenchBenchmarkHandler := frenchbenchmark.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/french-benchmark", frenchBenchmarkHandler.GetBenchmark)

	// Mission Completion Scorecard (Phase 170)
	scorecardHandler := scorecard.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/scorecard", scorecardHandler.GetScorecard)

	// Smart Quantity Optimizer (Phase 171)
	quantityOptimizerHandler := quantityoptimizer.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/items/{itemId}/quantity-optimizer", quantityOptimizerHandler.GetReport)

	// Purchase Timeline Planner (Phase 172)
	timelinePlannerHandler := timelineplanner.NewHandler(d.pool)
	r.Get("/api/missions/{missionId}/purchase-timeline", timelinePlannerHandler.GetTimeline)

	// LLM pool health — always registered; returns empty pool when provider is not routing
	{
		var checker *llm.HealthChecker
		if d.smartRouter != nil {
			checker = llm.NewHealthChecker(d.smartRouter.Pool(), d.smartRouter.Breakers())
		} else {
			checker = llm.NewHealthChecker(llm.ModelPool{}, nil)
		}
		r.Get("/api/health/llm", llm.HealthHandler(checker))
	}

	// Prometheus metrics endpoint (Phase 14)
	if d.metricsHandler != nil {
		r.Handle("/metrics", d.metricsHandler)
	}
}
