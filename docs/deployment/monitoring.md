# Monitoring & Observability

SCOUTER includes optional monitoring stack (Phase 14): Prometheus, Grafana, and cAdvisor.

---

## Starting the Monitoring Stack

```bash
# Start core + monitoring profile
make up-monitoring

# Or: start everything
make up-full
```

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / scouter |
| Prometheus | http://localhost:9090 | — |
| cAdvisor | http://localhost:8081 | — |
| Metrics endpoint | https://scouter.dev.local/api/metrics | — |

Enable metrics collection in `.env`:

```bash
METRICS_ENABLED=true
```

---

## Architecture

The monitoring stack operates independently of Traefik:

```
Backend (port 8080 internal)
    ├── GET /api/metrics           → Prometheus scrapes every 15s
    ├── GET /api/health            → Basic health check
    └── GET /api/health/llm        → LLM provider status

Prometheus (port 9090)
    │
    ├─→ Grafana (port 3000)         ← dashboards, alerts
    │
    └─→ Alertmanager               ← (optional) notification routing
```

**Note:** Prometheus scrapes the backend directly on `http://backend:8080/api/metrics` (internal Docker network). External access to `/api/metrics` goes through Traefik HTTPS.

---

## Metrics Collected

### HTTP Layer (chi middleware)

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_http_requests_total` | Counter | `method`, `route`, `status` |
| `scouter_http_request_duration_seconds` | Histogram | `method`, `route` |
| `scouter_http_request_size_bytes` | Histogram | `method`, `route` |

### LLM Layer (SmartRouter — Phase 9)

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_llm_requests_total` | Counter | `provider`, `model`, `status` |
| `scouter_llm_request_duration_seconds` | Histogram | `provider`, `model` |
| `scouter_llm_cascade_total` | Counter | `from_provider`, `to_provider` |
| `scouter_llm_circuit_open_total` | Counter | `provider` |

### Agent Layer

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_agent_calls_total` | Counter | `agent`, `status` |
| `scouter_agent_duration_seconds` | Histogram | `agent` |

### Scheduler (Phase 7)

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_price_checks_total` | Counter | `status` |
| `scouter_alerts_fired_total` | Counter | — |

---

## Grafana Dashboards

Pre-provisioned dashboards (auto-loaded on startup from `monitoring/grafana/provisioning/`):

### SCOUTER Overview

- Request rate (req/s)
- p50 / p95 / p99 latency
- Error rate by endpoint
- Active missions count

### LLM Intelligence

- Requests per provider (stacked area)
- LLM latency by model
- Cascade events (SmartRouter fallback frequency)
- Circuit breaker state timeline

### Agent Performance

- Top 10 agents by call volume
- Agent error rate
- Embedding worker queue depth
- Research/Pricing agent latency

### Infrastructure (via cAdvisor)

- CPU usage per container
- Memory usage per container
- Disk I/O (PostgreSQL)
- Network in/out

---

## Health Check Endpoints

### Backend Health

```bash
curl https://scouter.dev.local/api/health --insecure
# { "status": "ok", "db": "connected" }
```

### LLM Provider Status

```bash
curl https://scouter.dev.local/api/health/llm --insecure
# Returns status of all configured LLM providers (Ollama, Anthropic, etc.)
```

### Traefik Dashboard

Health of reverse proxy and routing:
```bash
http://localhost:8082/dashboard/
```

---

## Prometheus Configuration

See `monitoring/prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'scouter-backend'
    static_configs:
      - targets: ['backend:8080']
```

Alert rules in `monitoring/prometheus/rules/`:

```yaml
groups:
  - name: scouter
    rules:
      - alert: HighErrorRate
        expr: rate(scouter_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 2m

      - alert: LLMAllProvidersDown
        expr: count(scouter_llm_circuit_open_total) == count(count by (provider) (scouter_llm_requests_total))
        for: 1m

      - alert: SlowAPIResponse
        expr: histogram_quantile(0.95, scouter_http_request_duration_seconds_bucket) > 5
        for: 5m
```

---

## Logs

### Backend Logs

Backend uses structured slog JSON logging:

```bash
# View real-time backend logs
docker compose logs -f backend

# Filter to INFO level
docker compose logs -f backend | grep '"level":"INFO"'

# Sample output
{"time":"2026-03-15T10:00:00Z","level":"INFO","msg":"server started","port":8080}
{"time":"2026-03-15T10:00:01Z","level":"INFO","msg":"migrations applied","count":22}
{"time":"2026-03-15T10:01:00Z","level":"INFO","msg":"research completed","mission_id":"...","options":7,"duration_ms":3420}
```

### Traefik Logs

```bash
docker compose logs -f traefik
```

### Frontend Logs

Browser console (filter by `[SCOUTER]` prefix in DevTools).

---

## Debugging Monitoring Issues

### Prometheus can't scrape backend

Verify backend is healthy:
```bash
docker compose exec backend wget -O- http://localhost:8080/api/metrics
```

Check Prometheus scrape status:
```bash
curl http://localhost:9090/api/v1/targets
```

### Grafana dashboards not loading

Restart Grafana:
```bash
docker compose restart grafana
```

Check provisioning logs:
```bash
docker compose logs grafana | grep -i provisioning
```

### High memory usage

Check which container is consuming memory:
```bash
docker compose stats
```

Prometheus retention size is capped at 500MB (`storage.tsdb.retention.size` in docker-compose.yml).

---

## Disabling Monitoring

To free resources, omit the monitoring profile:

```bash
# Core only (no monitoring)
make up

# Stop monitoring services
docker compose --profile monitoring down
```

Or disable metrics collection:
```bash
# In .env
METRICS_ENABLED=false
```
