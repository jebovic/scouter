# System Architecture

## Overview

![System Architecture](../assets/system-architecture.svg)

SCOUTER Universal is a full-stack application with four core deployment units orchestrated via Docker Compose:

| Unit | Technology | Port (container) | Port (host) |
|------|-----------|---------|-------|
| **Frontend** | React 19 + Vite + Nginx | 80 | — |
| **Backend** | Go 1.23 + chi router | 8080 | — |
| **Database** | PostgreSQL 16 + pgvector | 5432 | 5432 |
| **Reverse Proxy** | Traefik v3.4 | 80, 443, 8082 | 80, 443, 8082 |

All services are orchestrated via Docker Compose with health-check-based dependency ordering. The reverse proxy (Traefik) exposes HTTPS on `*.dev.local` in development (via self-signed certificates) and handles routing to frontend/backend.

---

## Architectural Principles

### 1. Transport-Only LLM Interface

The LLM Provider interface is deliberately **transport-only**:

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}
```

Business logic lives in agents, not in the provider. This means:
- ResearchAgent owns: prompt construction, tool schema, response parsing, DB persistence
- PricingAgent owns: price prompt, merchant tool schema, deal scoring
- SmartRouter owns: capability matching, circuit breaking, cascade fallback

### 2. Cursor-Based Pagination

All list endpoints use cursor-based pagination with a probe-row pattern:

```
GET /api/missions?limit=20&cursor=<uuid>
→ { data: [...], next_cursor: "...", total: N }
```

### 3. UUID Primary Keys + Slug Routing

- All DB primary keys are UUIDs (no sequential IDs)
- Missions also have a `slug TEXT UNIQUE NOT NULL` for human-readable URLs
- API routes use slug for mission lookup, UUID internally

### 4. Zod at the Frontend Boundary

Every API response is validated by a Zod schema in `src/api/`. TypeScript types are inferred from schemas — no manual type duplication.

### 5. Immutability Throughout

- Backend: all repository methods return new values, no in-place mutation
- Frontend: React state updates are immutable (Tanstack Query manages cache immutably)
- CSS: all values via custom properties, no inline style mutation

---

## Component Diagram

```
┌────────────────────────────────────────────────────┐
│         Browser (https://scouter.dev.local)        │
│  React 19 · React Router v7 · Tanstack Query v5    │
│  22 pages · 11 component dirs · i18n EN/FR          │
└──────────────────┬─────────────────────────────────┘
                   │ HTTPS / JSON
                   ▼
      ┌────────────────────────────────┐
      │   Traefik v3.4 Reverse Proxy   │
      │  · HTTPS (*.dev.local)         │
      │  · Docker provider + TLS file  │
      │  · Routes to frontend + /api   │
      └────────┬──────────┬────────────┘
               │          │
    ┌──────────▼┐  ┌──────▼─────────┐
    │ Frontend  │  │ Backend chi    │
    │ Nginx:80  │  │ Go:8080        │
    └──────────┘  └────┬─────────────┘
                       │
          ┌────────────┴────────────────┐
          │    Core Services             │
          │  Mission · Options · Shopping│
          │  Notification · Purchase     │
          │  Settings · Export · Search  │
          └────────┬────────────────────┘
                   │
          ┌────────┴────────────────────┐
          │    AI Agent Layer            │
          │  ResearchAgent · PricingAgent│
          │  SmartRouter (Phase 9)       │
          │  · Circuit breakers          │
          │  · Rate limiting             │
          │  · Cascade fallback          │
          │         ↓                    │
          │  Ollama Heavy/Fast/Cloud     │
          │  → Anthropic                 │
          └────────┬────────────────────┘
                   │
          ┌────────▼────────────────────┐
          │  PostgreSQL 16 + pgvector    │
          │  · 22+ migrations            │
          │  · UUID PKs · JSONB          │
          │  · IVFFlat index             │
          │  · 1024-dim embeddings       │
          └─────────────────────────────┘
```

---

## Request Lifecycle

![Data Flow](../assets/data-flow.svg)

For a `POST /api/missions/:id/research` call:

1. **Browser**: Tanstack Query mutation → `POST /api/missions/:id/research`
2. **chi Middleware**: CORS check → body size check → Prometheus counter
3. **Handler**: JSON decode → validate → call `ResearchService.Run(ctx, missionID)`
4. **Service**: Fetch mission from DB → build `CompletionRequest` → `ResearchAgent.Discover()`
5. **ResearchAgent**: Build prompt + tool schema → `WithRequestOpts(ctx, hint)` → `SmartRouter.Complete()`
6. **SmartRouter**: Match capabilities → check circuit breaker → rate limit → call Ollama Heavy (if available)
7. **LLM**: Return tool_calls JSON → parsed by ResearchAgent → options extracted
8. **Repository**: Batch INSERT options → return inserted rows
9. **Embed Worker**: Option text sent to OllamaEmbedder (async, 2 goroutines) → `vector(1024)` stored
10. **Response**: `200 OK` + options JSON → Zod validated → React state updated

---

## Key Design Decisions

### SmartRouter Cascade

```
Ollama Heavy (qwen3:14b)
  → fail → Ollama Fast (qwen3:4b)
    → fail → Ollama Cloud (deepseek-v3.2:cloud)
      → fail → Anthropic (claude-sonnet-4-6)
```

"Fail" means infrastructure error (timeout, connection refused, circuit open) — **not** a bad LLM response. Bad responses trigger `RetryAsJSON` on the same provider.

### golang-migrate Auto-Run

Migrations run automatically at server startup:

```go
func runMigrations(db *sql.DB) error {
    driver, _ := postgres.WithInstance(db, &postgres.Config{})
    m, _ := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
    return m.Up()
}
```

This means the database is always in sync with the binary in production.

### Graceful Shutdown

The server waits up to **65 seconds** for in-flight requests (exceeds the 60s LLM call timeout):

```go
ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### httputil Buffer Pattern

All responses are buffered before writing headers to prevent split responses:

```go
func WriteJSON(w http.ResponseWriter, status int, v any) {
    buf := bytes.Buffer{}
    json.NewEncoder(&buf).Encode(v)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(buf.Bytes())
}
```

---

## Deployment & Reverse Proxy

### Traefik (Development)

Traefik v3.4 provides local HTTPS with self-signed certificates:

```bash
# Generate certificates (one-time)
make certs

# Start with HTTPS enabled
make up
# Access at https://scouter.dev.local (HTTP redirects to HTTPS)
```

**Certificate generation** uses OpenSSL to create:
- Local CA certificate (`certs/ca.crt`)
- Wildcard certificate for `*.dev.local` (`certs/dev.local.crt`, `certs/dev.local.key`)

**Traefik configuration:**
- Docker provider (auto-labels services)
- File provider for TLS certificates
- Dashboard at `http://localhost:8082`
- Auto-redirect HTTP → HTTPS

### Local Development (without Traefik)

For hot reload without HTTPS:

```bash
# Terminal 1: PostgreSQL
docker compose up postgres -d

# Terminal 2: Backend (Go)
cd backend && go run ./cmd/server  # :8080

# Terminal 3: Frontend (Vite)
cd frontend && npm run dev          # :5173
```

Set `ENV=development` in `.env` for CORS.

---

## Monitoring Stack (Phase 14)

Start with the monitoring profile:

```bash
make up-monitoring
```

| Service | URL | Purpose |
|---------|-----|---------|
| Prometheus | http://localhost:9090 | Metrics scraping |
| Grafana | http://localhost:3000 | Pre-provisioned dashboards |
| cAdvisor | http://localhost:8081 | Container resource metrics |
| Backend metrics | https://scouter.dev.local/api/metrics | Prometheus endpoint (Traefik HTTPS) |

**Internal scraping:** Prometheus scrapes backend on `http://backend:8080/api/metrics` (Docker network).
