# Live Debug Session Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute a two-phase live debug session against `https://scouter.dev.local` — automated Playwright sweep + guided reproduction — and produce a prioritized remediation plan.

**Architecture:** Phase 1 uses Playwright MCP to visit all known routes and capture console/network events. Phase 2 uses targeted Playwright interactions to reproduce specific user-reported issues while pulling Docker log deltas. All findings are correlated and written to a dated debug report.

**Tech Stack:** Playwright MCP, Docker Compose (`deployment/docker-compose.yml`), Go slog JSON logs, `https://scouter.dev.local` (Traefik + OS CA cert)

**Spec:** `docs/superpowers/specs/2026-03-16-live-debug-session-design.md`

---

## Chunk 1: Prerequisites Verification

### Task 1: Verify Docker Stack Is Running

**Files:** none

- [ ] **Step 1: Check container status**

```bash
docker compose -f deployment/docker-compose.yml ps --format "table {{.Name}}\t{{.Status}}"
```

Expected: `postgres`, `backend`, `frontend`, `traefik` all showing `Up` or `running`.
If any container is not running → run `make up` and wait for healthy status before continuing.

- [ ] **Step 2: Verify HTTPS endpoint responds**

```bash
curl -sf https://scouter.dev.local/api/health | head -c 200
```

Expected: JSON with `"status":"ok"` or `"status":"degraded"`.
If curl fails with SSL error → the OS CA is not trusted. Run `make certs` and reinstall the CA, then retry.
If curl fails with connection refused → Traefik is not routing correctly. Check `docker compose logs traefik`.

- [ ] **Step 3: Check at least one mission exists**

```bash
curl -sf https://scouter.dev.local/api/missions | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['results'][0]['slug'] if d.get('results') else 'NO_MISSIONS')"
```

Expected: a slug string like `macbook-pro-2024`.
If `NO_MISSIONS` → run `make up-seed` to seed sample data, then re-run this step.
Save the slug — it will be used as `MISSION_SLUG` in all subsequent Playwright steps.

---

## Chunk 2: Phase 1 — Automated Sweep

Record the sweep start time before the first Playwright navigation. This is `T0` for the Docker log pull at the end of Phase 1.

```bash
T0=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && echo "Sweep start: $T0"
```

### Task 2: Sweep — HQ Dashboard (`/`)

**Files:** none (runtime capture only)

- [ ] **Step 1: Navigate to root and capture**

Using Playwright MCP:
1. Open `https://scouter.dev.local/`
2. Wait for network idle (or 3s)
3. Take a screenshot for visual confirmation
4. Collect all console events and page errors captured since navigation

Record findings in this format:
```
ROUTE: /
  CONSOLE_ERRORS: [list or "none"]
  CONSOLE_WARNS: [list or "none"]
  PAGE_ERRORS: [list or "none"]
  FAILED_REQUESTS: [list method+path+status or "none"]
  SLOW_RESPONSES: [requests >3s or "none"]
```

### Task 3: Sweep — Mission Overview (`/missions/:slug`)

- [ ] **Step 1: Navigate and capture**

Using Playwright MCP:
1. Open `https://scouter.dev.local/missions/<MISSION_SLUG>`
2. Wait for network idle (or 3s)
3. Take a screenshot
4. Collect console/network events

Record findings (same format as Task 2).

### Task 4: Sweep — Options Explorer (`/missions/:slug/options`)

- [ ] **Step 1: Navigate and capture**

Using Playwright MCP:
1. Open `https://scouter.dev.local/missions/<MISSION_SLUG>/options`
2. Wait for network idle (or 3s)
3. Take a screenshot
4. Collect console/network events

Record findings.

### Task 5: Sweep — Shopping Tracker (`/missions/:slug/shopping`)

- [ ] **Step 1: Navigate and capture**

Using Playwright MCP:
1. Open `https://scouter.dev.local/missions/<MISSION_SLUG>/shopping`
2. Wait for network idle (or 3s)
3. Take a screenshot
4. Collect console/network events

Record findings.

### Task 6: Sweep — Search Page (`/search`)

- [ ] **Step 1: Navigate and capture**

Using Playwright MCP:
1. Open `https://scouter.dev.local/search`
2. Wait for network idle (or 3s)
3. Take a screenshot
4. Collect console/network events

Record findings.

### Task 7: Sweep — History, Stats, Settings

- [ ] **Step 1: Navigate `/history` and capture**

Using Playwright MCP: open `https://scouter.dev.local/history`, wait, screenshot, collect events.

- [ ] **Step 2: Navigate `/stats` and capture**

Using Playwright MCP: open `https://scouter.dev.local/stats`, wait, screenshot, collect events.

- [ ] **Step 3: Navigate `/settings` and capture**

Using Playwright MCP: open `https://scouter.dev.local/settings`, wait, screenshot, collect events.

Record findings for all three.

### Task 8: Pull Phase 1 Backend Logs

- [ ] **Step 1: Pull backend logs since T0**

```bash
docker compose -f deployment/docker-compose.yml logs backend --since "$T0" --no-log-prefix 2>&1
```

Save the full output as `BACKEND_LOGS_PHASE1`.

- [ ] **Step 2: Extract errors and warnings**

```bash
docker compose -f deployment/docker-compose.yml logs backend --since "$T0" --no-log-prefix 2>&1 \
  | grep -E '"level":"(ERROR|WARN)"'
```

Save as `BACKEND_ERRORS_PHASE1`.

### Task 9: Compile Phase 1 Baseline

- [ ] **Step 1: Correlate frontend + backend findings**

For each frontend network error found in Tasks 2–7:
- Look for a matching backend log entry within ±2s, same HTTP method + path
- Note whether the error originated backend-side (5xx) or frontend-side (request never reached backend)

- [ ] **Step 2: Summarize Phase 1 findings**

Produce a preliminary list:
```
PHASE 1 BASELINE:
  CRITICAL: [count] — [brief description of each]
  HIGH: [count] — [brief description]
  MEDIUM: [count] — [brief description]
  LOW: [count] — [brief description]
  ROUTES CLEAN: [list]
```

- [ ] **Step 3: Present baseline to user**

Show the summary and ask:
> "Phase 1 sweep complete. Here's what I found: [summary]. Do you have specific flows you want me to reproduce in Phase 2, or shall I move straight to the remediation report?"

---

## Chunk 3: Phase 2 — Guided Reproduction

Phase 2 runs once per user-described flow. Repeat Task 10 for each flow until the user says "done" or no more flows remain.

Record `LAST_PULL` time before each action — use it for the subsequent `--since` delta pull.

### Task 10: Execute a User-Described Flow

**Template — repeat for each flow:**

- [ ] **Step 1: Record action start time**

```bash
LAST_PULL=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && echo "Action start: $LAST_PULL"
```

- [ ] **Step 2: Execute the flow in Playwright**

Follow the user's description step by step. For each interaction (click, fill, navigate):
1. Perform the action in Playwright
2. Wait for network idle or response
3. Capture any console/network events triggered by this specific action

- [ ] **Step 3: Pull backend log delta**

```bash
docker compose -f deployment/docker-compose.yml logs backend --since "$LAST_PULL" --no-log-prefix 2>&1
```

Save output as `BACKEND_DELTA_<FLOW_NAME>`.

- [ ] **Step 4: Annotate what changed**

Compare the backend delta to the action taken. Note:
- Which API endpoints were called
- Any errors or warnings emitted by the backend
- Whether the frontend behavior matched expectation

- [ ] **Step 5: Ask for next flow**

> "Flow recorded. Any other flows to reproduce, or shall I write the remediation report?"

---

## Chunk 4: Remediation Report

### Task 11: Write the Debug Report

**Files:**
- Create: `docs/superpowers/debug-sessions/2026-03-16-debug-report.md`

- [ ] **Step 1: Create directory**

```bash
mkdir -p /home/jibei/projects/scouter/docs/superpowers/debug-sessions
```

- [ ] **Step 2: Write the report**

Create `docs/superpowers/debug-sessions/2026-03-16-debug-report.md` with this structure:

```markdown
# Debug Session — 2026-03-16

## Executive Summary
- Total issues: N (X critical, X high, X medium, X low)
- Pages with errors: [list]
- Backend errors observed: [count from slog]
- Phase 2 flows reproduced: [count]

## Findings

<!-- One section per finding, ordered CRITICAL → HIGH → MEDIUM → LOW -->

### [CRITICAL] <Title>
- **Route:** /path
- **Symptom:** exact console message or HTTP status
- **Triggered by:** user action (Phase 2) or page load (Phase 1)
- **Backend log:** `{ "time": "...", "level": "ERROR", "msg": "..." }` (if correlated)
- **Root cause hypothesis:** explanation
- **Fix:** `src/path/to/file.tsx:line` — what to change

...

## Raw Logs

### Frontend Console Events (by route)
<!-- paste captured events per route -->

### Backend Logs — Phase 1 Full Output
<!-- paste BACKEND_LOGS_PHASE1 -->

### Backend Logs — Phase 2 Deltas
<!-- paste each BACKEND_DELTA_<FLOW> with flow label -->
```

- [ ] **Step 3: Commit the report**

```bash
git add docs/superpowers/debug-sessions/2026-03-16-debug-report.md
git commit -m "docs: add live debug session report 2026-03-16"
```

- [ ] **Step 4: Present report to user**

Share the findings summary and top-priority fixes. Ask:
> "Report saved to `docs/superpowers/debug-sessions/2026-03-16-debug-report.md`. Here are the top issues: [CRITICAL/HIGH summary]. Want me to start fixing any of these now?"

---

## Severity Quick Reference

| Level | Criteria |
|-------|----------|
| `CRITICAL` | JS exceptions breaking a page, 5xx responses, auth failures |
| `HIGH` | 4xx on data-loading requests, `console.error` with stack trace |
| `MEDIUM` | `console.warn`, non-blocking failed requests |
| `LOW` | Missing data hints, slow responses >3s |

## Correlation Rule

Match frontend network error to backend log: same HTTP method + path, timestamps within ±2s.
If no backend log entry exists for a frontend 4xx/5xx → request likely never reached the backend (CORS, DNS, Traefik routing issue).
