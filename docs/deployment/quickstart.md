# Quick Start

Get SCOUTER running with Docker Compose, Traefik reverse proxy, and HTTPS.

---

## Prerequisites

- Docker + Docker Compose v2
- OpenSSL (for local HTTPS certificates)
- (Optional) [Ollama](https://ollama.com) for local LLM
- (Optional) Anthropic API key for cloud LLM

---

## 1. Generate Local HTTPS Certificates

SCOUTER uses Traefik with HTTPS on `*.dev.local`. Generate self-signed certificates once:

```bash
make certs
```

Output files in `certs/`:
- `ca.crt` — root CA certificate
- `dev.local.crt` — wildcard certificate
- `dev.local.key` — private key

**Windows (WSL2)**: Install CA certificate in Windows:
```powershell
# Run as Administrator
certutil -addstore -f "ROOT" certs/ca.crt
```

**macOS**:
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/ca.crt
```

**Linux**: Copy to system trust store:
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

---

## 2. Clone and Configure

```bash
git clone <repo-url>
cd scouter
cp .env.example .env
```

Edit `.env` to set your LLM provider:

```bash
# Required
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter

# LLM (choose one):

# Option A — Anthropic (cloud, best quality)
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Option B — Ollama (local, free)
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
OLLAMA_EMBED_MODEL=mxbai-embed-large

# Option C — SmartRouter (Ollama + Anthropic cascade)
LLM_PROVIDER=routing
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
```

---

## 3. Start the Stack

```bash
# Default: postgres + backend + frontend + traefik (HTTPS)
make up

# Or: start core + seed sample data
make up-seed

# Or: start everything (core + seed + monitoring)
make up-full
```

Docker Compose will:
1. Start PostgreSQL and wait for health check
2. Start backend, which applies migrations automatically
3. Start frontend (Nginx)
4. Start Traefik reverse proxy

---

## 4. Access the Application

| Service | URL | Notes |
|---------|-----|-------|
| **SCOUTER App** | https://scouter.dev.local | Main app — HTTPS via Traefik |
| **Backend API** | https://scouter.dev.local/api | API endpoints — routed by Traefik |
| **Backend health** | https://scouter.dev.local/api/health | Health check endpoint |
| **Traefik Dashboard** | http://localhost:8082 | Reverse proxy status |

---

## 5. Load Sample Data (Optional)

```bash
make seed
```

Inserts two sample missions (Home Server 2026, Summer Holiday 2026) so you can explore the UI immediately.

---

## Stack Architecture

```
                         ┌─────────────────────────┐
                         │  HTTPS (*.dev.local)    │
                         │  Traefik v3.4           │
                         │  Ports: 80, 443, 8082   │
                         └────────┬────────────────┘
                                  │
                    ┌─────────────┼──────────────┐
                    │             │              │
            ┌───────▼──────┐ ┌────▼─────┐ ┌────▼─────┐
            │  Frontend    │ │ Backend  │ │Dashboard │
            │  Nginx:80    │ │ Go:8080  │ │:8082     │
            └──────────────┘ └────┬─────┘ └──────────┘
                                  │
                         ┌────────▼─────────┐
                         │   PostgreSQL     │
                         │   pgvector       │
                         │   (port 5432)    │
                         └──────────────────┘
```

**Routing:**
- `https://scouter.dev.local/` → Frontend (Nginx)
- `https://scouter.dev.local/api/*` → Backend (Go)

**Certificate Storage:**
- `/certs/` directory (mounted into Traefik)

---

## Development Mode (Local)

For hot reload with Vite + Go air:

**Terminal 1 — Backend:**
```bash
cd backend
go run ./cmd/server
# Listens on :8080
```

**Terminal 2 — Frontend:**
```bash
cd frontend
npm install
npm run dev
# Vite dev server on :5173 with hot reload
```

**Terminal 3 — PostgreSQL (if not running Docker):**
```bash
docker compose up postgres -d
```

Set `ENV=development` in `.env` for permissive CORS.

Access frontend at `http://localhost:5173` (dev server, not via Traefik).

---

## Monitoring Stack

Enable detailed observability with Prometheus, Grafana, and cAdvisor:

```bash
# Start core stack + monitoring
make up-monitoring

# Or: everything
make up-full
```

| Service | URL | Credentials |
|---------|-----|-------------|
| Prometheus | http://localhost:9090 | — |
| Grafana | http://localhost:3000 | admin / scouter |
| cAdvisor | http://localhost:8081 | — |
| Metrics endpoint | https://scouter.dev.local/api/metrics | — |

Enable metrics collection in `.env`:
```bash
METRICS_ENABLED=true
```

Grafana comes pre-provisioned with SCOUTER dashboards.

---

## Docker Compose Profiles

```bash
# All available:
# - (default): postgres, backend, frontend, traefik
# - seed: one-shot migration container
# - monitoring: prometheus, grafana, cadvisor

# Examples:
make up                 # core only
make up-seed            # core + seed
make up-monitoring      # core + monitoring
make up-full            # core + seed + monitoring
```

---

## Stopping & Cleanup

```bash
# Stop all containers (keep volumes)
make down

# Stop and remove all volumes (DESTROYS DATABASE)
make clean-volumes

# Show all container status
make ps
```

---

## Ollama Setup

If using local Ollama:

```bash
# Install
curl -fsSL https://ollama.com/install.sh | sh

# Pull models
ollama pull qwen3:14b
ollama pull qwen3:4b
ollama pull mxbai-embed-large

# Start Ollama (usually a system service on macOS/Windows)
# On Linux: ollama serve &
```

**Inside Docker containers**, `host.docker.internal:11434` automatically resolves to your host machine.

---

## Troubleshooting

### "certificate verify failed" in browser

The self-signed CA certificate was not installed in your system trust store. Run `make certs` again and reinstall the CA certificate (see Prerequisites → Install CA).

### Backend won't start: "DATABASE_URL required"

Set `DATABASE_URL` in your `.env` file.

### Research returns no results

Check LLM configuration:
```bash
curl https://scouter.dev.local/api/health/llm --insecure
```

If all providers show errors, check that Ollama is running or your API key is valid.

### Frontend can't reach backend API

Verify Traefik routing is healthy:
```bash
curl http://localhost:8082/api/http/routers
```

Set `ENV=development` in `.env` for permissive CORS during development.

### Migrations fail at startup

Check PostgreSQL is healthy:
```bash
docker compose ps postgres
docker compose logs postgres
```

### Port conflicts

Change Traefik ports in `docker-compose.yml`:
```yaml
traefik:
  ports:
    - "8080:80"    # HTTP redirect
    - "8443:443"   # HTTPS
    - "8082:8080"  # Dashboard
```

Then update your hosts file to point to the correct IP.
