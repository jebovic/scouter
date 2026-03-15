# SCOUTER Universal

> Personal spending intelligence tool for smart major purchase decisions.
>
> Research, compare, and budget any big purchase from the first question through purchase and reflection.

**Version:** 0.1.0 · **Status:** Production-ready · **License:** MIT

---

## What SCOUTER Does

SCOUTER guides you from "Should I buy this?" through research, price tracking, purchase decision, and post-purchase reflection.

### The Flow

1. **Create a Mission** — Name your purchase goal, set budget, define constraints (weight, RAM, etc.)
2. **Run Research** — ResearchAgent (LLM + tool use) discovers and ranks options for you
3. **Hunt Prices** — PricingAgent checks real prices across merchants, calculates Total Cost of Ownership
4. **Compare** — Side-by-side comparison with radar charts, deal scores, and constraint checking
5. **Track Spending** — Price history charts, budget burn rate, deal alerts, merchant comparison
6. **Record Purchase** — Log what you actually bought, capture lessons learned
7. **Analyze Results** — Efficiency scorecard (A/B/C/D), spending insights, future recommendations

### Screenshots

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin: 20px 0;">

![Dashboard with missions](docs/assets/screenshot-dashboard-with-missions.png)
*Dashboard showing active missions with budget bars and status*

![Mission Overview](docs/assets/screenshot-mission-overview.png)
*Mission detail page with timeline and key metrics*

![Options Comparison](docs/assets/screenshot-mission-options.png)
*Compare options side-by-side with constraint checking*

![Shopping Tracker](docs/assets/screenshot-mission-shopping.png)
*Price history, budget burn, and deal alerts*

![Semantic Search](docs/assets/screenshot-search.png)
*Find similar options across all missions*

![Settings](docs/assets/screenshot-settings.png)
*Configure currency, language, and LLM provider*

</div>

---

## Key Features

### Research & Pricing Intelligence

| Feature | Capability |
|---------|-----------|
| **ResearchAgent** | Discovers options using LLM tool use, ranks by your constraints |
| **PricingAgent** | Hunts real prices across merchants, calculates TCO, scores deals |
| **Deal Scoring** | Real-time scoring combining price trend, urgency, budget fit |
| **Semantic Search** | Find similar options across all missions (pgvector, 300ms debounce) |
| **Price Alerts** | Background scheduler checks prices, fires notifications on changes |

### Comparison & Analysis

| Feature | Capability |
|---------|-----------|
| **Radar Charts** | Visual comparison of options against constraints |
| **Constraint Checker** | Highlight options that match/violate your requirements |
| **TCO Calculator** | Compare total cost of ownership (price + maintenance + shipping) |
| **Deal Calendar** | Flash sales and promotional calendar |
| **French Benchmark** | Compare price to French market median (bon prix / prix moyen / au-dessus) |

### Budget & Spending

| Feature | Capability |
|---------|-----------|
| **Budget Bar** | Visual budget burn progress per mission |
| **Purchase Timeline** | 4-week plan with budget distribution and French promo hints |
| **Quantity Optimizer** | Discount tiers [1, 2, 3, 5, 10] with FNV-32a discount curve |
| **Wishlist Prioritizer** | Rank items by urgency + trend + budget fit |
| **Scorecard** | Efficiency grade (A/B/C/D), analytics, 4-metric breakdown |

### Collaboration & Sharing

| Feature | Capability |
|---------|-----------|
| **Mission Sharing** | Share tokens for read-only public access |
| **Collaborators** | Invite others to join, delegate research |
| **Voting** | Collaborators vote on options |
| **Threaded Comments** | Discuss options with team |

### Quality & Usability

| Feature | Capability |
|---------|-----------|
| **i18n** | Full English + French translations (react-i18next) |
| **Keyboard Shortcuts** | `N` new mission · `R` research · `P` pricing |
| **Dark/Light Theme** | Automatic theme switching with system preference |
| **Responsive Design** | Mobile, tablet, desktop optimized |
| **PWA Support** | Install as app, offline capability |

---

## Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.23 · chi router · pgx/v5 · golang-migrate |
| **Frontend** | React 19 · TypeScript · Vite · React Router v7 · Tanstack Query v5 |
| **Database** | PostgreSQL 16 + pgvector (IVFFlat 1024-dim embeddings) |
| **LLM** | Anthropic claude-sonnet-4-6 (tool use) · Ollama (local) · SmartRouter |
| **Testing** | Go tests · Vitest + Testing Library · Playwright E2E (31 tests) |
| **Deployment** | Docker Compose · Traefik v3.4 (reverse proxy + HTTPS) · Prometheus · Grafana |

---

## Quick Start

### Prerequisites

- Docker + Docker Compose v2
- (Optional) [Ollama](https://ollama.ai) for local LLM, or [Anthropic API key](https://console.anthropic.com)

### 1. Clone & Configure

```bash
git clone <repo-url>
cd scouter
cp .env.example .env
```

Edit `.env` and set:

```bash
# Required: PostgreSQL
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter

# LLM (choose one):
# Option A: Anthropic (best quality, requires API key)
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Option B: Ollama (local, free, requires Ollama running on port 11434)
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b

# Option C: SmartRouter (tries Ollama first, falls back to Anthropic)
LLM_PROVIDER=routing
```

### 2. Start All Services

```bash
make up
# or: docker compose up --build
```

On first start:
- PostgreSQL initializes and becomes healthy
- Backend applies all 22+ migrations automatically
- Frontend builds and starts on Nginx
- Traefik starts and routes HTTPS traffic

### 3. Open & Load Sample Data

```
Frontend:  https://scouter.dev.local
API:       https://scouter.dev.local/api/health
Traefik:   http://localhost:8082
```

Load sample data (optional):
```bash
make seed
```

---

## Make Targets

See `make help` for full list. Common targets:

```bash
make up           # Start core stack (postgres + backend + frontend + traefik)
make up-seed      # Start core stack + load sample data
make up-monitoring # Start core + Prometheus + Grafana
make up-full      # Start everything (core + seed + monitoring)

make test         # Run backend Go tests
make test-frontend # Run frontend Vitest tests
make test-e2e     # Run Playwright E2E tests
make test-all     # Run all tests

make lint         # Lint Go code (go vet)
make lint-frontend # Lint frontend (eslint)
make typecheck    # TypeScript type check

make certs        # Generate local HTTPS certificates
make seed         # Load sample mission data
make migrate-down # Rollback one database migration
make down         # Stop all containers (keep volumes)
make clean-volumes # Stop and remove all volumes (DESTROYS DATA)
```

---

## Environment Variables

### Core Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | yes | — | PostgreSQL: `postgres://user:pass@host:5432/db` |
| `PORT` | no | `8080` | Backend listen port |
| `ENV` | no | `production` | `development` enables permissive CORS |

### LLM Provider Selection

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LLM_PROVIDER` | no | `ollama` | `anthropic` \| `ollama` \| `routing` |
| `ANTHROPIC_API_KEY` | conditional | — | Required when using Anthropic provider |

### Ollama Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OLLAMA_BASE_URL` | no | `http://host.docker.internal:11434` | Ollama server endpoint |
| `OLLAMA_HEAVY_MODEL` | no | `qwen3:14b` | Primary tool-use model (Phase 9+) |
| `OLLAMA_FAST_MODEL` | no | `qwen3:4b` | Fallback lightweight model (Phase 9+) |
| `OLLAMA_EMBED_MODEL` | no | `mxbai-embed-large` | Embedding model for semantic search (Phase 11+) |
| `OLLAMA_HEAVY_TIMEOUT` | no | `180` | Timeout in seconds for heavy model |
| `OLLAMA_FAST_TIMEOUT` | no | `60` | Timeout in seconds for fast model |

### Optional Features

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `METRICS_ENABLED` | no | `false` | Enable Prometheus metrics endpoint |
| `GF_SECURITY_ADMIN_PASSWORD` | no | `scouter` | Grafana admin password |

For complete reference, see [docs/deployment/environment.md](docs/deployment/environment.md).

---

## API Endpoints (170+)

### Core Resources

```
GET    /api/health                           # System health (DB, version)
GET    /api/health/llm                       # LLM provider status

GET    /api/missions                         # List (cursor-paginated)
POST   /api/missions                         # Create mission
GET    /api/missions/:slug                   # Get by slug
PATCH  /api/missions/:slug                   # Update mission
DELETE /api/missions/:slug                   # Delete mission
```

### AI Agents

```
POST   /api/missions/:id/research            # Trigger ResearchAgent
POST   /api/missions/:id/pricing             # Trigger PricingAgent
```

### Shopping & Options

```
GET    /api/missions/:id/options             # List options
POST   /api/missions/:id/options             # Add option

GET    /api/missions/:id/shopping            # List shopping items
POST   /api/missions/:id/shopping            # Add item + price
GET    /api/missions/:id/deal-score          # Mission deal score
```

### Intelligence

```
GET    /api/search?q=laptop                  # Semantic search (pgvector)
GET    /api/options/:id/similar              # Similar options

GET    /api/missions/:id/french-benchmark    # Market comparison
GET    /api/missions/:id/scorecard           # Efficiency grade (A/B/C/D)
GET    /api/missions/:id/purchase-timeline   # 4-week plan
GET    /api/wishlist/prioritized             # Ranked wishlist
```

### Collaboration & Sharing

```
POST   /api/missions/:slug/share             # Generate share token
DELETE /api/missions/:id/share               # Revoke token
GET    /api/shared/:token                    # Read-only public access

POST   /api/missions/:id/invites             # Invite collaborator
GET    /api/missions/:id/collaborators       # List collaborators
```

### Analytics

```
GET    /api/stats                            # Total spent + category breakdown
GET    /api/notifications                    # Price alerts + updates
```

For complete API reference with request/response examples, see [docs/technical/api.md](docs/technical/api.md).

---

## Project Structure

```
backend/
├── cmd/server/
│   ├── main.go           # entrypoint, env validation, startup
│   └── routes.go         # route registration (routeDeps + registerRoutes)
├── internal/             # 132 Go packages
│   ├── config/           # Config struct, Load() from env
│   ├── db/               # pgx pool, golang-migrate runner
│   ├── httputil/         # WriteJSON / WriteError (buffer-first)
│   ├── llm/              # Provider interface, SmartRouter, circuit breakers
│   ├── mission/          # model, repository, service, handler
│   ├── option/           # model, repository, service, handler
│   ├── shopping/         # model, repository, service, handler
│   ├── research/         # ResearchAgent (LLM tool use, option discovery)
│   ├── pricing/          # PricingAgent (price hunting, deal scoring)
│   ├── search/           # pgvector semantic search (Phase 11+)
│   ├── scorecard/        # Efficiency grading (Phase 170)
│   └── ...               # 100+ other domain packages
├── migrations/           # NNN_description.up/down.sql (golang-migrate)
└── Dockerfile

frontend/
├── src/
│   ├── api/              # Typed fetch wrappers + Zod schemas
│   ├── components/       # UI components with CSS Modules
│   │   ├── scouter/      # Design system (Card, Badge, BudgetBar…)
│   │   ├── mission/      # MissionCard, MissionForm…
│   │   ├── options/      # OptionCard, ComparisonTable…
│   │   └── shopping/     # ShoppingList, PriceHistoryChart…
│   ├── pages/            # Route-level components (HQDashboard, MissionOverview…)
│   ├── hooks/            # TanStack Query hooks (useMission, useOptions…)
│   ├── styles/           # theme.css (tokens), global.css
│   ├── i18n/             # en.json, fr.json translations
│   └── types/            # TypeScript types (mirrored from Zod)
├── e2e/                  # Playwright E2E tests (31 tests, 999 unit tests)
└── Dockerfile

docs/
├── README.md             # Documentation index
├── functional/           # User-facing features and workflows
├── technical/            # Architecture, API, database
├── development/          # Contributing, testing, conventions
└── deployment/           # Setup, environment, monitoring

deployment/
├── docker-compose.yml    # All services (postgres, backend, frontend, traefik, monitoring)
├── traefik/              # Traefik static config + TLS
├── monitoring/           # Prometheus + Grafana config
└── certs/                # Generated TLS certificates (gitignored)
```

---

## Documentation

Complete documentation is in the `/docs` directory:

| Section | Purpose |
|---------|---------|
| [Quick Start](docs/deployment/quickstart.md) | Get running in 5 minutes |
| [Runbook](docs/RUNBOOK.md) | Operations, monitoring, troubleshooting |
| [Codemaps](docs/CODEMAPS.md) | Navigate the codebase and architecture |
| [Functional Overview](docs/functional/overview.md) | Feature catalog and use cases |
| [User Guide](docs/functional/user-guide.md) | How to use SCOUTER end-to-end |
| [API Reference](docs/technical/api.md) | All 170+ endpoints with examples |
| [Architecture](docs/technical/architecture.md) | System design and data flow |
| [Contributing](docs/development/contributing.md) | Dev workflow, TDD, conventions |
| [Testing](docs/development/testing.md) | Test strategy and coverage |
| [Environment Variables](docs/deployment/environment.md) | Configuration reference |
| [Monitoring](docs/deployment/monitoring.md) | Prometheus + Grafana setup |

---

## Testing

### Backend

```bash
cd backend && go test ./... -v          # All unit + integration tests
```

Target: 80%+ coverage across all packages.

### Frontend

```bash
cd frontend && npm run test             # Vitest + Testing Library
cd frontend && npm run test:e2e         # Playwright (31 E2E tests)
cd frontend && npm run test:e2e:report  # View test report
```

### Combined

```bash
make test                                # Run backend tests
cd frontend && npm run test              # Run frontend unit tests
cd frontend && npm run test:e2e          # Run E2E tests
```

---

## Development

### Local Setup (with hot reload)

```bash
# Backend (Go)
cd backend && go run ./cmd/server

# Frontend (Vite dev server, in another terminal)
cd frontend && npm install && npm run dev
```

Set `ENV=development` in `.env` for permissive CORS.

### Ollama Setup (for local LLM)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull models
ollama pull qwen3:14b
ollama pull qwen3:4b
ollama pull mxbai-embed-large

# Verify
ollama list
```

Ollama must be running before starting SCOUTER. The `host.docker.internal` hostname automatically resolves from inside Docker.

### Local HTTPS with Traefik

SCOUTER uses Traefik v3.4 as a reverse proxy. To run with HTTPS locally:

#### 1. Generate Local Certificates

```bash
make certs
```

This generates:
- `deployment/certs/ca.crt` — Local CA certificate
- `deployment/certs/dev.local.crt` — Wildcard cert for `*.dev.local`
- `deployment/certs/dev.local.key` — Private key

#### 2. Install CA Certificate (Windows)

On Windows, install the CA in your trusted roots (one-time):

```cmd
certutil -addstore -f "ROOT" deployment\certs\ca.crt
```

On macOS/Linux, double-click `deployment/certs/ca.crt` in your file explorer to add to the system keychain.

#### 3. Update Hosts File

Add to your system hosts file:
- **Windows:** `C:\Windows\System32\drivers\etc\hosts`
- **macOS/Linux:** `/etc/hosts`

```
127.0.0.1 scouter.dev.local
127.0.0.1 api.dev.local
```

#### 4. Start the Stack

```bash
make dev
```

SCOUTER is now accessible at:
- **App:** https://scouter.dev.local (green padlock in Chrome)
- **API:** https://scouter.dev.local/api/health
- **Traefik Dashboard:** http://localhost:8082

HTTP requests on port 80 redirect to HTTPS on port 443.

---

## Contributing

SCOUTER follows a structured phase-based development workflow with:

- Test-driven development (TDD)
- Code review via specialized agents
- Semantic versioning
- Conventional commits

See [docs/development/contributing.md](docs/development/contributing.md) for detailed instructions.

---

## Support & Issues

- [GitHub Issues](https://github.com/jibei/scouter/issues) — Bug reports and feature requests
- [Discussions](https://github.com/jibei/scouter/discussions) — General questions
- [Roadmap](ROADMAP.md) — Planned features (172+ phases complete)

---

## License

MIT — See [LICENSE](LICENSE) file
