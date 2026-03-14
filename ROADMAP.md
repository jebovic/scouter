# SCOUTER — Release Roadmap

> **Active plan for the first journey.** Read this at the start of every session.
> Current status: Phases 1–8 complete. Phase 9 next.

## Phase Implementation Workflow (repeat for every phase)

```
1. /everything-claude-code:plan + architect  →  detailed plan for the phase
2. follow ECC tdd workflow to implerment the phase with specialized agents (go coding for backend, frontend agent with /frontend-design and /frontend-patterns skills for frontend), parallelize where possible
3. make test  (backend Go tests)
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
| 9 | Export, Share & Archive | Functional | Medium | ⬜ |
| 10 | Semantic Search (pgvector) | Agent + Infra | Medium | ⬜ |
| 11 | Mission Lifecycle & Post-Purchase | Functional | Medium | ⬜ |
| 12 | Settings, Data Management & Deployment | Infrastructure | Low | ⬜ |

**Execution order:** 3+4 in parallel → 5+6 in parallel → 7 → 8+9 in parallel → 10+11 in parallel → 12

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

## Phase 9: Export, Share & Archive

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

## Phase 10: Semantic Search & Smart Suggestions (pgvector)

**Type**: Agent + Infrastructure | **Priority**: Medium | **Complexity**: High
**Depends on**: Phase 5

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

## Phase 11: Mission Lifecycle & Post-Purchase Tracking

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

## Phase 12: Settings, Data Management & Deployment Polish

**Type**: Infrastructure | **Priority**: Low | **Complexity**: Medium
**Depends on**: Phase 9

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

## End of First Journey

After Phase 12, SCOUTER covers the complete lifecycle:
**Discover** → **Compare** → **Score & Decide** → **Track Prices** → **Get Alerts** → **Buy** → **Record Outcome** → **Search Past Research**

Next journey: multi-user auth, cloud deployment, collaborative household/team missions.
