# Quick Start

Get SCOUTER running in 5 minutes with Docker Compose.

---

## Prerequisites

- Docker + Docker Compose v2
- (Optional) [Ollama](https://ollama.com) for local LLM
- (Optional) Anthropic API key for cloud LLM

---

## 1. Clone and Configure

```bash
git clone <repo-url>
cd scouter
cp .env.example .env
```

Edit `.env`:

```bash
# Required
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter

# LLM (choose one):

# Option A — Anthropic (cloud, best quality)
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Option B — Ollama (local, free, requires Ollama running)
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
OLLAMA_EMBED_MODEL=mxbai-embed-large

# Option C — SmartRouter (tries Ollama first, falls back to Anthropic)
LLM_PROVIDER=routing
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
```

---

## 2. Start

```bash
# Default stack (postgres + backend + frontend)
make dev
# or:
docker compose up --build
```

On first start:
- PostgreSQL initializes and becomes healthy
- Backend waits for DB → applies all 22+ migrations automatically
- Frontend builds and starts Nginx

---

## 3. Open

| Service | URL |
|---------|-----|
| **SCOUTER App** | http://localhost:5173 |
| **Backend API** | http://localhost:8080/api/health |

---

## 4. Load Sample Data (Optional)

```bash
make seed
```

This creates a sample "Work Laptop" mission with options and shopping items so you can explore the UI immediately.

---

## Development Mode

For local development with hot reload:

```bash
# Backend (Go with air or go run)
cd backend
go run ./cmd/server

# Frontend (Vite dev server)
cd frontend
npm install
npm run dev
```

Set `ENV=development` in your `.env` for permissive CORS.

---

## Monitoring Stack

```bash
docker compose --profile monitoring up
```

| Service | URL |
|---------|-----|
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 (admin/admin) |
| cAdvisor | http://localhost:8081 |

Grafana comes pre-provisioned with SCOUTER dashboards.

Enable metrics in `.env`:

```bash
METRICS_ENABLED=true
```

---

## Stopping

```bash
# Stop containers
docker compose down

# Stop and delete volumes (wipes DB)
make clean
```

---

## Ollama Setup

If using Ollama locally:

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

Ollama must be running before starting SCOUTER. On macOS/Windows, Ollama runs as a system service. On Linux, run `ollama serve` in a background terminal.

The `host.docker.internal` hostname automatically resolves to your host machine from inside Docker containers.

---

## Troubleshooting

### Backend won't start: "DATABASE_URL required"

Set `DATABASE_URL` in your `.env` file.

### Research returns no results

Check LLM configuration:
```bash
curl http://localhost:8080/api/health/llm
```

If all providers show errors, check that Ollama is running or your API key is valid.

### Frontend can't reach backend

Check CORS: set `ENV=development` in `.env` for local development.

### Migrations fail

Check that `DATABASE_URL` points to a healthy PostgreSQL instance:
```bash
docker compose ps postgres
docker compose logs postgres
```

### Port conflicts

Change ports in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"   # backend on 8081
  - "5174:5173"   # frontend on 5174
```
