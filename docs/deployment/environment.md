# Environment Variables

Complete reference for all SCOUTER environment variables. See `.env.example` in the project root.

---

## Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://scouter:scouter@postgres:5432/scouter` |

---

## LLM Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `ollama` | Provider mode: `anthropic`, `ollama`, or `routing` (SmartRouter) |
| `ANTHROPIC_API_KEY` | — | Required when `LLM_PROVIDER=anthropic` or `routing` |

### Ollama (Local LLM)

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_BASE_URL` | `http://host.docker.internal:11434` | Local Ollama endpoint (accessible from Docker containers) |
| `OLLAMA_HEAVY_MODEL` | `qwen3:14b` | Primary tool-use model for complex research |
| `OLLAMA_FAST_MODEL` | `qwen3:4b` | Lightweight fallback for simple queries |
| `OLLAMA_EMBED_MODEL` | `mxbai-embed-large` | 1024-dimensional embedding model (Phase 11) |
| `OLLAMA_HEAVY_TIMEOUT` | `180` | Timeout in seconds for heavy model |
| `OLLAMA_FAST_TIMEOUT` | `60` | Timeout in seconds for fast model |
| `OLLAMA_MODEL` | — | Legacy alias (maps to `OLLAMA_HEAVY_MODEL`) |

### Ollama Cloud (Remote)

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_CLOUD_URL` | — | Remote Ollama endpoint (e.g. `https://ollama.com`) |
| `OLLAMA_CLOUD_MODEL` | — | Model identifier (e.g. `deepseek-v3.2:cloud`) |
| `OLLAMA_CLOUD_API_KEY` | — | Bearer token from ollama.com |
| `OLLAMA_CLOUD_RPM` | `10` | Rate limit (requests per minute) |

---

## Server & Deployment

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend listen port (inside Docker) |
| `ENV` | `production` | `development` enables permissive CORS (all origins) |
| `VITE_API_BASE_URL` | (relative) | Frontend API base URL (used in dev server only) |

---

## Observability

| Variable | Default | Description |
|----------|---------|-------------|
| `METRICS_ENABLED` | — | Set to `true` to enable Prometheus metrics at `GET /api/metrics` |
| `GF_SECURITY_ADMIN_PASSWORD` | `scouter` | Grafana admin password (for monitoring profile) |

---

## LLM_PROVIDER Modes

### `anthropic`

Uses only Anthropic's API (cloud). Requires `ANTHROPIC_API_KEY`.

```bash
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...
```

**Best for:** Production use, highest quality, no local setup required.

---

### `ollama`

Uses only local Ollama. No API key required.

```bash
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
OLLAMA_EMBED_MODEL=mxbai-embed-large
```

**Best for:** Development, cost-free local use, privacy, offline capability.

---

### `routing` (SmartRouter)

Capability-matched routing: tries Ollama first, cascades to Anthropic on failure. Each model has its own circuit breaker and rate limiter (Phase 9).

```bash
LLM_PROVIDER=routing
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
ANTHROPIC_API_KEY=sk-ant-...
```

**Cascade order:**
1. Ollama heavy model (complex queries)
2. Ollama fast model (fallback)
3. Ollama cloud (optional, if configured)
4. Anthropic (final fallback)

**Best for:** Production with cost optimization (free local Ollama when available, cloud as safety net).

---

## Docker Compose Injection

The `.env` file is automatically loaded by Docker Compose:

```yaml
# docker-compose.yml excerpt
services:
  backend:
    env_file: .env
    environment:
      DATABASE_URL: postgres://scouter:scouter@postgres:5432/scouter
      ENV: development
      LLM_PROVIDER: routing
```

Frontend uses relative URLs in development; `VITE_API_BASE_URL` is only used when building Vite with a custom base URL.

---

## Port Configuration

**Docker Compose (default):**
- PostgreSQL: `5432` (host machine)
- Backend: `8080` (inside container, not exposed to host)
- Frontend: `80` (inside container, not exposed to host)
- Traefik: `80` (HTTP, redirects to 443), `443` (HTTPS), `8082` (Dashboard)

**Local development (hot reload, without Docker):**
- Backend: `8080` (http://localhost:8080)
- Frontend: `5173` (http://localhost:5173, Vite dev server with hot reload)
- PostgreSQL: `5432` (Docker or local)

---

## Complete .env.example

See `/home/jibei/projects/scouter/.env.example` for the canonical reference.
