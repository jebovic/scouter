# Backend Deep Dive

Go 1.23 backend with chi router, pgx/v5, golang-migrate, and Anthropic SDK.

---

## Package Structure

```
backend/
  cmd/server/
    main.go          # Entry point: env validation, DB pool, LLM router, graceful shutdown
    routes.go        # routeDeps struct + registerRoutes() — all 170+ routes
  internal/
    config/          # Config struct, Load() from env, startup validation
    db/              # pgx pool init, golang-migrate runner (embeds migrations/)
    httputil/        # WriteJSON / WriteError (buffer-first, no header split)
    llm/             # Provider interface + SmartRouter + OllamaEmbedder
    mission/         # Model, Repository, Service, Handler
    option/          # Research results CRUD
    shopping/        # Price tracking, deal intel, history
    notification/    # CRUD + mark-read + unread-count
    research/        # ResearchAgent
    pricing/         # PricingAgent
    decision/        # Decision engine
    dealintel/       # Trend + DealScore (pure Go, no DB deps)
    scheduler/       # robfig/cron background jobs
    template/        # 15 built-in templates, compiled in binary
    export/          # JSON export handler, share token management
    search/          # Semantic search (pgvector cosine ANN)
    embedding/       # Text builder, async worker (2 goroutines), repository
    purchase/        # PurchaseRecord CRUD, stats, mission auto-advance to "done"
    settings/        # JSONB key-value settings store
    admin/           # DELETE /api/data with X-Confirm header
    metrics/         # Recorder interface, PrometheusRecorder, middleware
    # 110+ specialized agent packages (phases 7–172)
    wishlistprioritizer/   # Phase 168
    frenchbenchmark/       # Phase 169
    scorecard/             # Phase 170
    quantityoptimizer/     # Phase 171
    timelineplanner/       # Phase 172
    # ... 105+ more
  migrations/
    001_initial.up.sql / 001_initial.down.sql
    002_options.up.sql / 002_options.down.sql
    # ... up to 022+
  Dockerfile
```

---

## Core Patterns

### Repository Pattern

Every domain has a Repository that wraps pgx:

```go
type MissionRepository struct {
    pool *pgxpool.Pool
}

func (r *MissionRepository) FindBySlug(ctx context.Context, slug string) (*Mission, error) {
    var m Mission
    err := r.pool.QueryRow(ctx,
        `SELECT id, slug, name, budget, status FROM missions WHERE slug = $1`,
        slug,
    ).Scan(&m.ID, &m.Slug, &m.Name, &m.Budget, &m.Status)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    return &m, err
}
```

Key conventions:
- `errors.Is(err, pgx.ErrNoRows)` → returns domain `ErrNotFound`
- All queries use `$1` parameter placeholders (SQL injection safe)
- Context propagated to every DB call

### Service Layer

Services orchestrate repositories and agents:

```go
type MissionService struct {
    repo    MissionRepository
    agent   *research.ResearchAgent
    embeds  chan<- embedding.Task
}

func (s *MissionService) RunResearch(ctx context.Context, missionID uuid.UUID) ([]option.Option, error) {
    mission, err := s.repo.FindByID(ctx, missionID)
    // ... orchestrate agent, persist results, emit embed tasks
}
```

### Handler Pattern

```go
func (h *Handler) handleResearch(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
        return
    }
    opts, err := h.service.RunResearch(r.Context(), id)
    if err != nil {
        httputil.WriteError(w, http.StatusInternalServerError, err.Error())
        return
    }
    httputil.WriteJSON(w, http.StatusOK, opts)
}
```

---

## AI Agents

### ResearchAgent (`internal/research/`)

1. Fetches mission + constraints from DB
2. Builds structured prompt with constraint schema as tool definition
3. Calls `SmartRouter.Complete()` with `WithRequestOpts(ctx, RequestOpts{Capability: "tool-use"})`
4. Parses `tool_calls` from response → extracts option structs
5. Batch inserts options via Repository
6. Sends each option text to embed channel for async vectorization

### PricingAgent (`internal/pricing/`)

1. Loads options for mission
2. Builds price-hunting prompt (merchant list, price format tool schema)
3. Calls SmartRouter → parses prices per merchant
4. Calculates TCO (price + shipping + warranty + accessories)
5. Scores each deal (0–100): price vs market, trend, availability
6. Updates shopping_items table with prices + deal scores

### Specialized Agents (140+)

Each specialized agent follows the same pattern:
- Single `GET /api/missions/:id/<capability>` endpoint
- Handler validates params → calls agent
- Agent either: (a) queries DB + computes pure Go, or (b) calls SmartRouter
- Response cached (10–30min, in-process map with FNV-32a key)

Example — Scorecard (Phase 170):

```go
func (a *ScorecardAgent) Score(ctx context.Context, missionID uuid.UUID) (*Scorecard, error) {
    // 4 DB queries: price efficiency, research depth, time to decision, budget discipline
    // Returns grade A/B/C/D + breakdown
    // Cached 30min with missionID as key
}
```

---

## LLM Layer

See [LLM Routing](llm-routing.md) for full SmartRouter documentation.

### Provider Interface

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

type CompletionRequest struct {
    Model       string
    Messages    []Message
    Tools       []Tool       // tool use schema
    MaxTokens   int
    Temperature float64
}

type CompletionResponse struct {
    Content   string
    ToolCalls []ToolCall
    StopReason string
}
```

### AnthropicProvider

Uses the official Anthropic Go SDK with `param.NewOpt(v)` for optional fields:

```go
resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
    Model:     anthropic.Model(req.Model),
    Messages:  convertMessages(req.Messages),
    Tools:     param.NewOpt(convertTools(req.Tools)),
    MaxTokens: int64(req.MaxTokens),
})
```

### OllamaProvider

Calls local Ollama HTTP API at `OLLAMA_BASE_URL`. Supports heavy, fast, and embed models.

---

## Background Jobs

### Price-Check Scheduler (`internal/scheduler/`)

```go
c := cron.New()
c.AddFunc("@every 1h", func() {
    missions := repo.FindActive(ctx)
    for _, m := range missions {
        agent.CheckPrices(ctx, m.ID)
        // fires notifications if target price crossed
    }
})
c.Start()
```

### Embed Worker (`internal/embedding/`)

```go
// 2 goroutines consume from channel
for i := 0; i < 2; i++ {
    go func() {
        for task := range embedCh {
            vec, _ := embedder.Embed(ctx, task.Text)
            repo.UpsertEmbedding(ctx, task.OptionID, vec)
        }
    }()
}
```

---

## Error Handling

| Error Type | Handling |
|-----------|---------|
| Not found | `errors.Is(err, pgx.ErrNoRows)` → `ErrNotFound` → `404` |
| Validation | `httputil.WriteError(w, 400, msg)` |
| Internal | `httputil.WriteError(w, 500, "internal server error")` — never leaks DB details |
| LLM timeout | SmartRouter cascades to next provider |
| Circuit open | Skip that provider, try next |

---

## Metrics (Phase 14)

The `metrics.Recorder` interface is instrumented on:
- SmartRouter: request count, latency, provider selection
- All agents: call count, error rate, response time
- Scheduler: price check count, alert count
- HTTP middleware: request count, latency by route

```go
type Recorder interface {
    RecordRequest(provider, model string, duration time.Duration, err error)
    RecordAgentCall(agent string, duration time.Duration, err error)
}
```

Enable with `METRICS_ENABLED=true` → `GET /metrics` serves Prometheus text format.

---

## Route Registration (`cmd/server/routes.go`)

Routes are registered via `routeDeps` struct:

```go
type routeDeps struct {
    missionHandler    *mission.Handler
    optionHandler     *option.Handler
    researchHandler   *research.Handler
    pricingHandler    *pricing.Handler
    // ... 60+ handler fields
}

func registerRoutes(r chi.Router, d routeDeps) {
    r.Route("/api", func(r chi.Router) {
        r.Get("/health", d.healthHandler.Handle)
        r.Route("/missions", func(r chi.Router) {
            r.Get("/", d.missionHandler.List)
            r.Post("/", d.missionHandler.Create)
            r.Route("/{slug}", func(r chi.Router) {
                r.Get("/", d.missionHandler.Get)
                r.Patch("/", d.missionHandler.Update)
                // ... 40+ sub-routes per mission
            })
        })
        // ... all other routes
    })
}
```

`main.go` now imports only ~35 packages (reduced from 142 via `routes.go` extraction).
