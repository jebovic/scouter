# SCOUTER Universal — Codemaps

Quick reference guides to understand the codebase architecture and navigate key modules.

**Last Updated:** 2026-03-15 · **Version:** 0.1.0

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Backend Architecture](#backend-architecture)
3. [Frontend Architecture](#frontend-architecture)
4. [Data Flow](#data-flow)
5. [Key Modules](#key-modules)
6. [Directory Structure](#directory-structure)

---

## Project Overview

SCOUTER is a full-stack personal spending intelligence tool built with:

- **Backend:** Go 1.23 (132 packages, 170+ API endpoints)
- **Frontend:** React 19 (22 pages, 71+ tests)
- **Database:** PostgreSQL 16 + pgvector
- **LLM:** Anthropic + Ollama with SmartRouter
- **Deployment:** Docker Compose + Prometheus/Grafana

### Key Statistics

```
Backend:     132 Go packages
Frontend:    22 pages, 10+ component groups
Database:    22+ migrations, 1024-dim embeddings
LLM:         3 providers, circuit breaker protection
Tests:       999 unit tests (backend), 31 E2E tests (frontend)
API:         170+ endpoints (CRUD, agents, search, analytics)
Deployment:  Docker Compose (postgres, backend, frontend, monitoring)
```

---

## Backend Architecture

### Entry Point

```
cmd/server/main.go
├── Load config from environment
├── Initialize PostgreSQL connection pool
├── Run migrations automatically
├── Initialize LLM provider (SmartRouter)
├── Register routes via routeDeps struct
└── Start HTTP server on :8080

cmd/server/routes.go
├── routeDeps struct (all dependencies)
└── registerRoutes() function (chi router setup)
```

### Core Layers

```
HTTP Request
    ↓
[Handler] — chi middleware (CORS, logging)
    ↓
[Service] — business logic
    ↓
[Repository] — data access (pgx)
    ↓
PostgreSQL
```

### Package Organization

**Domain Packages** (each has model.go, repository.go, service.go, handler.go):

| Package | Responsibility | Key Entities |
|---------|-----------------|--------------|
| `mission` | Purchase goals | Mission (id, slug, budget, status) |
| `option` | Research results | Option (product, score, constraints) |
| `shopping` | Price tracking | ShoppingItem (merchant, price_history) |
| `notification` | Alerts | Notification (type, read_at) |
| `purchase` | Purchase records | PurchaseRecord (price, date, lessons) |

**Specialized Packages**:

| Package | Purpose | Entry Point |
|---------|---------|-------------|
| `research` | ResearchAgent (LLM discovery) | POST /api/missions/:id/research |
| `pricing` | PricingAgent (price hunting) | POST /api/missions/:id/pricing |
| `llm` | Provider interface + SmartRouter | internal/llm/smartrouter.go |
| `search` | pgvector semantic search | GET /api/search?q=... |
| `scorecard` | Efficiency grading (A/B/C/D) | GET /api/missions/:id/scorecard |
| `embed` | Embedding generation | Internal, async worker |
| `scheduler` | Background jobs (cron) | Price-check alerts |
| `template` | Built-in mission templates | GET /api/templates |

**Infrastructure Packages**:

| Package | Purpose |
|---------|---------|
| `config` | Environment validation, Config struct |
| `db` | PostgreSQL pool, golang-migrate runner |
| `httputil` | WriteJSON, WriteError helpers |
| `metrics` | Prometheus recorder interface |

### Request Flow Example: Research

```
POST /api/missions/{id}/research
    ↓
[Handler.Research]
    ↓
[Service.RunResearch]
    ├── Call LLM via SmartRouter
    ├── Parse tool-use response
    ├── Extract options from response
    ├── Validate against constraints
    └── Persist to database
    ↓
[Repository.CreateOptions] (batch)
    ↓
PostgreSQL: INSERT INTO options
    ↓
HTTP 200: { options: [...], count: 5 }
```

### LLM Integration (SmartRouter)

```
SmartRouter (internal/llm/smartrouter.go)
├── Provider Pool
│   ├── ollama-heavy (qwen3:14b, primary)
│   ├── ollama-fast (qwen3:4b, fallback)
│   ├── ollama-cloud (optional, cloud endpoint)
│   └── anthropic (claude-sonnet-4-6, final fallback)
│
├── Routing Logic
│   ├── Check capability requirements
│   ├── Try heavy model first
│   ├── Fall back to fast on timeout/error
│   ├── Fall back to cloud if available
│   └── Fall back to Anthropic (100% reliable)
│
└── Circuit Breakers
    └── Per-provider: detect failures, auto-recover
```

---

## Frontend Architecture

### Route Structure

```
<App>
├── <Layout>
│   ├── <Sidebar>          # Mission list drawer
│   ├── <Topnav>           # Logo, search, notifications, settings
│   │
│   └── <HQDashboard>      # /
│       ├── MissionCard[]
│       └── TemplateGallery
│
├── <MissionLayout>        # /missions/:slug/*
│   ├── <Topnav>
│   ├── <Breadcrumb>
│   └── <Outlet>
│       ├── <MissionOverview>      # /missions/:slug
│       ├── <OptionsExplorer>      # /missions/:slug/options
│       ├── <ShoppingTracker>      # /missions/:slug/shopping
│       └── ...
│
├── <SearchPage>           # /search?q=...
├── <HistoryPage>          # /history
├── <StatsPage>            # /stats
└── <SettingsPage>         # /settings
```

### Component Organization

```
src/components/
├── scouter/               # Design system
│   ├── Card              # Base card component
│   ├── Badge             # Status badges
│   ├── BudgetBar         # Progress bar
│   ├── StatusBadge       # buy/flash-sale/preorder/etc
│   ├── Topnav
│   ├── Sidebar
│   ├── LoadingPulse
│   └── ScouterGrid       # Layout grid
│
├── mission/               # Mission domain
│   ├── MissionCard
│   ├── MissionForm
│   ├── ConstraintEditor
│   ├── CategoryTemplate
│   └── TemplateGallery
│
├── options/               # Option domain
│   ├── OptionCard
│   ├── AttributeRenderer
│   ├── ComparisonTable
│   ├── ConstraintChecker
│   └── RadarChart
│
└── shopping/              # Shopping domain
    ├── ShoppingList
    ├── MerchantGroup
    ├── ShoppingItemRow
    ├── PriceHistoryChart
    └── CostBreakdown
```

### State Management

**TanStack Query v5** (server state):

```
src/hooks/
├── useMission(slug)              → useQuery + useMutation
├── useOptions(missionId)         → useQuery
├── useShopping(missionId)        → useQuery
├── useSearch(query)              → useQuery (300ms debounce)
└── useResearch(missionId)        → useMutation (triggers ResearchAgent)
```

**Context API** (client state):

```
src/contexts/
├── sidebar.tsx                   → Sidebar open/close state
├── onboarding.tsx               → 3-step tutorial state
└── App.tsx global state
```

**Local Storage**:

```
localStorage keys:
├── onboarding-dismissed        → Boolean
├── currency                    → String (EUR, USD, etc)
└── locale                      → String (en-US, fr-FR, etc)
```

### API Client Layer

```
src/api/
├── client.ts                    # apiFetch wrapper, Zod parsing
├── mission.ts                   # Mission schemas + fetchMission()
├── option.ts                    # Option schemas + fetchOptions()
├── shopping.ts                  # Shopping schemas
├── search.ts                    # Search schemas
├── stats.ts                     # Stats schemas
└── ...                          # One file per resource
```

**Pattern:**

```ts
// src/api/mission.ts
export const MissionSchema = z.object({ /* ... */ });
export type Mission = z.infer<typeof MissionSchema>;

export async function getMission(slug: string): Promise<Mission> {
  return apiFetch(`/api/missions/${slug}`, MissionSchema);
}

// src/hooks/useMission.ts
export function useMission(slug: string) {
  return useQuery({
    queryKey: ['mission', slug],
    queryFn: () => getMission(slug),
  });
}
```

### Styling

**CSS Modules** (no raw `style={{}}` for layout):

```
src/components/
├── Card.tsx
├── Card.module.css         # .card { background: var(--surface); }

src/styles/
├── theme.css              # CSS custom properties (--primary, --surface, etc)
└── global.css             # Reset, base typography
```

**Design System Tokens:**

```css
/* Color palette */
--primary: #3b82f6;           /* Blue */
--success: #10b981;           /* Green */
--warning: #f59e0b;           /* Amber */
--danger: #ef4444;            /* Red */
--surface: #ffffff;           /* Background */
--surface-dark: #f3f4f6;      /* Secondary bg */
--text-primary: #1f2937;      /* Dark text */
--text-secondary: #6b7280;    /* Gray text */
```

---

## Data Flow

### Mission Creation → Research → Purchase

```
1. User creates mission
   Frontend: MissionForm → POST /api/missions
   ↓ Backend: Handler → Service → Repository → INSERT

2. Database: missions table populated
   ↓

3. User clicks "Run Research"
   Frontend: useMutation('research') → POST /api/missions/:id/research
   ↓ Backend: ResearchAgent

4. ResearchAgent flow
   ├── Construct prompt with mission + constraints
   ├── Call SmartRouter.Complete() → LLM
   ├── LLM returns tool-use response with options
   ├── Parse options from tool response
   ├── Validate options against constraints
   └── Persist options to database

5. Database: options table populated
   Frontend: useQuery refreshes, shows OptionCard[]
   ↓

6. User clicks "Run Pricing"
   Frontend: useMutation('pricing') → POST /api/missions/:id/pricing
   ↓ Backend: PricingAgent

7. PricingAgent flow
   ├── For each option:
   │   ├── Call LLM to find merchants
   │   ├── Parse merchant URLs from tool response
   │   ├── Scrape/estimate prices
   │   ├── Calculate TCO (price + shipping + tax)
   │   └── Score deal quality
   └── Persist to shopping_items + price_history

8. Database: shopping_items + price_history populated
   Frontend: ShoppingTracker shows PriceHistoryChart[]
   ↓

9. User records purchase
   Frontend: POST /api/missions/:id/purchase
   ↓ Backend: PurchaseService

10. Database: purchase_records table, mission status → "done"
    Frontend: Mission appears in /history
    Scorecard endpoint enabled
```

### Semantic Search Flow

```
1. User types "laptop" in search box
   Frontend: useSearch('laptop', { debounce: 300ms })
   ↓

2. GET /api/search?q=laptop
   ↓ Backend: SearchHandler

3. EmbeddingWorker flow
   ├── Async worker reads pending embeddings from channel
   ├── Call Ollama /api/embed endpoint
   ├── Store in options.embedding (pgvector)
   └── Mark as embedded

4. SearchHandler: Cosine similarity search
   ├── Generate embedding for "laptop"
   ├── SELECT ... FROM options WHERE embedding <=> $1 < 0.5
   ├── ORDER BY similarity DESC
   └── Return top 10 results

5. Frontend: Display SearchDropdown with results
   User presses Enter → Navigate to /search?q=laptop
```

---

## Key Modules Deep Dive

### Research Agent (`internal/research/`)

**Files:**
- `model.go` — ResearchRequest, Option
- `agent.go` — ResearchAgent, Run() method
- `toolschema.go` — LLM tool definition
- `prompt.go` — Prompt construction from mission
- `parser.go` — Parse LLM response

**Flow:**

```go
func (a *ResearchAgent) Run(ctx context.Context, mission Mission) error {
  // 1. Build prompt
  prompt := a.buildPrompt(mission)

  // 2. Call LLM with tool schema
  response, err := a.llm.Complete(ctx, CompletionRequest{
    Messages: []Message{{ Role: "user", Content: prompt }},
    Tools: []Tool{{ Name: "discover_options", ... }},
  })

  // 3. Parse tool response
  options := a.parseOptions(response.ToolUse)

  // 4. Validate against constraints
  for _, opt := range options {
    if !opt.Satisfies(mission.Constraints) {
      continue  // skip
    }
    // 5. Persist to database
    repo.CreateOption(ctx, opt)
  }
}
```

### Pricing Agent (`internal/pricing/`)

Similar structure to research agent. Discovers merchants, hunts prices, calculates TCO.

**Tool Use Schema:**
```json
{
  "name": "find_prices",
  "description": "Search for prices of a product across merchants",
  "input_schema": {
    "type": "object",
    "properties": {
      "product": { "type": "string" },
      "merchants": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

### SmartRouter (`internal/llm/smartrouter.go`)

**Core method:**

```go
func (r *SmartRouter) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
  // 1. Extract routing hints from context
  opts := GetRequestOpts(ctx)

  // 2. Get candidate providers based on required capabilities
  candidates := r.pool.ForCapabilities(opts.RequiredCapabilities)

  // 3. Try providers in order (heavy → fast → cloud → anthropic)
  for _, provider := range candidates {
    resp, err := provider.Complete(ctx, req)
    if err == nil {
      return resp, nil
    }
    if errors.Is(err, context.DeadlineExceeded) {
      continue  // timeout, try next
    }
  }

  // 4. All providers failed
  return nil, fmt.Errorf("all LLM providers exhausted")
}
```

### Search Index (`internal/search/`)

Uses pgvector for similarity search:

```sql
-- CREATE INDEX ON options USING IVFFlat (embedding vector_cosine_ops)
-- WITH (lists = 10)

-- Query similar options
SELECT id, name, embedding <=> $1 AS distance
FROM options
WHERE embedding IS NOT NULL
ORDER BY embedding <=> $1 LIMIT 10;
```

---

## Directory Structure

### Root Level

```
scouter/
├── README.md                    # This file + quick start
├── ROADMAP.md                   # Phase-based development plan
├── CLAUDE.md                    # Project-specific dev instructions
├── Makefile                     # dev, test, seed, migrate, clean
├── docker-compose.yml           # Services definition
├── .env.example                 # Environment template
├── .gitignore
│
├── backend/                     # Go backend
├── frontend/                    # React frontend
├── docs/                        # Documentation
├── monitoring/                  # Prometheus + Grafana
└── archives/                    # Old phases, reference material
```

### Backend Deep Dive

```
backend/
├── cmd/
│   ├── server/
│   │   ├── main.go              # Entrypoint
│   │   ├── routes.go            # Route registration
│   │   └── ...
│   └── migrate/
│       └── main.go              # Migration CLI tool
│
├── internal/                    # 132 packages
│   ├── config/                  # Config{} struct, Load()
│   ├── db/                      # PostgreSQL pool, migrations
│   ├── httputil/                # WriteJSON, WriteError
│   ├── llm/                     # Provider interface, SmartRouter
│   │   ├── provider.go          # Interface definition
│   │   ├── smartrouter.go       # Routing logic
│   │   ├── anthropic.go         # Anthropic implementation
│   │   ├── ollama.go            # Ollama implementation
│   │   └── circuit_breaker.go   # Failure handling
│   │
│   ├── mission/                 # Domain: missions
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── research/                # Agent: discovery
│   ├── pricing/                 # Agent: price hunting
│   ├── search/                  # Vector similarity search
│   ├── scorecard/               # Efficiency grading
│   ├── embed/                   # Embedding generator
│   ├── scheduler/               # Background jobs (cron)
│   │
│   ├── option/                  # Domain: options
│   ├── shopping/                # Domain: shopping items
│   ├── purchase/                # Domain: purchases
│   ├── notification/            # Domain: alerts
│   ├── template/                # Built-in templates
│   ├── settings/                # User settings (JSONB)
│   └── ...                      # 100+ more packages
│
├── migrations/                  # Database migrations
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   ├── ...
│   └── 022_*.up/down.sql
│
└── Dockerfile
```

### Frontend Deep Dive

```
frontend/
├── src/
│   ├── api/                     # API client layer (Zod + fetch)
│   │   ├── client.ts            # apiFetch wrapper
│   │   ├── mission.ts
│   │   ├── option.ts
│   │   ├── shopping.ts
│   │   ├── search.ts
│   │   └── ...                  # One per resource
│   │
│   ├── components/              # React components
│   │   ├── scouter/             # Design system (6 components)
│   │   ├── mission/             # Mission domain (5 components)
│   │   ├── options/             # Option domain (5 components)
│   │   ├── shopping/            # Shopping domain (5 components)
│   │   └── common/              # Shared (Breadcrumb, Modal, etc)
│   │
│   ├── pages/                   # Route-level components
│   │   ├── HQDashboard.tsx      # /
│   │   ├── MissionOverview.tsx  # /missions/:slug
│   │   ├── OptionsExplorer.tsx  # /missions/:slug/options
│   │   ├── ShoppingTracker.tsx  # /missions/:slug/shopping
│   │   ├── SearchPage.tsx       # /search
│   │   ├── HistoryPage.tsx      # /history
│   │   ├── StatsPage.tsx        # /stats
│   │   └── SettingsPage.tsx     # /settings
│   │
│   ├── hooks/                   # Custom React hooks (TanStack Query)
│   │   ├── useMission.ts
│   │   ├── useOptions.ts
│   │   ├── useShopping.ts
│   │   ├── useSearch.ts
│   │   └── useResearch.ts
│   │
│   ├── contexts/                # React Context
│   │   ├── sidebar.tsx
│   │   └── onboarding.tsx
│   │
│   ├── types/                   # TypeScript type definitions
│   │   ├── mission.ts           # Mirrored from Zod schemas
│   │   ├── option.ts
│   │   └── ...
│   │
│   ├── i18n/                    # Internationalization
│   │   ├── index.ts             # Hook: useTranslation()
│   │   ├── en.json              # English strings
│   │   └── fr.json              # French strings
│   │
│   ├── styles/
│   │   ├── theme.css            # CSS custom properties
│   │   ├── global.css           # Reset + base styles
│   │   └── ...                  # Component .module.css files
│   │
│   ├── App.tsx                  # Root router + layout
│   └── main.tsx                 # Entry point
│
├── e2e/                         # Playwright E2E tests
│   ├── *.spec.ts                # Test files (31 tests)
│   ├── fixtures.ts              # Test helpers
│   ├── screenshots/             # Test screenshots
│   └── test-results/            # Test reports
│
├── package.json
├── tsconfig.json
├── vite.config.ts
└── Dockerfile
```

### Database Schema

```
postgresql
├── missions
│   ├── id (UUID)
│   ├── slug (TEXT UNIQUE)
│   ├── name, budget, currency, category
│   ├── status (active|done|archived)
│   ├── constraints (JSONB)
│   └── timestamps
│
├── options
│   ├── id (UUID)
│   ├── mission_id (FK)
│   ├── name, description, attributes (JSONB)
│   ├── embedding (vector(1024))
│   ├── score, constraint_fit
│   └── timestamps
│
├── shopping_items
│   ├── id (UUID)
│   ├── mission_id (FK)
│   ├── name, merchant, price
│   ├── target_price, status
│   └── timestamps
│
├── price_history
│   ├── id (UUID)
│   ├── shopping_item_id (FK)
│   ├── price, merchant, date
│   └── deal_score
│
├── purchase_records
│   ├── id (UUID)
│   ├── mission_id (FK)
│   ├── option_id, price, date
│   ├── lessons (TEXT)
│   └── timestamps
│
├── notifications
│   ├── id (UUID)
│   ├── type, message, read_at
│   └── timestamps
│
├── settings
│   ├── key (TEXT PRIMARY KEY)
│   └── value (JSONB)
│
└── ... (22+ tables total)
```

---

## Quick Navigation

**I want to...**

| Task | Go To |
|------|-------|
| Understand the user flow | [User Guide](functional/user-guide.md) |
| Run locally | [Quick Start](deployment/quickstart.md) |
| Add a new API endpoint | [Contributing](development/contributing.md) |
| Debug a problem | [Runbook](RUNBOOK.md) |
| Understand request routing | Backend > `cmd/server/routes.go` |
| Add a new LLM provider | Backend > `internal/llm/provider.go` + implement |
| Add a new frontend page | Frontend > `src/pages/` + update router in `App.tsx` |
| Write a test | [Testing](development/testing.md) |
| Check API endpoints | [API Reference](technical/api.md) |
| Monitor in production | [Monitoring](deployment/monitoring.md) |

---

**For more context:** See [technical/architecture.md](technical/architecture.md) for detailed system diagrams.
