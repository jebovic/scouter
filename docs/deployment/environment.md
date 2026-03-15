# Environment Variables

Complete reference for all SCOUTER environment variables.

---

## Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/scouter` |

---

## LLM Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | `ollama` | Provider mode: `anthropic`, `ollama`, or `routing` |
| `ANTHROPIC_API_KEY` | — | Required when `LLM_PROVIDER=anthropic` or `routing` |

### Ollama (Local)

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_BASE_URL` | `http://host.docker.internal:11434` | Local Ollama endpoint |
| `OLLAMA_MODEL` | — | Legacy alias → heavy model |
| `OLLAMA_HEAVY_MODEL` | `qwen3:14b` | Primary tool-use model (Phase 9) |
| `OLLAMA_FAST_MODEL` | `qwen3:4b` | Lightweight fallback (Phase 9) |
| `OLLAMA_EMBED_MODEL` | `mxbai-embed-large` | Embedding model, 1024-dim (Phase 11) |

### Ollama Cloud

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_CLOUD_URL` | — | Remote Ollama endpoint (e.g. `https://ollama.com`) |
| `OLLAMA_CLOUD_MODEL` | — | e.g. `deepseek-v3.2:cloud` |
| `OLLAMA_CLOUD_API_KEY` | — | Bearer token from ollama.com |

---

## Server

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend listen port |
| `ENV` | `production` | `development` enables permissive CORS (all origins) |

---

## Observability

| Variable | Default | Description |
|----------|---------|-------------|
| `METRICS_ENABLED` | — | Set to `true` to enable Prometheus endpoint at `/metrics` |

---

## LLM_PROVIDER Modes

### `anthropic`

Uses only Anthropic's API. Requires `ANTHROPIC_API_KEY`.

```bash
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-api03-...
```

Best for: production use, highest quality results.

### `ollama`

Uses local Ollama only. No API key required.

```bash
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
OLLAMA_EMBED_MODEL=mxbai-embed-large
```

Best for: development, cost-free local use, privacy.

### `routing`

SmartRouter: tries Ollama first, cascades to Anthropic on failure.

```bash
LLM_PROVIDER=routing
OLLAMA_BASE_URL=http://host.docker.internal:11434
OLLAMA_HEAVY_MODEL=qwen3:14b
OLLAMA_FAST_MODEL=qwen3:4b
ANTHROPIC_API_KEY=sk-ant-api03-...
```

Best for: production with cost optimization (use free local when available, cloud as fallback).

---

## .env.example

```bash
# ── Required ──────────────────────────────
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter

# ── LLM Provider (choose one) ─────────────

# Anthropic cloud
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=

# Ollama local
# LLM_PROVIDER=ollama
# OLLAMA_BASE_URL=http://host.docker.internal:11434
# OLLAMA_HEAVY_MODEL=qwen3:14b
# OLLAMA_FAST_MODEL=qwen3:4b
# OLLAMA_EMBED_MODEL=mxbai-embed-large

# SmartRouter (Ollama + Anthropic fallback)
# LLM_PROVIDER=routing

# ── Server ────────────────────────────────
PORT=8080
ENV=production  # set to "development" for permissive CORS

# ── Observability ─────────────────────────
# METRICS_ENABLED=true
```

---

## Docker Compose Variable Injection

Variables are automatically read from `.env` by Docker Compose:

```yaml
# docker-compose.yml
services:
  backend:
    environment:
      DATABASE_URL: ${DATABASE_URL}
      LLM_PROVIDER: ${LLM_PROVIDER:-ollama}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
      OLLAMA_BASE_URL: ${OLLAMA_BASE_URL:-http://host.docker.internal:11434}
      PORT: ${PORT:-8080}
      ENV: ${ENV:-production}
```
