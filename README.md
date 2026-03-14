# SCOUTER Universal

Personal spending intelligence — research, compare, and budget any major purchase.

## What it does

1. **Create a Mission** — name a purchase goal, set a budget and constraints
2. **Run Research** — ResearchAgent calls the LLM via tool use to discover and rank options
3. **Run Pricing** — PricingAgent hunts prices across merchants, calculates TCO, flags deals
4. **Track spending** — Shopping tracker with price history and budget burn rate

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.23+, chi, pgx/v5, golang-migrate |
| Frontend | React 19, TypeScript, Vite, Tanstack Query v5, React Router v7, Zod |
| Database | PostgreSQL 16 + pgvector |
| LLM | Anthropic claude-sonnet-4-6 (tool use) · Ollama (text-only, no tools) |
| Deploy | Docker Compose |

## Quick start

```bash
cp .env.example .env          # fill in ANTHROPIC_API_KEY
make dev                      # builds and starts postgres + backend + frontend
```

- Backend API: `http://localhost:8080`
- Frontend: `http://localhost:5173`
- Health: `http://localhost:8080/api/health`

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | yes | — | PostgreSQL connection string |
| `ANTHROPIC_API_KEY` | yes* | — | *Required when `LLM_PROVIDER=anthropic` |
| `LLM_PROVIDER` | no | `anthropic` | `anthropic` or `ollama` |
| `OLLAMA_BASE_URL` | no | — | e.g. `http://localhost:11434` |
| `OLLAMA_MODEL` | no | — | e.g. `llama3.2` |
| `PORT` | no | `8080` | Backend listen port |
| `ENV` | no | `production` | `development` enables permissive CORS |

## API overview

```
GET    /api/health
GET    /api/missions
POST   /api/missions
GET    /api/missions/{slug}
PUT    /api/missions/{slug}
DELETE /api/missions/{slug}

POST   /api/missions/{id}/research          # trigger ResearchAgent
POST   /api/missions/{id}/pricing           # trigger PricingAgent

GET    /api/missions/{id}/options
POST   /api/missions/{id}/options
GET    /api/missions/{id}/options/{optionID}
PUT    /api/missions/{id}/options/{optionID}
DELETE /api/missions/{id}/options/{optionID}

GET    /api/missions/{id}/shopping
POST   /api/missions/{id}/shopping
GET    /api/missions/{id}/shopping/{itemID}
PUT    /api/missions/{id}/shopping/{itemID}
DELETE /api/missions/{id}/shopping/{itemID}
POST   /api/missions/{id}/shopping/{itemID}/snapshots
GET    /api/missions/{id}/shopping/{itemID}/snapshots
```

Note: mission CRUD uses `{slug}` (human-readable), sub-resources use `{id}` (UUID from the GET response).

## Make targets

```bash
make dev          # docker compose up --build
make dev-build    # rebuild from scratch
make test         # go test ./...
make lint         # go vet ./...
make seed         # POST a sample mission
make migrate-down # roll back last migration
make clean        # docker compose down -v
```

## Project structure

```
backend/
  cmd/server/main.go        entrypoint
  internal/
    config/                 env loading + validation
    db/                     pgx pool + golang-migrate runner
    httputil/               shared WriteJSON / WriteError helpers
    mission/                model, repo, service, handler
    option/                 model, repo, service, handler
    shopping/               model, repo, service, handler
    llm/                    Provider interface + Anthropic + Ollama
    research/               ResearchAgent
    pricing/                PricingAgent
  migrations/

frontend/
  src/
    api/                    typed fetch + Zod schemas
    components/
      scouter/              design system (Card, Badge, BudgetBar…)
      mission/              MissionCard, MissionForm…
      options/              OptionCard, ComparisonTable…
      shopping/             ShoppingList, MerchantGroup…
    pages/                  HQDashboard, MissionOverview, OptionsExplorer, ShoppingTracker
    hooks/                  useMission, useOptions, useShopping, useResearch, usePriceIntel
    styles/                 theme.css (tokens), global.css
    i18n/                   en.json, fr.json
    types/                  TypeScript types mirrored from Zod schemas
```
