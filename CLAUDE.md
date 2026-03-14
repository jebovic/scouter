# SCOUTER Universal — Claude Instructions

## ACTIVE ROADMAP → READ FIRST
**`ROADMAP.md`** is the source of truth for all remaining work.
- Phases 1–8 complete. Phase 9 next: Ollama Smart Routing (detailed spec in ROADMAP.md).
- Each phase session starts with: `/everything-claude-code:plan` + architect review
- Phase workflow:
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

## Core Rules
- Always read a file before editing it
- Never produce documentation unless explicitly asked
- Update this file continuously, keeping it minimal
- Prefer editing existing files over creating new ones
- Always append "2026" to web searches
- Responses: short, direct, no filler
- Immutability: always return new values, never mutate
- Before commit, ensure no sensitive data will be send to git

## Project Goal
Full-stack personal spending intelligence tool. Research, compare, and budget any major purchase.
Go backend + React frontend + PostgreSQL. ResearchAgent and PricingAgent call the LLM API via tool use for structured output.

## Tech Stack
- Backend: Go 1.23+, chi router, pgx/v5, golang-migrate, Anthropic SDK
- Frontend: React 19 + TypeScript, Vite, Tanstack Query v5, React Router v7, Zod
- Database: PostgreSQL 16 + pgvector (`vector(1024)` — Voyage AI v3 compatible, nullable until used)
- LLM: AnthropicProvider (claude-sonnet-4-6, tool-use for structured output) | OllamaProvider (phase 2)
- Deployment: Docker Compose (postgres, backend, frontend)

## Key Architecture Rules
- LLM Provider interface is **transport-only**: `Complete(ctx, CompletionRequest) (CompletionResponse, error)`
- ResearchAgent (`internal/research/`) and PricingAgent (`internal/pricing/`) own prompt construction, tool schema, response parsing
- All DB primary keys are UUIDs; missions also have a `slug TEXT UNIQUE NOT NULL` for URL routing
- Zod validates all API responses at the TypeScript boundary (in `src/api/`)
- golang-migrate for SQL migrations (up/down), auto-run at server startup

## File Structure
```
backend/
  cmd/server/main.go        -- entrypoint, env validation, chi router, migrate on start
  internal/
    config/                 -- Config struct, Load() from env, validation
    db/                     -- pgx pool init, golang-migrate runner (embeds migrations/)
    mission/                -- model, repository (pgx), service, handler
    option/                 -- model, repository, service, handler
    shopping/               -- model, repository, service, handler (items + price_history + deal-score endpoint)
    httputil/               -- WriteJSON / WriteError helpers (buffer-first, no header split)
    llm/                    -- Provider interface + CompletionRequest/Response + AnthropicProvider + OllamaProvider stub
    research/               -- ResearchAgent: tool schema, prompt builder, LLM call, option parser, DB persist
    pricing/                -- PricingAgent: tool schema, prompt builder, LLM call, price parser, DB persist
    dealintel/              -- Trend + DealScore calculation (pure Go, no DB deps)
    notification/           -- model, repository, handler (CRUD + mark-read + unread-count)
    scheduler/              -- robfig/cron background job (price-check alerts on active missions)
    template/               -- built-in template registry (15 templates, compiled in binary)
  migrations/               -- 001–006 up/down SQL files (golang-migrate format)
  Dockerfile

frontend/
  src/
    api/                    -- typed fetch wrappers + Zod schemas per resource
    components/
      scouter/              -- Card, Badge, BudgetBar, StatusBadge, Topnav, LoadingPulse, ScouterGrid
      mission/              -- MissionCard, MissionForm, ConstraintEditor, CategoryTemplate, TemplateCard, TemplateGallery, TemplatePreview
      options/              -- OptionCard, AttributeRenderer, ComparisonTable, ConstraintChecker, RadarChart
      shopping/             -- ShoppingList, MerchantGroup, ShoppingItemRow, PriceHistoryChart, CostBreakdown
    pages/                  -- HQDashboard, MissionOverview, OptionsExplorer, ShoppingTracker
    hooks/                  -- useMission, useOptions, useShopping, useResearch, usePriceIntel
    styles/
      theme.css             -- all SCOUTER CSS custom properties
      global.css            -- reset, base typography
    i18n/                   -- index.ts, en.json, fr.json
    types/                  -- mission.ts, option.ts, shopping.ts (mirrored from Zod schemas)
    main.tsx
  Dockerfile
```

## Environment Variables
See `.env.example` — required: `DATABASE_URL`. Phase 9 adds model pool vars.

| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | — | required |
| `ANTHROPIC_API_KEY` | — | required when `LLM_PROVIDER=anthropic` |
| `LLM_PROVIDER` | `ollama` | `anthropic` or `ollama` |
| `OLLAMA_BASE_URL` | `http://host.docker.internal:11434` | local Ollama |
| `OLLAMA_MODEL` | `qwen2.5:7b` | legacy alias → heavy model |
| `OLLAMA_HEAVY_MODEL` | `qwen3:14b` | Phase 9: primary tool-use model |
| `OLLAMA_FAST_MODEL` | `qwen3:4b` | Phase 9: lighter fallback |
| `OLLAMA_EMBED_MODEL` | `mxbai-embed-large` | Phase 11: 1024-dim embeddings |
| `OLLAMA_CLOUD_URL` | (empty) | Phase 9: `https://ollama.com` when enabled |
| `OLLAMA_CLOUD_MODEL` | (empty) | e.g. `deepseek-v3.2:cloud` |
| `OLLAMA_CLOUD_API_KEY` | (empty) | Bearer token from ollama.com |
| `PORT` | `8080` | backend listen port |
| `ENV` | `production` | `development` enables permissive CORS |

## Backend Status (Phases 1–8 complete)
- All handlers use `httputil.WriteJSON`/`WriteError` — no raw error leaks
- `errors.Is(err, pgx.ErrNoRows)` throughout all repositories
- `param.NewOpt(v)` used in Anthropic SDK; CORS gated on `ENV=development`; 1 MiB body cap
- Cursor-based pagination: `httputil.ParsePageParams` + `BuildPagedResponse[T]`, probe-row pattern
- Graceful shutdown timeout: 65s (exceeds 60s LLM call timeout)
- **Phase 7**: `internal/dealintel/` (trend+score, pure Go, TDD); `internal/notification/` (CRUD + mark-read); `internal/scheduler/` (robfig/cron, price-check alerts); `shopping_items.target_price`; `notifications` table (migration 006)
- **Phase 8**: `internal/template/` (registry, 15 built-in templates, compiled in binary); `GET /api/templates`, `GET /api/templates/:id`

## Frontend Status (Phases 1–8 complete)
- **CSS Modules**: all components use co-located `.module.css` files; no raw `style={{}}` for layout/theming
- **Responsive**: breakpoints at 640px and 1024px across all pages and components
- **Skeleton loading**: `Skeleton` (card/row/chart variants) + `SkeletonGrid` via `ScouterGrid`
- **Empty states**: `EmptyState` with icon, title, description, CTA action across all pages
- **Onboarding**: `useOnboarding` + `OnboardingOverlay` (3-step, localStorage dismissed, wired in App)
- **Keyboard shortcuts**: `useKeyboardShortcuts` hook (ref-stable, preventDefault); `N` new mission, `R` research, `P` pricing
- **Sidebar**: collapsible mission list drawer, wired in App via `SidebarContext` (`src/contexts/sidebar.tsx`)
- **Breadcrumb**: `Breadcrumb` component with `missionSlug` + `missionName` + `subPage` props
- **Templates (Phase 8)**: `TemplateCard`, `TemplateGallery`, `TemplatePreview` (accessible modal); `useTemplates` (24h stale); `MissionForm` `initialValues` prop; `HQDashboard` wired end-to-end
- **Layout migration (Phase 7)**: React Router v7 Outlet pattern — `Layout.tsx` (root shell: sidebar+onboarding), `MissionLayout.tsx` (page wrapper+Topnav); mission pages now return just `<main>` content
- **Deal intel (Phase 7)**: `TrendBadge`, `DealScoreBadge`, `PriceSparkline` in `ShoppingItemRow`; `NotificationBell` in `Topnav`; `useNotifications` (60s poll); `getDealScore` API; `target_price` inline edit
- **Test suite**: Vitest + jsdom + Testing Library; 71 tests across 10 files

## CSS Conventions
- Import `frontend/src/styles/theme.css` for all SCOUTER tokens
- Card-based layout, 16px border-radius, `var(--surface)` background
- Status badges: `buy`(green), `flash-sale`(orange+pulse), `preorder`(gold), `defer`(text-dim), `watch`(purple), `crisis`(coral), `recommended`(cyan), `rejected`(coral-dim)
- Dynamic values use CSS custom properties: `style={{ '--var': value } as React.CSSProperties}` consumed as `var(--var)` in module CSS
