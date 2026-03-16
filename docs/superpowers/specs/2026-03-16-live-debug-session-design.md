# Live Debug Session — Design Spec

**Date:** 2026-03-16
**Project:** scouter
**Status:** Approved

---

## Overview

A structured live debug workflow where the user navigates the app while Claude captures frontend console logs (via Playwright) and backend container logs (via Docker), then produces a prioritized remediation plan.

---

## Architecture & Flow

### Phase 1 — Automated Sweep (~2 min)

Playwright visits every known route in the app automatically:

- `/` — HQ Dashboard
- `/missions/:slug` — Mission Overview (first available mission)
- `/missions/:slug/options` — Options Explorer
- `/missions/:slug/shopping` — Shopping Tracker
- `/search` — Semantic Search
- `/history` — Purchase History
- `/stats` — Stats Page
- `/settings` — Settings

For each route, Playwright captures:
- `console.error`, `console.warn`, `console.log` events
- Uncaught JS exceptions (`pageerror`)
- HTTP 4xx/5xx responses
- Failed requests (DNS, timeout, CORS)

Backend logs are pulled once at the end of the sweep using `--since <T0>`.

Output: a baseline error map per route — what's broken before any user interaction.

### Phase 2 — Guided Reproduction

The user describes specific flows to reproduce (e.g. "add item to shopping list, then check deal score"). Claude executes each step in Playwright, pulls backend log deltas between actions (using `--since <last_pull>`), and annotates what changed.

Repeat until all known issues are reproduced or exhausted.

### Final Output — Remediation Plan

Findings consolidated into a dated markdown report at:
`docs/superpowers/debug-sessions/YYYY-MM-DD-debug-report.md`

---

## Log Capture Mechanics

### Frontend (Playwright)

```
page.on('console', msg)      → all console levels with timestamp + text
page.on('pageerror', err)    → uncaught JS exceptions with stack trace
page.on('response', res)     → filter 4xx/5xx, log method + path + status
page.on('requestfailed', req) → failed requests with failure reason
```

Each event stamped: `{ timestamp, url, level, message }`

### Backend (Docker)

```bash
docker compose -f deployment/docker-compose.yml logs backend \
  --since <T0> --no-log-prefix
```

Pulled as delta after each Phase 2 action. Backend emits structured JSON via `log/slog` — parseable by timestamp, level, and message fields.

### Correlation

Frontend network errors are correlated with backend log entries within a ±2s timestamp window, matched by HTTP method + path.

### TLS / Domain

- Target: `https://scouter.dev.local`
- Playwright uses the OS CA store (Traefik's self-signed CA installed via `make certs`)
- `--ignore-https-errors` is NOT used — real cert validation enforced

---

## Remediation Plan Output

File: `docs/superpowers/debug-sessions/YYYY-MM-DD-debug-report.md`

### Structure

```markdown
# Debug Session — YYYY-MM-DD

## Executive Summary
- Total issues: N (critical/high/medium/low breakdown)
- Pages with errors: [list]
- Backend error rate observed

## Findings

### [SEVERITY] Title
- **Route:** /path
- **Symptom:** description
- **Backend log:** correlated slog entry (if matched)
- **Root cause hypothesis:** explanation
- **Fix:** specific file:line recommendation

## Raw Logs
### Frontend Console Events (by route)
### Backend Log Delta (full slog output from session)
```

### Severity Classification

| Level | Criteria |
|-------|----------|
| `CRITICAL` | JS exceptions breaking a page, 5xx responses, auth failures |
| `HIGH` | 4xx on data-loading requests, console.errors with stack traces |
| `MEDIUM` | console.warns, non-blocking failed requests |
| `LOW` | Missing data hints in logs, slow responses >3s |

---

## Constraints & Prerequisites

- Docker Compose stack must be running (`make up`) before starting
- `make certs` CA must be installed in system trust store
- Playwright MCP available in the Claude Code session
- At least one mission with data seeded (for route coverage)
