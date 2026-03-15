# SCOUTER — Release Roadmap

> **Active plan for the first journey.** Read this at the start of every session.
> Current status: Phases 1–41 complete. Auto-improvement loop v3 active — architecting Phase 42+.

## Phase Implementation Workflow (repeat for every phase)

```
1. /everything-claude-code:plan + architect  →  detailed plan for the phase
2. follow ECC tdd workflow to implerment the phase with specialized agents (go coding for backend, frontend agent with /frontend-design and /frontend-patterns skills for frontend), parallelize where possible
3. run tests  (backend Go tests)
4. ecc go review
5. npm run build + npm run typecheck  (frontend)
6. frontend-design review phase frontend changes
7. update documentations (readme, claude.md, roadmap)
8. Architect review front, back, architectural direction
9. fix
10. commit and push everything
11. close phase
```

---

## Phase Status

| Phase | Name | Type | Priority | Status |
|-------|------|------|----------|--------|
| 1+2 | Core backend + frontend | Functional | — | ✅ Done |
| 3 | Decision Engine & Scoring | Functional | Critical | ✅ Done |
| 4 | Polish & Hardening | Infrastructure | Critical | ✅ Done |
| **5** | Agent Feedback Loop | Agent | **High** | ✅ Done |
| 6 | Interface Overhaul (responsive, onboarding) | Interface | High | ✅ Done |
| 7 | Price Alerts & Deal Intelligence | Functional | High | ✅ Done |
| 8 | Mission Templates & Quick-Start | Interface | Medium | ✅ Done |
| 9 | Ollama Smart Routing & Model Optimization | Infrastructure | High | ✅ Done |
| 10 | Export, Share & Archive | Functional | Medium | ✅ Done |
| **11** | **Semantic Search (pgvector)** | **Agent + Infra** | **Medium** | ✅ Done |
| 12 | Mission Lifecycle & Post-Purchase | Functional | Medium | ✅ Done |
| 13 | Settings, Data Management & Deployment | Infrastructure | Low | ✅ Done |
| **14** | **Observability & Monitoring Stack** | **Infrastructure** | **Medium** | ✅ Done |
| 15 | Real-Time Price Intelligence (Open Food Facts) | Functional | High | ✅ Done |
| 16 | Collaborative Missions (invite + voting) | Functional | High | ✅ Done |
| 17 | AI Negotiation Coach | Agent | Medium | ✅ Done |
| 18 | PWA + Mobile UX (offline, swipe, BottomNav) | Interface | Medium | ✅ Done |
| 19 | Multi-Currency & i18n Completion | Interface | Low | ✅ Done |
| **20** | **Interactive Analytics Dashboard (recharts)** | **Interface** | **High** | ✅ Done |
| 21 | Wish List & Price Drop Alerts | Functional | High | ✅ Done |
| 22 | Weighted Comparison Matrix | Functional | Medium | ✅ Done |
| 23 | Smart Budget Forecaster (AI) | Agent | Medium | ✅ Done |
| 24 | Price Alert Toggle + French Market UX | Functional | High | ✅ Done |
| 25 | EAN Barcode Lookup (Open Food Facts) | Functional | Medium | ✅ Done |
| 26 | Spending Persona (AI buyer archetype) | Agent | Medium | ✅ Done |
| 27 | Price History Chart (recharts LineChart) | Interface | Medium | ✅ Done |
| **28** | **Smart Notifications Center** | **Interface** | **High** | ✅ Done |
| 29 | French Market Integrations (travel APIs) | Functional | High | ✅ Done |
| 30 | Social Proof & Review Aggregation | Functional | Medium | ✅ Done |
| 31 | Product Price Comparison (LeLynx / idealo FR) | Functional | High | ✅ Done |
| 32 | AI Purchase Timing Advisor | Agent | Medium | ✅ Done |
| 33 | Mission Export to PDF/MD/JSON | Interface | Low | ✅ Done |
| 34 | Dark/Light Theme Toggle | Interface | Low | ✅ Done |
| 35 | Budget Envelope (envelope budgeting method) | Functional | Medium | ✅ Done |
| 36 | Purchase-to-Envelope Linking (actual spend) | Functional | High | ✅ Done |
| 37 | French Seasonal Deal Calendar | Interface | Medium | ✅ Done |
| 38 | Smart Budget Alerts (envelope threshold push) | Functional | High | ✅ Done |
| 39 | Receipt Scanner (OCR upload) | Functional | Medium | ✅ Done |
| 40 | Quick Mission Duplicate / Template from Mission | Interface | Low | ✅ Done |
| 41 | Merchant Affiliate Deep Links (Fnac, Darty, Boulanger) | Functional | Medium | ✅ Done |
| 42 | Real-time Stock Availability (French retailers) | Functional | High | ✅ Done |
| 43 | AI Shopping Summary Report (PDF/email) | Agent | Medium | ✅ Done |
| 44 | Voice Input for Mission Creation (Web Speech API) | Interface | Low | ✅ Done |
| 45 | Performance Dashboard (Lighthouse + bundle analysis) | Infrastructure | Low | ✅ Done |
| 46 | Smart Comparison Mode (side-by-side option viewer) | Interface | High | ✅ Done |
| 47 | Mission AI Coach (proactive tips during research) | Agent | Medium | ✅ Done |
| 48 | French VAT Calculator (TVA 20%/5.5% per category) | Functional | Medium | ✅ Done |
| 49 | Gamification: Scouter Badges & Saving Milestones | Interface | Low | ✅ Done |
| 50 | Multi-Mission Budget Rollup Dashboard | Interface | High | ✅ Done |
| 51 | Mission Timeline & Activity Feed | Interface | Medium | ✅ Done |
| 52 | Smart Price Prediction (trending extrapolation) | Agent | High | ✅ Done |
| 53 | Collaborative Wishlist Sharing (invite link) | Functional | Medium | ✅ Done |
| 54 | AI Product Substitute Finder (French market) | Agent | High | ✅ Done |
| 55 | Seasonal Savings Calendar (French promotions) | Interface | Medium | ✅ Done |
| 56 | Quick Note / Annotation on Options | Interface | Low | ✅ Done |
| 57 | Mission Health Score (AI composite rating) | Agent | High | ✅ Done |
| 58 | Price Drop Email Digest (weekly summary) | Functional | Medium | 📋 Planned |
| 59 | Smart Retailer Radar (live stock + price aggregator) | Functional | High | ✅ Done |
| 60 | AI Deal Explainer (why is this a good deal?) | Agent | Medium | ✅ Done |
| 61 | Mission Collaboration Threads (inline comments) | Functional | Medium | ✅ Done |
| 62 | Option Price Alert (per-option threshold) | Functional | High | ✅ Done |
| 63 | French Public Holidays & Closing Days Widget | Interface | Low | ✅ Done |
| 64 | Loyalty Points Tracker (Fnac, Cdiscount) | Functional | Medium | ✅ Done |
| 65 | Smart Category Auto-Tagging (AI) | Agent | Medium | ✅ Done |
| 66 | Budget Variance Heatmap (month over month) | Interface | Medium | ✅ Done |
| 67 | Mission Cloning with Deep Copy | Functional | Low | ✅ Done |
| 68 | AI Shopping List Optimizer (best order to buy) | Agent | High | ✅ Done |
| 69 | Smart Receipt Analyzer (AI itemization) | Agent | Medium | ✅ Done |
| 70 | Price Benchmark vs Market Average (AI) | Agent | High | ✅ Done |
| 71 | Seasonal Price Calendar (buy at the right time) | Interface | Medium | ✅ Done |
| 72 | AI Mission Summary Card (executive brief) | Agent | Medium | ✅ Done |
| 73 | Multi-Currency Live Converter (ECB rates) | Functional | Medium | ✅ Done |
| 74 | Smart Budget Rebalancer (AI envelope optimizer) | Agent | High | ✅ Done |
| 75 | Product Comparison Matrix Export (CSV) | Functional | Medium | ✅ Done |
| 76 | AI Negotiation Coach (haggling tips) | Agent | Medium | ✅ Done |
| 77 | Purchase Timeline Gantt (when to buy each item) | Interface | Medium | ✅ Done |
| 78 | Smart Deal Aggregator (French promo feeds) | Integration | High | 📋 Planned |
| 79 | AI Substitution Suggester (cheaper alternatives) | Agent | High | ✅ Done |
| 80 | Mission Goal Tracker (progress vs deadline) | Interface | Medium | 📋 Planned |
| 81 | Smart Price History Export (CSV download) | Functional | Low | 📋 Planned |
| 82 | AI Shopping Persona Insights (spending archetype) | Agent | Medium | 📋 Planned |
| 83 | French Retailer Promo API Integration (Dealabs feed) | Integration | High | 📋 Planned |
| 84 | Collaborative Wishlist Voting (multi-user) | Functional | High | 📋 Planned |

**Execution order:** 3+4 in parallel → 5+6 in parallel → 7 → 8 → **9** → 10+11 in parallel → 12+13 in parallel → **14** → 15–19 sequential → **20–27** auto-improvement loop → **28–30** ongoing → **31–35** auto-improvement loop v2 → **36–45** auto-improvement loop v3

---

## Phase 3: Decision Engine & Scoring

**Type**: Functional | **Priority**: Critical | **Complexity**: High
**Depends on**: nothing (can start immediately)

**User Value**: Transforms SCOUTER from a data collector into a decision tool — weighted 0-100 score per option, answering "which one should I actually buy?"

### Features
- Weighted scoring algorithm: constraints + attributes → composite score
- User-configurable weight sliders per dimension (price vs quality vs features)
- Auto-rank options by composite score with visual indicators
- **DecisionAgent** — new LLM agent, natural-language recommendation from scored options
- Hard constraints gate options out; soft constraints reduce score
- Budget fit factor (sweet-spot scoring, not just cheapest)

### Backend
- New `internal/decision/` package: `ScoreEngine` (pure Go) + `DecisionAgent` (LLM)
- `POST /api/missions/:id/decide` — runs scoring + decision agent
- `GET /api/missions/:id/decision` — returns cached result
- New `decisions` table
- `ALTER TABLE missions ADD COLUMN weight_profile JSONB DEFAULT '{}'`

### Frontend
- `DecisionPanel` on MissionOverview (top recommendation card)
- Weight slider UI per attribute dimension
- Score badge on OptionCard (0-100, color gradient)
- Score column in ComparisonTable (sortable)
- "Run Decision" button on MissionOverview

### DB migrations
```sql
CREATE TABLE decisions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  scores JSONB NOT NULL,
  summary TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE missions ADD COLUMN weight_profile JSONB NOT NULL DEFAULT '{}';
```

---

## Phase 4: Polish & Hardening

**Type**: Infrastructure + Interface | **Priority**: Critical | **Complexity**: Medium
**Depends on**: nothing (can run parallel to Phase 3)

**User Value**: Eliminates hangs, crashes, and silent failures. Trustworthy for decisions worth hundreds or thousands of dollars.

### Backend fixes
- Add `go-playground/validator` struct tags to all `CreateRequest`/`UpdateRequest` types
- Validation middleware or per-handler validation before service calls
- `context.WithTimeout(ctx, 60*time.Second)` wrapping all LLM `Complete` calls
- Graceful shutdown: `signal.NotifyContext` + `httpServer.Shutdown(ctx)`
- Change `PUT /api/missions/{slug}` → `PATCH /api/missions/{slug}`
- Cursor-based pagination (`?cursor=&limit=`) on all list endpoints

### Frontend fixes
- Add Zod `.parse()` in every `api/*.ts` after `apiFetch` (fulfills CLAUDE.md promise)
- `ErrorBoundary` component wrapping each route and major section
- `ToastProvider` + `useToast` hook via React context
- Toast wired into all mutation hooks (`onSuccess`, `onError`)
- Pass `mission.currency` to `PriceHistoryModal` (remove hardcoded `"USD"`)

---

## Phase 5: Agent Feedback Loop & Refinement ✅ COMPLETE

**Type**: Agent + Functional | **Priority**: High | **Complexity**: High
**Depends on**: Phase 3

**Completed**: 2026-03-14. Commits: `08e947a` (frontend fixes), `1b67ba1` (architect review fixes H1/H2/H3/M4).

**Deferred to Phase 7** (noted for context):
- M1: `recordRun` duplicated in research/pricing — consolidate into shared helper
- M2: price diff tracking (numeric delta) — belongs with Phase 7 Deal Intelligence
- M3: aggressive LLM name normalization in `normalize.go` — Phase 7 when real data available
- L1/L2/L3: FeedbackInput dedup, shopping diff items, normalize_test coverage — minor cleanup

**User Value**: Iterative refinement — "I liked A and C, find more like A" instead of one-shot LLM calls.

### Features
- Research/pricing endpoints accept optional `{"feedback": "...", "pinned": [...]}`
- Prompts include previous results + feedback (multi-turn context)
- **Pin** options/items — excluded from deletion on re-run
- **Reject** options with reason (fed as negative examples)
- Agent run history with diff view (new / removed / changed)

### Backend
- Migration 005: new `agent_runs` table (mission_id, agent_type, input_snapshot JSONB, result_snapshot JSONB, feedback, created_at); `pinned`/`rejected`/`reject_reason` on `options`; `pinned` on `shopping_items`
- New `internal/agentrun/` package (model, repository, handler)
- `GET /api/missions/:id/agent-runs`
- `PATCH /api/missions/:id/options/:id/pin`, `PATCH /api/missions/:id/options/:id/reject`
- `PATCH /api/missions/:id/shopping/:id/pin`
- `DeleteByMission` on options/shopping skips pinned rows
- ResearchAgent + PricingAgent: accept optional `FeedbackInput`; include pinned/rejected context in prompt; record agent run after each invocation

### Frontend
- `FeedbackModal` after each agent run (textarea + submit/skip)
- Pin/unpin toggle + `PINNED` chip on `OptionCard` and `ShoppingItemRow`
- `RejectModal` + `REJECTED: {reason}` chip on `OptionCard`
- `AgentRunHistory` accordion on MissionOverview
- `DiffBadge` (NEW / REMOVED / CHANGED) on option and shopping item cards
- New hooks: `useAgentRuns`, `useOptionActions`, `useShoppingActions`, `useDiffIndicators`

### Risks
- Prompt token budget: cap pinned/rejected summaries to 10 items, feedback to 500 chars
- Stale pinned accumulation: plan "clear all pinned" action
- Diff fragility on LLM name variation: normalize + structural fallback

---

## Phase 6: Interface Overhaul — Responsive, Onboarding, Empty States

**Type**: Interface | **Priority**: High | **Complexity**: Medium
**Depends on**: Phase 4

**User Value**: Usable on phones; new users know what to do; every empty state is actionable.

### Features
- Mobile-responsive layouts (breakpoints: 640px, 1024px)
- CSS modules replacing all inline `style={{}}`
- 3-step onboarding overlay for first-time users (localStorage dismissed)
- Rich empty states with embedded CTAs ("No options yet — Run Research Agent")
- Skeleton loading (card-shaped, row-shaped, chart-shaped) replacing spinner-only states
- Keyboard shortcuts: `N` new mission, `R` research, `P` pricing, `/` search
- Breadcrumb navigation; collapsible mission list sidebar

---

## Phase 7: Price Alerts & Deal Intelligence

**Type**: Functional + Agent | **Priority**: High | **Complexity**: High
**Depends on**: Phase 4

**User Value**: Proactive monitoring — SCOUTER tells you when "the laptop dropped 15%" instead of you checking manually.

### Features
- Target price per shopping item (alert threshold)
- Cron job re-runs PricingAgent on active missions periodically
- Price trend per item: "dropping / stable / rising" from snapshot history
- Deal score: current vs historical average vs target (% below)
- In-app notification feed in Topnav bell icon

### Backend
- `target_price NUMERIC` on `shopping_items`
- New `internal/scheduler/` with `robfig/cron`
- New `notifications` table + `internal/notification/` package
- `GET /api/notifications`, `PATCH /api/notifications/:id/read`

### Frontend
- Target price inline edit on ShoppingItemRow
- Trend arrow + deal score badge
- Notification bell with unread count in Topnav
- Price trend sparkline per item

---

## Phase 8: Mission Templates & Quick-Start Flows

**Type**: Interface + Functional | **Priority**: Medium | **Complexity**: Low
**Depends on**: Phase 6

**User Value**: Common purchases (laptop, TV, trip) pre-fill constraints in 30 seconds.

### Features
- 10-15 built-in templates (compiled into binary, no DB/LLM)
- Each template: name, icon, category, constraints, cost categories, weight profile
- "Start from template" gallery on HQDashboard
- Template preview modal before applying

### Backend
- New `internal/template/` package (hardcoded Go structs)
- `GET /api/templates`, `GET /api/templates/:slug`

### Frontend
- `TemplateCard`, `TemplatePreview` components
- MissionForm accepts template as initial values

---

## Phase 9: Ollama Smart Routing & Model Optimization

**Type**: Infrastructure | **Priority**: High | **Complexity**: Medium
**Depends on**: Phase 8 (structurant — makes all agents smarter before building on them)

**User Value**: Turns the Ollama-only constraint into an asset. Agents automatically use the best available local model, fall back gracefully across the pool (heavy → fast → cloud), and surface model identity + quality degradation so the user always knows what ran.

### Design Principles
- `Provider` interface **unchanged** — routing hints carried via `context.Context` (`RequestOpts`), not on `CompletionRequest`
- JSON fallback stays in **agents** (they own their schemas); shared helper `RetryAsJSON` in `internal/llm/`
- **Capability-matched priority pool** instead of named tiers (heavy/fast/cloud = defaults, configurable)
- Separate `EmbedProvider` interface stub prepared for Phase 11 (pgvector)

### Features
- Multi-model Ollama pool: configure 2-3 models via env; each has capability flags (tool-use, long-context)
- `SmartRouter` routes `CapToolUse` requests to tool-capable models only; cascades on infra error
- Ollama cloud model as last-resort (API key auth, separate rate limiter + token budget)
- JSON-mode fallback in ResearchAgent/PricingAgent: on tool-parse failure, retry without tools asking for JSON
- `GET /api/health/llm` — per-model status (healthy, circuit state, last latency)
- Structured `slog` per LLM call: model name, latency, tokens, was_fallback, degraded
- Frontend: `LLMStatus` dot in Topnav polls health every 60s

### Backend
- `internal/llm/requestopts.go` — `RequestOpts`, `WithRequestOpts(ctx)`, `GetRequestOpts(ctx)`, `Capability` bitmask
- `internal/llm/pool.go` — `ModelEntry` + `ModelPool` with `ForCapabilities()` filter
- `internal/llm/smart_router.go` — replaces `RoutingProvider` as default; circuit breaker per model
- `internal/llm/fallback.go` — `RetryAsJSON` shared helper
- `internal/llm/health.go` — `PoolHealth`, `ModelStatus`
- `internal/llm/ollama.go` — add optional API key header + `Ping(ctx)` method
- `internal/llm/provider.go` — add `ModelName`, `Degraded`, `Attempts` to `CompletionResponse`; add `EmbedProvider` stub
- `internal/config/config.go` — new env vars; `OLLAMA_MODEL` backward-compat alias
- `cmd/server/main.go` — wire `SmartRouter`, add `/api/health/llm`
- Research + Pricing agents — `WithRequestOpts(ctx, CapToolUse)` + `RetryAsJSON` fallback
- Decision agent — `WithRequestOpts(ctx, label)` + wire existing text fallback through helper

### Model Selection (research-backed, 2025)

| Role | Model | Pull tag | Size | Notes |
|------|-------|----------|------|-------|
| Heavy (tool use) | Qwen3 14B | `qwen3:14b` | 8.8 GB | Best tool-call fidelity at 14B; beats qwen2.5:14b; add `/no_think` in system prompt for agent loops |
| Fast (text only) | Qwen3 4B | `qwen3:4b` | 2.4 GB | Near-14B quality at 4B; also supports tool use if needed |
| Embeddings | mxbai-embed-large | `mxbai-embed-large` | 638 MB | Exactly 1024 dims — zero schema change; SOTA on MTEB English |
| Cloud fallback | DeepSeek V3.2 | `deepseek-v3.2:cloud` | cloud | 671B via Ollama cloud API; same `/api/chat` endpoint, `Authorization: Bearer` header |

**Gotchas confirmed by research:**
- `deepseek-r1` in Ollama registry has no tool-calling template (GitHub #8517, #12719) — do NOT use for agents
- `nomic-embed-text` (default/latest) = 768 dims — wrong for our `vector(1024)` schema; use `mxbai-embed-large` instead
- `phi4` and `gemma3` tool-call templates are broken in Ollama; fine for chat, not for agent loops
- Ollama cloud: endpoint `https://ollama.com/api/chat`, same request format as local, add `Authorization: Bearer $OLLAMA_CLOUD_API_KEY`

### New Environment Variables
| Variable | Default | Notes |
|---|---|---|
| `OLLAMA_HEAVY_MODEL` | `qwen3:14b` | Primary, tool-use capable (was `qwen2.5:7b`) |
| `OLLAMA_FAST_MODEL` | `qwen3:4b` | Lighter fallback for text tasks |
| `OLLAMA_HEAVY_TIMEOUT` | `180` | Seconds |
| `OLLAMA_FAST_TIMEOUT` | `60` | Seconds |
| `OLLAMA_CLOUD_URL` | (empty) | `https://ollama.com` when enabled |
| `OLLAMA_CLOUD_MODEL` | (empty) | e.g. `deepseek-v3.2:cloud` |
| `OLLAMA_CLOUD_API_KEY` | (empty) | Bearer token from ollama.com/settings/tokens |
| `OLLAMA_CLOUD_RPM` | `10` | Rate limit for cloud |
| `OLLAMA_EMBED_MODEL` | `mxbai-embed-large` | Phase 11 prep — 1024 dims, no schema change |
| `OLLAMA_MODEL` | `qwen2.5:7b` | Legacy alias → heavy model |

### Frontend
- `LLMStatus` dot component in Topnav (green/yellow/red per model, polls `/api/health/llm` every 60s)

### No DB Changes

---

## Phase 10: Export, Share & Archive

**Type**: Functional | **Priority**: Medium | **Complexity**: Medium
**Depends on**: Phase 4

**User Value**: Share research with partner/friend; archive completed missions cleanly.

### Features
- Export: PDF (full report), JSON (backup), Markdown (paste into notes)
- Shareable read-only UUID-token link (no auth required)
- Archive mission (hidden from dashboard, recoverable)

### Backend
- New `internal/export/` package (PDF, JSON, Markdown)
- `GET /api/missions/:id/export?format=pdf|json|markdown`
- `POST /api/missions/:id/share`, `GET /api/shared/:token`
- `share_token TEXT UNIQUE` + `archived_at TIMESTAMP` on missions

### Frontend
- Export dropdown on MissionOverview
- Share button → clipboard + toast
- `/shared/:token` read-only page
- Archive/unarchive on MissionCard and MissionOverview

---

## Phase 11: Semantic Search & Smart Suggestions (pgvector)

**Type**: Agent + Infrastructure | **Priority**: Medium | **Complexity**: High
**Depends on**: Phase 5 (and benefits from `OLLAMA_EMBED_MODEL` configured in Phase 9)

**User Value**: "Find that laptop with good battery under $1000" — natural language search across all missions + cross-mission suggestions.

### Features
- Embed all options via Voyage AI v3 (activates existing `embedding vector(1024)` column)
- `GET /api/search?q=...` — vector similarity search
- `GET /api/options/:id/similar` — top-5 similar options across missions
- Async embedding on option creation (does not block)
- IVFFlat index on pgvector column

### Frontend
- Debounced search bar in Topnav with instant dropdown
- `/search` full-page results
- "Similar items" on OptionCard expanded view

---

## Phase 12: Mission Lifecycle & Post-Purchase Tracking

**Type**: Functional + Interface | **Priority**: Medium | **Complexity**: Medium
**Depends on**: Phase 3

**User Value**: Close the loop — record what you actually paid, rate satisfaction, build personal purchase history.

### Features
- Record actual purchase (date, final price, merchant, chosen option)
- Satisfaction rating (1-5) + review text
- "Lessons learned" field on mission
- Mission timeline view (research → decision → purchase → review)
- `/history` page — all completed missions with outcomes
- `/stats` page — total spent, savings vs budget, category breakdown

### Backend
- New `purchase_records` table + `internal/purchase/` package
- `POST /api/missions/:id/purchase`, `GET /api/missions/:id/purchase`
- `GET /api/stats`
- `lessons TEXT` on missions

### Frontend
- Purchase form on MissionOverview (visible in "buying"/"done" phase)
- Star rating component
- Vertical timeline visualization
- New `/history` and `/stats` routes + pages

---

## Phase 13: Settings, Data Management & Deployment Polish

**Type**: Infrastructure | **Priority**: Low | **Complexity**: Medium
**Depends on**: Phase 10

**User Value**: Configure preferences, manage data, deploy with confidence.

### Features
- `/settings` page: default currency, locale, LLM provider toggle, theme
- Data import (JSON backup restore) + reset ("delete all data")
- Enhanced `/api/health` with component status (DB ping, LLM ping)
- Structured logging with correlation IDs
- LLM cost estimator shown before running an agent
- Docker Compose production hardening

### Backend
- New `settings` table (key TEXT PK, value JSONB)
- `GET /api/settings`, `PATCH /api/settings`
- `POST /api/import`, `DELETE /api/data`

### Frontend
- `/settings` page
- Import/export section, danger zone with confirmation modal
- Cost estimate tooltip on agent buttons

---

## Phase 14: Observability & Monitoring Stack

**Type**: Infrastructure | **Priority**: Medium | **Complexity**: Medium
**Depends on**: Phase 9 (SmartRouter is the main instrumentation target)

**User Value**: Full operational visibility — know when Ollama is struggling, which model is slow, how many alerts fired, and what each container is consuming. Pre-provisioned dashboards require zero setup.

### Design Decisions
- **Metrics interface**: domain-scoped sub-interfaces (`LLMRecorder`, `AgentRecorder`, `SchedulerRecorder`, `HTTPRecorder`) injected via constructors — no global Prometheus registry
- **HTTP path labels**: `chi.RouteContext().RoutePattern` — returns `/api/missions/{slug}`, not actual UUIDs; bounded cardinality (~15 routes)
- **Compose structure**: `profiles: ["monitoring"]` in existing `docker-compose.yml` — matches existing `seed` profile pattern
- **Token tracking**: instrumented in `SmartRouter` only (single chokepoint, captures fallbacks)
- **Alertmanager**: skipped initially — use Grafana alert panels instead
- **NoopRecorder** as default in all constructors — no nil checks needed

### Backend — `backend/internal/metrics/`

New package with 5 files:
- `recorder.go` — four sub-interfaces (`LLMRecorder`, `AgentRecorder`, `SchedulerRecorder`, `HTTPRecorder`)
- `prometheus.go` — `PrometheusRecorder` implementing all interfaces; custom `prometheus.Registry` (not global)
  - `scouter_llm_calls_total` (CounterVec: model, label, result)
  - `scouter_llm_call_duration_seconds` (HistogramVec: model; buckets 0.5→180s)
  - `scouter_llm_tokens_total` (CounterVec: model, direction=input|output)
  - `scouter_llm_fallbacks_total` (CounterVec: model)
  - `scouter_agent_runs_total` (CounterVec: agent_type, result)
  - `scouter_scheduler_runs_total` (CounterVec: job_type, result)
  - `scouter_scheduler_alerts_triggered_total` (Counter)
  - `scouter_http_requests_total` (CounterVec: method, route, status)
  - `scouter_http_request_duration_seconds` (HistogramVec: method, route)
  - `scouter_active_missions` (GaugeFunc — queried at scrape time)
  - `scouter_unread_notifications` (GaugeFunc — queried at scrape time)
- `noop.go` — `NoopRecorder` (all methods no-ops; used in tests and when `METRICS_ENABLED=false`)
- `middleware.go` — chi HTTP middleware; wraps `ResponseWriter` to capture status code
- `recorder_test.go` + `middleware_test.go` — unit tests via `prometheus/testutil`

### Modified Files

| File | Change |
|------|--------|
| `backend/internal/llm/smart_router.go` | Add `LLMRecorder` field + `WithRecorder` option; instrument `Complete()` (latency, tokens, fallback, circuit events) |
| `backend/internal/research/agent.go` | Add `AgentRecorder`; record end-to-end run success/failure |
| `backend/internal/pricing/agent.go` | Same |
| `backend/internal/decision/agent.go` | Same |
| `backend/internal/scheduler/orchestrator.go` | Add `SchedulerRecorder`; record missionsChecked + alertsTriggered per run |
| `backend/cmd/server/main.go` | Wire `PrometheusRecorder`; register `/metrics`; add HTTP middleware; pass recorder to all agents/router/scheduler |
| `backend/go.mod` | Add `github.com/prometheus/client_golang` |

### Docker Compose — `profiles: ["monitoring"]`

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `prometheus` | `prom/prometheus:v3.3.1` | 9090 | Scrapes backend (30s) + cAdvisor (30s) |
| `grafana` | `grafana/grafana:11.6.0` | 3000 | Pre-provisioned dashboards |
| `cadvisor` | `gcr.io/cadvisor/cadvisor:v0.51.0` | internal | Per-container CPU/memory/network |

Prometheus storage: `--storage.tsdb.retention.size=500MB`

### Monitoring Config Files

```
monitoring/
  prometheus/
    prometheus.yml             -- scrape jobs: scouter-backend, cadvisor
    rules/
      alerts.yml               -- alert rules (see below)
  grafana/
    provisioning/
      datasources/
        prometheus.yml         -- auto-configure Prometheus datasource
      dashboards/
        dashboard.yml          -- dashboard provider pointing to /dashboards/
    dashboards/
      scouter-application.json -- LLM + agents + HTTP + scheduler panels
      scouter-infrastructure.json -- cAdvisor container + Go runtime panels
```

### Prometheus Alert Rules

| Alert | Condition | Severity |
|-------|-----------|----------|
| `AllLLMModelsDown` | All circuit breakers open > 2m | critical |
| `HighLLMLatency` | p95 > 30s for 5m | warning |
| `HighLLMErrorRate` | failure rate > 30% for 5m | warning |
| `SchedulerNotRunning` | no run in 2h | warning |
| `BackendDown` | `up{job="scouter-backend"} == 0` for 1m | critical |

### Grafana — Application Dashboard (`scouter-application.json`)

Priority panels:
1. LLM call p50/p95/p99 latency by model
2. LLM fallback rate + error rate by model
3. Token consumption over time (input vs output)
4. Circuit breaker state table (closed/half-open/open per model)
5. Agent runs by type (success/failure ratio)
6. HTTP request rate + p95 latency by route
7. Scheduler runs + alerts triggered
8. Active missions gauge + unread notifications gauge

### Grafana — Infrastructure Dashboard (`scouter-infrastructure.json`)

Panels using cAdvisor:
- Per-container CPU/memory/network I/O
- Container restart count
- Go runtime: goroutines, GC pause duration, heap usage

### New Environment Variables

| Variable | Default | Notes |
|---|---|---|
| `METRICS_ENABLED` | `true` | Set `false` to disable (uses NoopRecorder) |
| `GF_SECURITY_ADMIN_PASSWORD` | `scouter` | Set in docker-compose (Grafana admin) |

### No DB Changes

### Risks

| Level | Risk | Mitigation |
|-------|------|-----------|
| HIGH | Cardinality explosion on HTTP paths | `chi.RouteContext().RoutePattern` bounds to ~15 values |
| MEDIUM | cAdvisor Docker socket on WSL2/Docker Desktop | Works on Linux/WSL2; document any host-specific flags |
| MEDIUM | Prometheus disk growth | `--storage.tsdb.retention.size=500MB` |
| LOW | Metric double-registration panic | Custom `prometheus.Registry` per instance, never global |

### Success Criteria

- [ ] `GET /metrics` returns Prometheus exposition format with all custom metrics + Go runtime
- [ ] LLM metrics include correct model/label/result labels
- [ ] HTTP middleware captures route patterns (not raw URLs)
- [ ] `docker compose --profile monitoring up` starts Prometheus + Grafana + cAdvisor
- [ ] Prometheus targets page shows `scouter-backend` and `cadvisor` as UP
- [ ] Grafana loads with pre-provisioned dashboards showing real data
- [ ] Alert rules visible in Prometheus alerts page
- [ ] `make test` passes with no regressions (all new Recorder injections default to NoopRecorder)
- [ ] `backend/internal/metrics/` achieves 80%+ test coverage

---

## Phase 44: Voice Input for Mission Creation

**Type**: Interface | **Priority**: Low | **Complexity**: Low
**Status**: ✅ Done (2026-03-15)

**User Value**: Faster mission creation — speak the mission name instead of typing, perfect for quick hands-free entry.

### Frontend — New Hook & Component

`src/hooks/useSpeechRecognition.ts`:
- Wraps Web Speech API (SpeechRecognition + webkitSpeechRecognition for Safari)
- State: `{ isListening, transcript, error, isSupported }`
- Methods: `startListening()`, `stopListening()`
- Configured: continuous=false, interimResults=true, lang='fr-FR'
- Returns partial transcripts while listening; normalizes on stop
- Graceful degradation when API unavailable

`src/components/mission/VoiceInputButton.tsx`:
- Microphone button next to mission name field
- Shows 🎤 (idle), ⏹ (recording)
- Pulsing red border when listening
- Appends transcript to name field (space-separated)
- Error display for permission denied, network errors, etc.
- Shows "browser not supported" message when unavailable
- Exported in `VoiceInputButton.module.css` with pulsing animation

### Modified Files

| File | Change |
|------|--------|
| `frontend/src/components/mission/MissionForm.tsx` | Import VoiceInputButton; wrap name input + button in flex row |
| `frontend/src/components/mission/MissionForm.module.css` | Add .nameRow and .nameFlex for layout |

### Test Coverage

- `useSpeechRecognition.test.ts`: 16 tests covering support detection, lifecycle (start/stop/end), transcript accumulation, error handling, graceful degradation
- `VoiceInputButton.test.tsx`: 13 tests covering render states (listening, idle, not supported), click handlers, transcript relay, error display

### Results

- All 29 new tests pass
- TypeScript strict mode passes
- Build succeeds (Vite + PWA)
- No breaking changes to existing features

---

## End of First Journey

After Phase 13, SCOUTER covers the complete lifecycle:
**Discover** → **Compare** → **Score & Decide** → **Track Prices** → **Get Alerts** → **Buy** → **Record Outcome** → **Search Past Research**

Phase 14 adds full operational observability across the entire stack.

Next journey: multi-user auth, cloud deployment, collaborative household/team missions.

---

## Auto-Improvement Loop — Phase 15+

> All planned phases complete. Entering continuous improvement. Tracked here for the user.

### Phase 15: Real-Time Price Intelligence via Public APIs ✅ COMPLETE
**Goal**: Integrate public price APIs to enrich research with live market data, focusing on the French market.
- **Aviationstack API**: Real-time flight prices for travel missions (flights, holidays)
- **Open Food Facts**: Food & grocery product comparisons
- **Kelkoo / Idealo affiliate feed**: French e-commerce price aggregation
- **Backmarket API**: Refurbished device prices (eco-friendly option)
- New backend `internal/priceapi/` package with pluggable adapter pattern
- Frontend: live price injection into option cards during research
- New "Live Prices" tab in OptionsExplorer

### Phase 16: Collaborative Missions ✅ COMPLETE
**Goal**: Allow household/team to share and collaborate on missions.
- Mission invite system with share links (extend existing share tokens)
- Real-time collaborative annotation on options
- Voting/thumbs on options, aggregate score visible to all participants
- "Household" context for budget pooling

### Phase 17: AI-Powered Negotiation Coach ✅ COMPLETE
**Goal**: After identifying the best product/price, coach the user on how to negotiate.
- New `NegotiationAgent` analyzing option attributes + market prices
- Produces structured negotiation script: opening offer, walk-away price, counter-offer script
- Frontend: "Coach Me" CTA on PurchaseForm after option selected
- Tracks negotiation outcomes in purchase_records (actual vs suggested)

### Phase 18: Progressive Web App + Mobile UX ✅ COMPLETE
**Goal**: Make Scouter usable offline and installable on mobile.
- PWA manifest + service worker with Workbox (offline shell + cached API responses)
- Push notifications for price alerts via Web Push API (replaces polling)
- Bottom navigation bar on mobile
- Swipe gestures on mission cards (archive left, research right)

### Phase 19: Multi-Currency & i18n Completion (Planned)
**Goal**: Full French/English UX with proper localization.
- All strings moved to i18n JSON keys (eliminate hardcoded English in components)
- Number/currency/date formatting respects locale setting from Settings
- French market: EUR default, date format DD/MM/YYYY
- RTL-ready CSS (for future language support)


---

## Phase 20: Interactive Analytics Dashboard

**Type**: Interface | **Priority**: High | **Complexity**: Medium
**Status**: 🔄 In Progress

**User Value**: Transform the Stats page from text/CSS bars into a rich, interactive analytics dashboard with recharts — spend trends over time, category donut chart, budget vs actual bar chart. Data storytelling for smarter future purchases.

### Backend
- `GET /api/stats/monthly` — monthly aggregation: `{ month: "2026-01", totalSpent, totalBudget, purchaseCount }`
- Query: `DATE_TRUNC('month', pr.purchased_at)` GROUP BY month ORDER BY month DESC LIMIT 12

### Frontend
- Install `recharts` package
- `src/api/stats.ts`: add `monthlyStatsSchema`, `fetchMonthlyStats`
- `src/hooks/useStats.ts`: add `useMonthlyStats`
- `src/components/charts/SpendTrendChart.tsx` + `.module.css`: recharts AreaChart, locale-aware Y-axis
- `src/components/charts/CategoryDonutChart.tsx` + `.module.css`: recharts PieChart with legend
- `src/components/charts/BudgetVsActualChart.tsx` + `.module.css`: recharts BarChart, grouped
- `src/pages/StatsPage.tsx`: integrate all 3 charts below existing summary cards

---

## Phase 21: Wish List & Price Drop Alerts

**Type**: Functional | **Priority**: High | **Complexity**: Medium
**Status**: 📋 Planned

**User Value**: Track products of interest before committing to a mission. Get alerted when prices drop. Perfect for French deal-hunting culture (soldes, promotions).

### Backend
- Migration `014_wish_list.up.sql`: `wish_list_items(id, user_session, name, url, target_price, last_seen_price, currency, created_at, alerted_at)`
- CRUD handler: `GET/POST/DELETE /api/wishlist`
- Scheduler job: hourly price check via HTTP HEAD/GET + price extraction heuristic, creates notification on drop

### Frontend
- `WishListPage` (`/wishlist` route): add item form, list with current vs target price
- `WishListItem` component: price delta badge, delete, edit target price
- `TopNav`: add Wishlist link

---

## Phase 22: Weighted Comparison Matrix

**Type**: Functional | **Priority**: Medium | **Complexity**: Medium
**Status**: 📋 Planned

**User Value**: Side-by-side structured comparison of shortlisted options with user-defined criteria weights. Export-ready matrix for confident final decisions.

### Backend
- `GET /api/missions/:id/comparison` — returns options with attributes, scores, constraints check
- `POST /api/missions/:id/comparison/weights` — persist weight config per mission (jsonb column on missions)

### Frontend
- `ComparisonMatrix` component: sticky header with option names, rows per attribute, weight sliders
- Visual winner highlighting (highest weighted score per row)
- Integrated into `OptionsExplorer` as a new "Compare" tab

---

## Phase 23: Smart Budget Forecaster

**Type**: Agent | **Priority**: Medium | **Complexity**: High
**Status**: 📋 Planned

**User Value**: AI-powered budget risk analysis — given mission category, budget, and historical spending patterns, predict overspend probability and suggest budget adjustments.

### Backend
- `internal/forecast/agent.go`: ForecastAgent using LLM tool-use; inputs: category stats, current budget, market price data
- Output: `{ riskScore: 0-100, predictedSpend: float, confidence: float, suggestions: string[] }`
- `POST /api/missions/:id/forecast`

### Frontend
- `ForecastPanel` component on MissionOverview: risk gauge, predicted spend range, AI suggestions
- `useForecast` hook with mutation + caching

---

## Phase 24: Price Alert Toggle + French Market UX ✅ COMPLETE

**Status**: ✅ Done
Added price alert toggle per shopping item, real-time notification on price drop, French locale formatting.

---

## Phase 25: EAN Barcode Lookup (Open Food Facts) ✅ COMPLETE

**Status**: ✅ Done
EAN barcode scanner using device camera → Open Food Facts API → auto-populate shopping item.

---

## Phase 26: Spending Persona — AI Buyer Archetype ✅ COMPLETE

**Status**: ✅ Done
PersonaAgent uses LLM tool-use to analyze purchase history and generate an archetype (e.g. "Budget Optimiser", "Impulse Buyer") with traits and tips. Displayed as PersonaCard on StatsPage.

### Backend
- `internal/persona/` — agent, repository, handler, model
- `GET /api/persona` (latest), `POST /api/persona` (run agent)
- Migration `018_persona.up.sql`

### Frontend
- `frontend/src/components/persona/PersonaCard.tsx`
- `usePersona` / `useRunPersona` hooks
- Wired into StatsPage

---

## Phase 27: Price History Chart ✅ COMPLETE

**Status**: ✅ Done
recharts LineChart per shopping item showing price evolution over time.

### Backend
- `GetPriceHistory(ctx, itemID, limit)` on `shopping.Repository`
- `GET /api/shopping/{itemID}/price-history?limit=N` endpoint
- `PriceHistoryPoint` model (`price`, `source`, `recordedAt`)

### Frontend
- `PriceHistoryChart` component (recharts AreaChart, gradient fill)
- `PriceHistoryModal` on ShoppingItemRow

---

## Phase 28: Smart Notifications Center

**Type**: Interface | **Priority**: High | **Complexity**: Medium
**Status**: ✅ Done (2026-03-15)

**User Value**: Replace the 60s-polling dropdown with a dedicated, actionable notifications page. Users can filter by mission, mark all read, delete, and click-through to the relevant item.

### Backend
- `DELETE /api/notifications/:id` — delete single notification
- `POST /api/notifications/mark-all-read` — bulk mark-read
- Filter query params on `GET /api/notifications`: `?missionId=&type=`

### Frontend
- `/notifications` full-page route: `NotificationsPage`
- Group notifications by mission (accordion sections)
- Filter bar: All / Price Alerts / Deal Score / System
- Bulk action toolbar: "Mark all read", "Clear dismissed"
- Per-notification delete button + navigation link to relevant mission
- `useNotifications` extended with delete/mark-all mutations
- `NotificationBell` dropdown shows "View all" → `/notifications`

---

## Phase 29: French Market Integrations

**Type**: Functional | **Priority**: High | **Complexity**: High
**Status**: ✅ Done (2026-03-15)

**User Value**: Real-time travel + product prices from French-market public APIs — Aviationstack for flights, SNCF (French rail) open data for trains, Google Shopping-style price comparison.

### Backend
- `internal/travel/` — TravelAgent: Aviationstack flight search, SNCF journey search
- Fetch cheapest fares for a route + date range → price history snapshots
- `GET /api/travel/flights?from=CDG&to=NCE&date=2026-04-01`
- `GET /api/travel/trains?from=Paris&to=Lyon&date=2026-04-01`
- Rate limiting + caching (1h TTL in memory)

### Frontend
- `TravelSearchWidget` in `MissionOverview` for travel-type missions
- Flight/train results shown as ShoppingItems with airline/operator badge
- i18n: French station/airport names

---

## Phase 30: Social Proof & Review Aggregation

**Type**: Functional | **Priority**: Medium | **Complexity**: High
**Status**: 🔄 In Progress (2026-03-15)

**User Value**: Pull community sentiment for a product from public sources — Trustpilot, Amazon FR reviews summary, Reddit mentions — to help users validate their shortlisted options with social proof.

### Backend
- `internal/reviews/` — ReviewAgent: LLM summarisation of public review excerpts
- `GET /api/options/:id/reviews` — fetch + cache review summary
- Sentiment score + top pro/con bullets

### Frontend
- `ReviewSummaryCard` on OptionCard detail view
- Sentiment gauge (positive/neutral/negative %)
- "Last refreshed" timestamp + manual refresh button

---

## Phase 31: Product Price Comparison (idealo FR / LeLynx)

**Type**: Functional | **Priority**: High | **Complexity**: High
**Status**: 📋 Planned

**User Value**: Compare live prices across French e-commerce retailers (Fnac, Cdiscount, Amazon FR, Darty) for any product in the user's mission options list — surfacing the best deal in one click.

### Backend
- `internal/pricecomp/` — PriceCompAgent: LLM-powered extraction of e-commerce price listings
- `GET /api/options/:id/price-comparison` — returns price comparison across retailers
- In-memory 30min TTL cache
- Scraping-safe: user-agent rotation, request throttling

### Frontend
- `PriceComparisonPanel` in OptionsExplorer detail view
- Retailer badge (logo/color coded), current price, availability indicator
- Sort by price ascending; highlight lowest price
- "Buy Now" external link button per retailer

---

## Phase 32: AI Purchase Timing Advisor

**Type**: Agent | **Priority**: Medium | **Complexity**: Medium
**Status**: 📋 Planned

**User Value**: "Should I buy now or wait?" — the LLM analyzes seasonal patterns, price history, product lifecycle, and upcoming sale events (Black Friday, French sales periods — Soldes d'été/d'hiver) to recommend optimal purchase timing.

### Backend
- `internal/timing/` — TimingAgent: analyzes price history + product category → timing recommendation
- `POST /api/missions/:id/timing-advice` — runs agent, returns JSON {recommendation, rationale, waitUntil, confidence}
- Cached 24h per mission

### Frontend
- `TimingAdvisorCard` on MissionOverview (below ForecastPanel)
- Recommendation badge: BUY_NOW / WAIT / UNSURE
- Rationale text + estimated wait period
- "Next sale event" countdown chip (Soldes calendar hardcoded for FR market)

---

## Phase 33: Mission Export to PDF

**Type**: Interface | **Priority**: Low | **Complexity**: Medium
**Status**: 📋 Planned

**User Value**: Download a full mission report as a shareable PDF — options table, budget analysis, decision recommendation, purchase history — perfect for submitting to employers for reimbursement or sharing with family.

### Backend
- Extend `internal/export/` — add PDF generation via `github.com/jung-kurt/gofpdf` or wkhtmltopdf
- `GET /api/missions/:id/export?format=pdf` (existing endpoint extended)
- Generate structured PDF: mission name, category, budget, options table, decision summary

### Frontend
- Add PDF download button to existing Export section in MissionOverview
- Show progress indicator during generation (server-side render)

---

## Phase 34: Smart Dark/Light Mode

**Type**: Interface | **Priority**: Low | **Complexity**: Low
**Status**: 📋 Planned

**User Value**: Respect system OS light/dark preference and allow manual override — persisted in localStorage. Improve readability and energy efficiency on OLED screens.

### Frontend
- `useTheme` hook: detects `prefers-color-scheme`, reads localStorage override
- `ThemeContext` + `ThemeProvider` in App root
- CSS variable swap: define `[data-theme="light"]` overrides for all `--` tokens in theme.css
- Theme toggle button in Settings page + Topnav quick toggle
- Smooth transition: `transition: background-color 0.2s, color 0.2s` on `:root`

---

## Phase 35: Envelope Budgeting Integration

**Type**: Functional | **Priority**: Medium | **Complexity**: Medium
**Status**: 📋 Planned

**User Value**: Apply the zero-based envelope budgeting method to missions — allocate your total monthly budget across multiple missions and see how each purchase affects your overall financial picture.

### Backend
- `internal/envelope/` — envelope model (id, name, monthly_amount, currency)
- `envelopes` table + migrations
- CRUD API: `GET/POST /api/envelopes`, `PATCH/DELETE /api/envelopes/:id`
- Link missions to envelopes: `mission.envelope_id` FK

### Frontend
- `EnvelopesPage` at `/envelopes`
- Visual envelope cards (fill gauge % of budget used vs allocated)
- Mission list within each envelope
- Budget health score across all envelopes (overbudget = red, healthy = cyan)
- Nav link in Layout sidebar
