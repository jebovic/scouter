# SCOUTER Universal — Claude Instructions

## ACTIVE ROADMAP → READ FIRST
**`ROADMAP.md`** is the source of truth for all remaining work.
- Phases 1–4 complete. Phase 5 planned (detailed spec in ROADMAP.md) — ready to implement.
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
  9. fix → close phase

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
    shopping/               -- model, repository, service, handler (items + price_history)
    httputil/               -- WriteJSON / WriteError helpers (buffer-first, no header split)
    llm/                    -- Provider interface + CompletionRequest/Response + AnthropicProvider + OllamaProvider stub
    research/               -- ResearchAgent: tool schema, prompt builder, LLM call, option parser, DB persist
    pricing/                -- PricingAgent: tool schema, prompt builder, LLM call, price parser, DB persist
  migrations/               -- 001_init.sql / 001_init.down.sql (golang-migrate format)
  Dockerfile

frontend/
  src/
    api/                    -- typed fetch wrappers + Zod schemas per resource
    components/
      scouter/              -- Card, Badge, BudgetBar, StatusBadge, Topnav, LoadingPulse, ScouterGrid
      mission/              -- MissionCard, MissionForm, ConstraintEditor, CategoryTemplate
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
See `.env.example` — required: `DATABASE_URL`, `ANTHROPIC_API_KEY`

| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | — | required |
| `ANTHROPIC_API_KEY` | — | required when `LLM_PROVIDER=anthropic` |
| `LLM_PROVIDER` | `anthropic` | `anthropic` or `ollama` |
| `OLLAMA_BASE_URL` | — | required when `LLM_PROVIDER=ollama` |
| `OLLAMA_MODEL` | — | required when `LLM_PROVIDER=ollama` |
| `PORT` | `8080` | backend listen port |
| `ENV` | `production` | `development` enables permissive CORS |

## Backend Status
Backend is **fully implemented** (Phases 1–4 complete, all go-reviewer issues resolved):
- All handlers use `httputil.WriteJSON`/`WriteError` — no raw error leaks
- `errors.Is(err, pgx.ErrNoRows)` throughout all repositories
- JSON marshal/unmarshal errors propagated everywhere
- `param.NewOpt(v)` used in Anthropic SDK (not struct literal)
- CORS gated on `ENV=development`; 1 MiB request body cap in chi middleware
- Cursor-based pagination on all list endpoints (`httputil.ParsePageParams` + `BuildPagedResponse[T]`, probe-row pattern)
- Service-layer input validation in `mission.Service.Create` (defense-in-depth)
- Graceful shutdown timeout: 65s (exceeds 60s LLM call timeout)

## CSS Conventions
- Import `frontend/src/styles/theme.css` for all SCOUTER tokens
- Card-based layout, 16px border-radius, `var(--surface)` background
- Status badges: `buy`(green), `flash-sale`(orange+pulse), `preorder`(gold), `defer`(text-dim), `watch`(purple), `crisis`(coral), `recommended`(cyan), `rejected`(coral-dim)
