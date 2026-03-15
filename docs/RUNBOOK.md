# SCOUTER Universal — Runbook

Operational guide for running, monitoring, and troubleshooting SCOUTER in production and development.

---

## Table of Contents

1. [Startup & Shutdown](#startup--shutdown)
2. [Health Checks](#health-checks)
3. [Logging & Debugging](#logging--debugging)
4. [Database Operations](#database-operations)
5. [Backup & Recovery](#backup--recovery)
6. [Performance Tuning](#performance-tuning)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

---

## Startup & Shutdown

### Prerequisites (First Time Only)

Generate local HTTPS certificates:

```bash
make certs
```

On Windows (WSL2), install the CA certificate in Windows:
```powershell
# Run as Administrator
certutil -addstore -f "ROOT" certs/ca.crt
```

### Starting the Full Stack

```bash
# Docker Compose (core only: postgres + backend + frontend + traefik)
make up

# Or: with sample data
make up-seed

# Or: with monitoring (Prometheus + Grafana)
make up-monitoring

# Or: everything
make up-full
```

**Expected startup sequence:**
1. PostgreSQL initializes and becomes healthy (~5s)
2. Backend waits for DB → applies migrations → starts on `:8080` (internal) (~3s)
3. Frontend builds and starts on `:80` (internal) (~10s)
4. Traefik starts and begins routing HTTPS to services (~2s)

**Services and URLs:**

| Service | URL | Port (host) |
|---------|-----|-----|
| SCOUTER App | https://scouter.dev.local | 443 |
| Backend API | https://scouter.dev.local/api | 443 |
| Traefik Dashboard | http://localhost:8082 | 8082 |
| PostgreSQL | localhost:5432 | 5432 |
| Prometheus | http://localhost:9090 | 9090 (with `--profile monitoring`) |
| Grafana | http://localhost:3000 | 3000 (with `--profile monitoring`) |

### Stopping Services

```bash
# Stop all containers (keep database)
make down

# Stop and delete all volumes (WARNING: destroys database)
make clean-volumes

# View status
make ps
```

### Local Development (Hot Reload, Without Traefik)

For development with live reload (no HTTPS):

```bash
# Terminal 1: Start PostgreSQL
docker compose up postgres -d

# Terminal 2: Start backend
cd backend
go run ./cmd/server
# Listens on http://localhost:8080

# Terminal 3: Start frontend (Vite dev server)
cd frontend
npm install
npm run dev
# Listens on http://localhost:5173 with hot reload
```

Set `ENV=development` in `.env` for permissive CORS.

**Access:** http://localhost:5173 (frontend dev server, not via Traefik)

---

## Health Checks

### System Health (via Traefik HTTPS)

```bash
# With self-signed certificate (ignore warning)
curl https://scouter.dev.local/api/health --insecure
```

**Response (OK):**
```json
{
  "status": "ok",
  "db": "connected",
  "version": "0.1.0"
}
```

**Response (Degraded):**
```json
{
  "status": "degraded",
  "db": "disconnected"
}
```

### System Health (Direct, Internal)

If backend is running locally on :8080:

```bash
curl http://localhost:8080/api/health
```

### LLM Provider Status (via Traefik)

```bash
curl https://scouter.dev.local/api/health/llm --insecure
```

**Response:**
```json
{
  "healthy": true,
  "providers": [
    { "name": "ollama-heavy", "status": "ok", "circuit": "closed" },
    { "name": "ollama-fast", "status": "ok", "circuit": "closed" },
    { "name": "anthropic", "status": "ok", "circuit": "closed" }
  ]
}
```

### Traefik Health

Check reverse proxy routing:

```bash
# Traefik dashboard
http://localhost:8082/dashboard/

# Check routers
curl http://localhost:8082/api/http/routers

# Check services
curl http://localhost:8082/api/http/services
```

### Service Logs

```bash
# Backend logs
docker compose logs backend -f

# Frontend logs
docker compose logs frontend -f

# Database logs
docker compose logs postgres -f

# All logs
docker compose logs -f
```

### Container Status

```bash
docker compose ps

# Expected output:
# NAME         COMMAND                 STATE      PORTS
# postgres     postgres -c fsync=off   Up (healthy)
# backend      ./server               Up
# frontend     nginx -g daemon off    Up
```

---

## Logging & Debugging

### Backend Logging

Backend uses structured JSON logging with `log/slog`:

```bash
docker compose logs backend | grep -i error
docker compose logs backend | jq .level   # Filter by log level
```

**Log levels:** DEBUG, INFO, WARN, ERROR

### Frontend Debugging

```bash
# Browser DevTools
# 1. Open http://localhost:5173
# 2. Press F12 for Developer Tools
# 3. Console tab shows JavaScript errors
# 4. Network tab shows API calls

# Or check browser console
docker compose logs frontend | tail -50
```

### Increase Verbosity

**Backend:**
```bash
# In .env, add:
# LOG_LEVEL=debug
docker compose restart backend
```

**Frontend:**
```bash
# In browser console:
localStorage.debug = '*'
# Reload page
```

---

## Database Operations

### Database Connection

```bash
# Interactive psql session
docker compose exec postgres psql -U scouter -d scouter

# Quick query
docker compose exec postgres psql -U scouter -d scouter -c "SELECT COUNT(*) FROM missions;"
```

### Migrations

```bash
# Check migration status
docker compose logs backend | grep -i "migration"

# Manually apply migrations
docker compose exec backend ./server

# Rollback last migration (development only)
make migrate-down
cd backend && go run cmd/migrate/main.go down 1

# Create new migration
# 1. Create files: backend/migrations/NNN_description.up.sql and .down.sql
# 2. Restart backend: docker compose restart backend
```

### Common Queries

```bash
# Count missions
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT COUNT(*) FROM missions;"

# List active missions
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT id, slug, name, status FROM missions WHERE status = 'active';"

# Check option count
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT COUNT(*) FROM options;"

# Check shopping items with price history
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT id, name, merchant, price FROM shopping_items LIMIT 10;"
```

### Backup Database

```bash
# Manual backup
docker compose exec postgres pg_dump -U scouter scouter > backup-$(date +%Y%m%d).sql

# Verify backup
file backup-*.sql
wc -l backup-*.sql
```

### Restore from Backup

```bash
# Restore from backup file
docker compose exec -T postgres psql -U scouter scouter < backup-20260315.sql
```

---

## Backup & Recovery

### Backup Strategy

**Daily Backups (recommended):**

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/scouter_$TIMESTAMP.sql"

docker compose exec -T postgres pg_dump -U scouter scouter > "$BACKUP_FILE"
gzip "$BACKUP_FILE"

# Keep only last 7 days
find "$BACKUP_DIR" -name "scouter_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

Run via cron:

```bash
# 2 AM daily
0 2 * * * cd /path/to/scouter && ./backup.sh
```

### Recovery Procedure

```bash
# 1. Stop backend to prevent conflicts
docker compose stop backend

# 2. List available backups
ls -lh backup-*.sql.gz

# 3. Restore from backup
gunzip -c backup-20260315.sql.gz | docker compose exec -T postgres psql -U scouter scouter

# 4. Restart backend
docker compose up backend

# 5. Verify health
curl http://localhost:8080/api/health
```

---

## Performance Tuning

### PostgreSQL Tuning

Edit `docker-compose.yml` and add PostgreSQL parameters:

```yaml
postgres:
  environment:
    POSTGRES_INIT_ARGS: "-c shared_buffers=256MB -c effective_cache_size=1GB -c work_mem=64MB"
```

**Key parameters:**
- `shared_buffers` — Buffer pool (default: 128MB, increase to 1/4 available RAM)
- `effective_cache_size` — Total cache estimate (1/2 available RAM)
- `work_mem` — Per-operation memory (default: 4MB)

### pgvector Index Tuning

```bash
# Check index size
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT schemaname, tablename, indexname, pg_size_pretty(pg_relation_size(indexrelid)) \
   FROM pg_indexes WHERE tablename = 'options' AND indexname LIKE '%vector%';"

# Reindex if needed
docker compose exec postgres psql -U scouter -d scouter -c \
  "REINDEX INDEX options_embedding_idx;"
```

### Backend Connection Pooling

Check `internal/db/pool.go` for pgx pool settings. Default:

```go
connPool := pgxpool.NewConfig()
connPool.MaxConns = 25        // max connections
connPool.MinConns = 5         // min connections
connPool.MaxConnIdleTime = 30 * time.Minute
```

Increase `MaxConns` under high load:

```go
connPool.MaxConns = 50  // for 100+ concurrent users
```

### Cache Configuration

The project uses in-memory cache with TTLs:

| Endpoint | Cache TTL | Size |
|----------|-----------|------|
| `/api/missions/:id/french-benchmark` | 20 min | ~1 KB |
| `/api/missions/:id/scorecard` | 30 min | ~2 KB |
| `/api/missions/:id/purchase-timeline` | 20 min | ~1 KB |
| `/api/wishlist/prioritized` | 15 min | ~5 KB |

To disable caching (development):

```bash
# In .env
CACHE_ENABLED=false
```

---

## Monitoring

### Prometheus Metrics

Enable metrics in `.env`:

```bash
METRICS_ENABLED=true
```

Access metrics at `http://localhost:8080/metrics` (Prometheus format).

**Key metrics:**

```
# Request latency
backend_request_duration_seconds_bucket{method="POST",path="/api/missions/:id/research"}

# LLM call latency
llm_request_duration_seconds{provider="ollama-heavy",model="qwen3:14b"}

# Database query latency
db_query_duration_seconds{query="get_mission_by_slug"}

# Cache hit rate
cache_hits_total / (cache_hits_total + cache_misses_total)
```

### Grafana Dashboards

```bash
docker compose --profile monitoring up
```

**Pre-built dashboards:**
- SCOUTER System Overview
- LLM Provider Status
- Database Performance
- Request Latency

Access at `http://localhost:3000` (admin/admin).

### Alerts (Optional)

Set up alerts in `monitoring/prometheus.yml`:

```yaml
groups:
  - name: scouter
    rules:
      - alert: HighErrorRate
        expr: rate(backend_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: LLMProviderDown
        expr: llm_provider_status{status="unhealthy"} == 1
        for: 1m
        annotations:
          summary: "LLM provider is down"
```

---

## Troubleshooting

### Backend won't start: "DATABASE_URL required"

```bash
# Fix: Set DATABASE_URL in .env
DATABASE_URL=postgres://scouter:scouter@postgres:5432/scouter

# Restart
docker compose up backend
```

### Backend fails: "connection refused"

PostgreSQL may not be ready. Logs show:

```
Error connecting to database: connection refused
```

**Fix:**

```bash
# Wait for PostgreSQL to become healthy
docker compose up postgres
# Wait ~10 seconds for "database system is ready"

# Then start backend
docker compose up backend
```

### Research returns no results

LLM provider may be misconfigured.

```bash
# Check provider status
curl http://localhost:8080/api/health/llm

# Example response:
# {"healthy": false, "providers": [{"name": "ollama-heavy", "status": "error"}]}

# If using Ollama:
# 1. Verify Ollama is running: curl http://localhost:11434/api/tags
# 2. Pull model if missing: ollama pull qwen3:14b
# 3. Check logs: docker compose logs backend | grep ollama
```

### Frontend can't reach backend (CORS error)

```
Access to XMLHttpRequest blocked by CORS policy
```

**Fix:**

```bash
# Set development mode in .env
ENV=development

# Restart backend
docker compose restart backend

# Or manually handle CORS in backend by editing internal/httputil/cors.go
```

### Migrations fail

```
Migration failed: schema version out of sync
```

**Fix:**

```bash
# 1. Check migration status
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT version, dirty FROM schema_migrations;"

# 2. If dirty, rollback manually
docker compose exec postgres psql -U scouter -d scouter -c \
  "UPDATE schema_migrations SET dirty = false WHERE version = (SELECT MAX(version) FROM schema_migrations);"

# 3. Restart backend
docker compose up backend
```

### Port 5173 or 8080 already in use

```bash
# Find process using port
lsof -i :5173
lsof -i :8080

# Kill process or use different port
docker compose down
# Edit docker-compose.yml ports and retry
```

### High memory usage

```bash
# Check container memory
docker stats

# If backend is high:
# - Increase Go garbage collection: GOGC=80
# - Reduce pgx connection pool: connPool.MaxConns

# If PostgreSQL is high:
# - Increase shared_buffers in docker-compose.yml
# - Run VACUUM ANALYZE to optimize
```

### Slow queries

```bash
# Enable query logging (PostgreSQL)
docker compose exec postgres psql -U scouter -d scouter -c \
  "ALTER SYSTEM SET log_min_duration_statement = 1000;"  # Log queries >1s

docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT pg_reload_conf();"

# Check slow logs
docker compose logs postgres | grep "duration:"

# Analyze query plan
docker compose exec postgres psql -U scouter -d scouter -c \
  "EXPLAIN ANALYZE SELECT * FROM missions WHERE status = 'active';"
```

### LLM provider circuit breaker open

Backend uses circuit breakers to handle failing LLM providers.

```bash
# Check status
curl http://localhost:8080/api/health/llm

# If circuit is "open":
# 1. Check logs: docker compose logs backend | grep circuit
# 2. Restart the failing provider (Ollama)
# 3. Circuit automatically closes after timeout (default: 30s)
```

### Semantic search not working

pgvector embeddings may not be indexed.

```bash
# Check embeddings
docker compose exec postgres psql -U scouter -d scouter -c \
  "SELECT COUNT(*) FROM options WHERE embedding IS NOT NULL;"

# Reindex embeddings
curl -X POST http://localhost:8080/api/search/reindex

# Check logs
docker compose logs backend | grep -i "reindex\|embed"
```

---

## Emergency Procedures

### Service Recovery

```bash
# 1. Stop all services
docker compose down

# 2. Check for stuck processes
docker ps -a | grep scouter

# 3. Restart cleanly
docker compose up --build
```

### Database Rollback

```bash
# Rollback one migration
make migrate-down

# Or manually:
cd backend && go run cmd/migrate/main.go down 1
docker compose up backend
```

### Complete Data Wipe (DANGER)

```bash
# Delete all data
curl -X DELETE http://localhost:8080/api/data \
  -H "X-Confirm: yes"

# Or via Docker
docker compose down -v
docker compose up postgres
```

---

## Support

For detailed troubleshooting, see:
- [Contributing Guide](development/contributing.md) — Dev setup issues
- [API Reference](technical/api.md) — Endpoint-specific problems
- [Environment Variables](deployment/environment.md) — Config reference

For bugs and feature requests, open a GitHub issue.
