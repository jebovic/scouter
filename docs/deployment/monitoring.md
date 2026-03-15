# Monitoring

SCOUTER ships with a full observability stack (Phase 14): Prometheus, Grafana, and cAdvisor.

---

## Starting the Monitoring Stack

```bash
docker compose --profile monitoring up
```

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |
| cAdvisor | http://localhost:8081 | — |
| SCOUTER metrics | http://localhost:8080/metrics | — |

Enable metrics collection in your `.env`:

```bash
METRICS_ENABLED=true
```

---

## Metrics Collected

### HTTP Layer (chi middleware)

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_http_requests_total` | Counter | `method`, `route`, `status` |
| `scouter_http_request_duration_seconds` | Histogram | `method`, `route` |
| `scouter_http_request_size_bytes` | Histogram | `method`, `route` |

### LLM Layer (SmartRouter)

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

### Scheduler

| Metric | Type | Labels |
|--------|------|--------|
| `scouter_price_checks_total` | Counter | `status` |
| `scouter_alerts_fired_total` | Counter | — |

---

## Grafana Dashboards

Pre-provisioned dashboards (loaded automatically on startup):

### SCOUTER Overview

- Request rate (req/s)
- p50 / p95 / p99 latency
- Error rate by endpoint
- Active missions count

### LLM Intelligence

- Requests per provider (stacked area)
- LLM latency by model
- Cascade events (when SmartRouter falls back)
- Circuit breaker state timeline

### Agent Performance

- Top 10 agents by call volume
- Agent error rate
- Embedding worker queue depth

### Infrastructure

- CPU usage per container (cAdvisor)
- Memory usage per container
- Disk I/O (PostgreSQL)
- Network in/out

---

## Alerting

Prometheus alert rules (in `monitoring/alerts.yml`):

```yaml
groups:
  - name: scouter
    rules:
      - alert: HighErrorRate
        expr: rate(scouter_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Error rate > 5% for 2 minutes"

      - alert: LLMAllProvidersDown
        expr: scouter_llm_circuit_open_total == count(up{job="scouter-backend"})
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "All LLM providers have open circuits"

      - alert: SlowAPIResponse
        expr: histogram_quantile(0.95, scouter_http_request_duration_seconds_bucket) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "p95 latency > 5s for 5 minutes"
```

---

## Health Checks

### Backend Health

```bash
curl http://localhost:8080/api/health
# { "status": "ok", "db": "connected" }
```

### LLM Health

```bash
curl http://localhost:8080/api/health/llm
# { "healthy": true, "providers": [...] }
```

### Docker Health Checks

```yaml
# docker-compose.yml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
  interval: 10s
  timeout: 5s
  retries: 3
  start_period: 30s
```

---

## Logs

Backend uses structured slog JSON logging:

```bash
# View logs
docker compose logs -f backend

# Sample output
{"time":"2026-03-15T10:00:00Z","level":"INFO","msg":"server started","port":8080}
{"time":"2026-03-15T10:00:01Z","level":"INFO","msg":"migrations applied","count":22}
{"time":"2026-03-15T10:01:00Z","level":"INFO","msg":"research completed","mission_id":"...","options":7,"duration_ms":3420}
```

Frontend logs go to browser console. Filter by component using the `[SCOUTER]` prefix.
