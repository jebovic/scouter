# SCOUTER — Release Roadmap

> **Active plan.** Read this at the start of every session.
> Current status: **v0.1.0 released** (Phases 1–172 complete). Starting v0.2.0.

## Phase Implementation Workflow (repeat for every phase)

```
1. /everything-claude-code:plan + architect  →  detailed plan for the phase
2. follow ECC tdd workflow to implement the phase with specialized agents
   (go coding for backend, frontend agent with /frontend-design and /frontend-patterns for frontend)
   parallelize where possible
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

## v0.1.0 — SCOUTER Universal (Released 2026-03-15)

**172 phases shipped. Full lifecycle from research to purchase scorecard.**

### Milestone Delivery Summary

| Area | What shipped |
|------|-------------|
| **Core Agents** | ResearchAgent + PricingAgent (tool-use), DecisionAgent (ScoreEngine + LLM rationale), NegotiationCoach, SpendingPersona, PurchaseTimingAdvisor |
| **LLM Routing** | SmartRouter (heavy→fast→cloud→Anthropic priority pool, per-model circuit breakers + rate limiters, context-based capability hints) |
| **Price Intelligence** | Deal-score algorithm, price alerts, French market benchmark, watchlist, barcode lookup, seasonal calendar |
| **Mission Lifecycle** | research → options → shopping → purchase → scorecard; export (JSON/PDF/MD), share token, archive, duplicate |
| **Collaboration** | Invite + voting, wishlist sharing, collaborative annotations |
| **Budget** | Envelope budgeting (localStorage), multi-mission rollup, VAT calculator, spend velocity tracker |
| **Infrastructure** | PostgreSQL 16 + pgvector (IVFFlat, 1024-dim), golang-migrate, cursor pagination, graceful shutdown (65s), Prometheus + Grafana + cAdvisor |
| **Frontend** | React 19 + TypeScript + Vite, PWA (offline/installable), dark/light theme, i18n (EN/FR), semantic search dropdown, recharts analytics, gamification badges, keyboard shortcuts |

### Key Architectural Decisions (load-bearing for v0.2.0)

**LLM Provider interface is transport-only.**
`Complete(ctx, CompletionRequest) (CompletionResponse, error)` — routing hints are carried via `context.Context` (`RequestOpts`, `WithRequestOpts`), never on `CompletionRequest`. This keeps the interface stable while allowing SmartRouter to route by capability.

**JSON fallback belongs to agents, not the router.**
`RetryAsJSON` is a shared helper in `internal/llm/`, but agents own their tool schemas and retry logic. The router only cascades on infra errors (circuit open, timeout).

**ScoreEngine is pure Go; AI adds rationale only.**
`internal/decision/` separates deterministic scoring (pure Go, testable) from LLM explanation. Same pattern applies to all compute modules (dealintel, scorecard, etc.).

**Compute modules: FNV-32a + in-memory cache, no DB migration.**
Phases 90–172 compute modules follow: FNV-32a seed for deterministic pseudo-random → in-memory TTL cache (10–60min) → zero DB migration. New compute features should follow this pattern unless real persistence is needed.

**pgvector: async worker, nullable until populated.**
`embedding vector(1024)` is nullable. `internal/embedding/` runs a 2-goroutine async worker channel. Options are embedded after creation; search degrades gracefully if embeddings are missing. IVFFlat index with `lists=10`.

**Metrics: domain-scoped sub-interfaces + NoopRecorder default.**
`internal/metrics/` defines `LLMRecorder`, `AgentRecorder`, `SchedulerRecorder`, `HTTPRecorder`. All constructors accept recorder interfaces; `NoopRecorder` is the default — no nil checks needed anywhere. HTTP path labels use `chi.RouteContext().RoutePattern` for bounded cardinality.

**Routes extracted to `cmd/server/routes.go`.**
`routeDeps` struct + `registerRoutes` function. `main.go` wires env/config/DB/providers; `routes.go` owns all HTTP route registration. Keep this separation for v0.2.0.

**Cursor-based pagination with probe-row pattern.**
`httputil.ParsePageParams` + `BuildPagedResponse[T]`. All list endpoints use this — do not add offset pagination.

---

## v0.2.0 — Next Milestone

> Phases planned here. Fill in as the milestone takes shape.

### Candidate Themes

- **Multi-user auth** — JWT/session auth, user accounts, household collaboration at the auth layer
- **Cloud deployment** — production-ready Kubernetes/Fly.io/Railway config, secrets management, CI/CD pipeline
- **Real external data** — live price feeds from actual French retailer APIs (Fnac, Cdiscount, Amazon FR)
- **Mobile-native UX** — Capacitor wrapper or React Native bridge for iOS/Android app store distribution

### Phase Status

| Phase | Name | Type | Priority | Status |
|-------|------|------|----------|--------|
| 173 | Mission & Option Management | Frontend | High | ✅ Done |
| 174 | _(next phase)_ | — | — | 📋 Planned |

---

## Phase 173 — Mission & Option Management (2026-03-17)

**Goal:** Close three friction points blocking effective mission management.

**Delivered:**
- **MissionActionBar** — edit/archive/delete actions in MissionOverview header; `MissionEditModal` pre-fills current values, calls `PATCH /api/missions/{slug}` on save
- **MissionCard ⋯ menu** — archive/delete overflow menu on HQ Dashboard cards; archived badge variant
- **HQ Dashboard "Show archived" toggle** — `include_archived=true` query, archived missions shown with badge and Unarchive option
- **OptionEditModal** — edit name, badge, price range, notes, warnings via `PUT /api/missions/{missionId}/options/{optionId}`
- **OptionCard ⋯ menu** — edit/delete overflow menu alongside existing pin/reject buttons
- **"Unpin All"** — renamed from "Clear Pinned"; now parallel-unpins all pinned options (keeps them in explorer); partial-failure toast if some fail
- **ShortlistPanel** — buying-phase panel surfacing pinned options as compact cards with Select button; pre-fills purchase form (name + price midpoint); replace-confirm dialog if purchase record already exists; empty-state prompt with link to Options Explorer
- **i18n** — all new keys in `en.json` and `fr.json`
- **E2E tests** — 11 tests in `frontend/e2e/mission-management.spec.ts` covering the full management flow
