# SCOUTER — Release Roadmap

> **Active plan for the first journey.** Read this at the start of every session.
> Current status: Phases 1–14 complete. All planned phases delivered. Auto-improvement loop active.

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

**Execution order:** 3+4 in parallel → 5+6 in parallel → 7 → 8 → **9** → 10+11 in parallel → 12+13 in parallel → **14**

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

## End of First Journey

After Phase 13, SCOUTER covers the complete lifecycle:
**Discover** → **Compare** → **Score & Decide** → **Track Prices** → **Get Alerts** → **Buy** → **Record Outcome** → **Search Past Research**

Phase 14 adds full operational observability across the entire stack.

Next journey: multi-user auth, cloud deployment, collaborative household/team missions.

---

## Auto-Improvement Loop — Phase 15+

> All planned phases complete. Entering continuous improvement. Tracked here for the user.

### Phase 15: Real-Time Price Intelligence via Public APIs (Planned)
**Goal**: Integrate public price APIs to enrich research with live market data, focusing on the French market.
- **Aviationstack API**: Real-time flight prices for travel missions (flights, holidays)
- **Open Food Facts**: Food & grocery product comparisons
- **Kelkoo / Idealo affiliate feed**: French e-commerce price aggregation
- **Backmarket API**: Refurbished device prices (eco-friendly option)
- New backend `internal/priceapi/` package with pluggable adapter pattern
- Frontend: live price injection into option cards during research
- New "Live Prices" tab in OptionsExplorer

### Phase 16: Collaborative Missions (Planned)
**Goal**: Allow household/team to share and collaborate on missions.
- Mission invite system with share links (extend existing share tokens)
- Real-time collaborative annotation on options
- Voting/thumbs on options, aggregate score visible to all participants
- "Household" context for budget pooling

### Phase 17: AI-Powered Negotiation Coach (Planned)
**Goal**: After identifying the best product/price, coach the user on how to negotiate.
- New `NegotiationAgent` analyzing option attributes + market prices
- Produces structured negotiation script: opening offer, walk-away price, counter-offer script
- Frontend: "Coach Me" CTA on PurchaseForm after option selected
- Tracks negotiation outcomes in purchase_records (actual vs suggested)

### Phase 18: Progressive Web App + Mobile UX (Planned)
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

