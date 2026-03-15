# System Architecture

## Overview

![System Architecture](../assets/system-architecture.svg)

SCOUTER Universal is a full-stack application with three deployment units:

| Unit | Technology | Port |
|------|-----------|------|
| **Frontend** | React 19 + Vite + Nginx | 5173 |
| **Backend** | Go 1.23 + chi router | 8080 |
| **Database** | PostgreSQL 16 + pgvector | 5432 |

All three are orchestrated via Docker Compose with health-check-based dependency ordering.

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
┌─────────────────────────────────────────────────────┐
│                    Browser (PWA)                     │
│  React 19 · React Router v7 · Tanstack Query v5     │
│  22 pages · 11 component dirs · i18n EN/FR           │
└──────────────────┬──────────────────────────────────┘
                   │ HTTP / JSON
                   ▼
┌─────────────────────────────────────────────────────┐
│              chi Router (Go 1.23)                    │
│  170+ routes · CORS · 1MiB body cap · Prometheus    │
└───────────┬─────────────────────────────────────────┘
            │
   ┌────────┴─────────────────────────────────────┐
   │          Core Services                        │
   │  Mission · Options · Shopping · Notification  │
   │  Purchase · Settings · Export · Search        │
   └────────┬─────────────────────────────────────┘
            │
   ┌────────┴─────────────────────────────────────┐
   │          AI Agent Layer                       │
   │  ResearchAgent · PricingAgent · 140+ agents   │
   │          ↓                                    │
   │      SmartRouter (Phase 9)                    │
   │  Circuit breakers · Rate limiting · Cascade   │
   │          ↓                                    │
   │  Ollama Heavy/Fast/Cloud → Anthropic          │
   └────────┬─────────────────────────────────────┘
            │
   ┌────────┴─────────────────────────────────────┐
   │          PostgreSQL 16 + pgvector             │
   │  22+ migrations · UUID PKs · JSONB            │
   │  IVFFlat index · 1024-dim embeddings          │
   └──────────────────────────────────────────────┘
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

## Monitoring Stack (Phase 14)

Start with the monitoring profile:

```bash
docker compose --profile monitoring up
```

| Service | URL | Purpose |
|---------|-----|---------|
| Prometheus | :9090 | Metrics scraping |
| Grafana | :3000 | Pre-provisioned dashboards |
| cAdvisor | :8080 | Container resource metrics |
| Backend metrics | :8080/metrics | Prometheus endpoint |
