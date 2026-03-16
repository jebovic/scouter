# Async Research Jobs — Design Spec

**Date:** 2026-03-16
**Status:** Approved
**Feature:** Non-blocking research execution with job tracking

---

## Problem

When a user triggers research on the Options screen, the UI freezes for up to 60 seconds while the LLM call completes. The HTTP request blocks the browser, there is no progress feedback, and the timeout ceiling is artificially constrained by the HTTP request lifecycle.

---

## Goals

- Research runs asynchronously — the UI never blocks
- Research jobs are visible as a list (ongoing + last run) directly in the Options screen
- Jobs survive page refreshes (persisted to DB)
- The LLM timeout can be extended to 5+ minutes without user impact
- One job at a time per mission (concurrent runs blocked)

---

## Non-Goals

- WebSocket or SSE real-time push (polling is sufficient)
- Job cancellation
- Cross-mission job queue or priority system
- Pricing agent async (research only, for now)
- Per-route rate limiting on the trigger endpoint (future work)

---

## Architecture

### Backend

**Execution model:** Goroutine-per-job. When a research job is created, a goroutine is spawned and detached from the HTTP request context. The goroutine uses a fresh `context.Background()` as root, so it is not bounded by any HTTP deadline.

**Status ownership:** The handler inserts the job with `status='pending'` and returns immediately. The goroutine transitions the job to `running` as its first step (updating `started_at`), then to `done` or `failed` when complete. This ensures the 202 response always reflects the actual DB state at the time of return.

**Startup cleanup:** On server start, before routes are registered, any job with `status IN ('running', 'pending')` is reset to `failed` with `error = 'server restarted'`. This prevents stale rows from blocking future jobs. If the cleanup query itself fails, log the error and continue — it is non-fatal.

**LLM timeout:** The agent's internal LLM context timeout is bumped from 60 seconds to 5 minutes in `internal/research/agent.go`. The goroutine does **not** add its own outer timeout — the agent owns the single timeout layer.

**Panic recovery:** `runResearchJob` must contain a `defer func() { if r := recover(); r != nil { /* UPDATE job status=failed, error="internal panic" */ } }()` as its first statement. A bare goroutine panic crashes the entire server process.

**Mission deletion:** `research_jobs.mission_id` has `ON DELETE CASCADE`. If a mission is deleted while a job is running, the job row is deleted. The goroutine's subsequent `UPDATE research_jobs SET status=... WHERE id=$1` will update 0 rows — this is acceptable and logged as a warning. The agent's DB writes (`DeleteByMission`, `Create`) will return `pgx.ErrNoRows` or FK errors, which the goroutine treats as a standard failure path.

---

## Data Model

New migration: `025_research_jobs`

```sql
CREATE TABLE research_jobs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id    UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  status        TEXT NOT NULL DEFAULT 'pending',  -- pending | running | done | failed
  feedback      TEXT,
  error         TEXT,
  options_count INT,
  started_at    TIMESTAMPTZ,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON research_jobs(mission_id, created_at DESC);
```

**Status lifecycle:** `pending` → `running` → `done` | `failed`

The existing `agent_runs` table is unchanged — the research agent continues writing audit snapshots as before. After a rollback (down migration drops `research_jobs`), `agent_runs` rows from async jobs become orphans — acceptable since there is no FK between the two tables.

---

## API

### Trigger research (modified)

```
POST /api/missions/:id/research
Body: { feedback?: string }

Response 409: { error: "research already running" }   — if status=running|pending exists
Response 202: { job_id: UUID, status: "pending" }
```

The endpoint no longer blocks. It inserts the job row, spawns the goroutine, and returns immediately with `status: "pending"`.

### List jobs

```
GET /api/missions/:id/research/jobs
Response 200: {
  jobs: [{
    id: UUID,
    status: "pending" | "running" | "done" | "failed",
    options_count: int | null,
    error: string | null,
    feedback: string | null,
    started_at: string | null,
    completed_at: string | null,
    created_at: string
  }]
}
```

Returns the last 10 jobs for the mission, newest first. Used for both history display and status polling.

### Get single job

```
GET /api/missions/:id/research/jobs/:job_id
Response 200: {
  id: UUID,
  status: "pending" | "running" | "done" | "failed",
  options_count: int | null,
  error: string | null,
  feedback: string | null,
  started_at: string | null,
  completed_at: string | null,
  created_at: string
}
```

Used by the frontend polling loop to check a specific in-progress job.

---

## Backend Implementation

### New package: `internal/researchjob/`

- `model.go` — `ResearchJob` struct
- `repository.go` — `Create`, `UpdateStatus`, `GetByMission`, `GetByID`, `FailStaleJobs`
- `handler.go` — HTTP handlers for the 3 endpoints above
- `runner.go` — `runResearchJob(jobID, missionID, feedback string, agent, repo, log)` goroutine function

### Handler flow (POST /research)

```
1. Parse missionID from URL
2. Load mission from DB (404 if not found)
3. Check for active job:
     SELECT id FROM research_jobs WHERE mission_id=$1 AND status IN ('pending','running') LIMIT 1
   → 409 { error: "research already running" } if found
4. Decode optional feedback from request body
5. INSERT INTO research_jobs (mission_id, feedback) returning full row → job (status='pending')
6. go runResearchJob(job.ID, mission, job.Feedback, a.agent, a.repo, a.logger)
7. WriteJSON 202 { job_id: job.ID, status: "pending" }
```

### Goroutine function

```
runResearchJob(jobID, mission, feedback, agent, repo, logger):
  defer func() {
    if r := recover(); r != nil {
      repo.UpdateStatus(context.Background(), jobID, "failed", "internal panic", 0)
      logger.Printf("panic in research job %s: %v", jobID, r)
    }
  }()

  // Transition to running
  repo.UpdateStatus(context.Background(), jobID, "running", "", 0)  // sets started_at

  // Run agent (owns the 5-minute LLM timeout internally)
  result, err := agent.Run(context.Background(), mission, feedbackInputOrNil(feedback))

  if err != nil {
    repo.UpdateStatus(context.Background(), jobID, "failed", err.Error(), 0)
    return
  }

  // UpdateStatus returns (rowsAffected int64, err error)
  // If rowsAffected == 0, the mission was deleted (ON DELETE CASCADE removed the job row)
  // Log as warning and return — this is not an error condition
  n, err := repo.UpdateStatus(context.Background(), jobID, "done", "", len(result.Options))
  if err != nil {
    logger.Printf("warn: failed to update research job %s to done: %v", jobID, err)
  } else if n == 0 {
    logger.Printf("warn: research job %s not found after completion (mission deleted?)", jobID)
  }
```

`feedbackInputOrNil(s string) *FeedbackInput`: returns `&FeedbackInput{Feedback: s}` if `s != ""`, else `nil`.

### Routing

The existing `researchHandler` handles `POST /api/missions/{missionID}/research`. The new `researchJobHandler` registers on the same mission subrouter:

```
POST   /api/missions/{missionID}/research           → researchJobHandler.Trigger  (replaces old handler)
GET    /api/missions/{missionID}/research/jobs      → researchJobHandler.List
GET    /api/missions/{missionID}/research/jobs/{jobID} → researchJobHandler.Get
```

The old `research.Handler` `Trigger` method is removed; the research agent is injected into `researchjob.Handler` instead.

### Startup cleanup (routes.go)

After DB pool is ready, before routes are registered:

```go
if err := researchJobRepo.FailStaleJobs(context.Background()); err != nil {
    log.Printf("warn: failed to clean stale research jobs: %v", err)
}
```

---

## Frontend Implementation

### New hook: `useResearchJobs(missionId)`

- Fetches `GET /api/missions/:id/research/jobs`
- **Polling:** `refetchInterval: (query) => query.data?.jobs.some(j => j.status === 'running' || j.status === 'pending') ? 3000 : false`
- **Transition detection:** `useRef<Record<string, string>>({})` keyed by `job.id → previous status`. On each poll, compare each job's current status to its previous. On `running|pending → done` transition: invalidate `['options', missionId]` + show success toast. On `→ failed` transition: show error toast. Update the ref after processing.
- **Unmount:** TanStack Query's built-in cleanup stops the refetch interval automatically when the component unmounts and no other subscribers hold the query. No manual `useEffect` cleanup is required.

### Modified hook: `useResearch.ts` → rename export to `useTriggerResearch`

The file `src/hooks/useResearch.ts` is updated in place; the exported hook is renamed from `useResearch` to `useTriggerResearch`. All existing consumers (`OptionsExplorer.tsx`) are updated to use the new name.

- `POST /api/missions/:id/research` with optional feedback
- Receives `{ job_id, status: "pending" }` immediately
- On success: invalidates `['research-jobs', missionId]` — the polling hook picks up the new running job on next refetch
- On 409: shows "Research is already running" toast
- No longer awaits completion

### UI changes in `OptionsExplorer.tsx`

**Status bar — last run exists, idle:**
```
Last run: Mar 16 14:22 · 4 options found · show history ↓
```

**Last run was a failure:**
```
Last run: Mar 16 14:22 · Failed · show history ↓
```

**Job running:**
```
[⚡ Run Research — disabled]   🔄 Research running…
Last run: Mar 16 14:22 · 4 options found · show history ↓
```

**"show history" expands inline list:**
```
✓ Mar 16 14:22 · 4 options
✓ Mar 15 09:11 · 3 options
✗ Mar 14 22:03 · Failed: timeout
```

**On job done:** options list auto-refreshes (cache invalidation), toast "Research complete — X options found".
**On job failed:** toast "Research failed: [error message]", button re-enabled.

### i18n keys (en.json / fr.json)

New keys under the existing `"research"` namespace (verify no collision with existing keys):

```json
"research": {
  "runButton": "Run Research",
  "running": "Research running…",
  "lastRunSuccess": "Last run: {{date}} · {{count}} options found",
  "lastRunFailed": "Last run: {{date}} · Failed",
  "showHistory": "show history",
  "hideHistory": "hide history",
  "complete": "Research complete — {{count}} options found",
  "failed": "Research failed: {{error}}",
  "alreadyRunning": "Research is already running",
  "historyStatus": {
    "done": "✓",
    "failed": "✗",
    "running": "🔄",
    "pending": "⏳"
  }
}
```

---

## Migration

- Up: `backend/internal/db/migrations/025_research_jobs.up.sql`
- Down: `backend/internal/db/migrations/025_research_jobs.down.sql` → `DROP TABLE research_jobs;`

---

## Testing

### Backend
- Unit tests for `researchjob` repository: `Create`, `UpdateStatus`, `GetByMission`, `FailStaleJobs`
- Handler tests: 202 on trigger, 409 on concurrent, 200 on list/get
- Goroutine integration test: mock agent completes → job status flips to `done`
- Panic recovery test: mock agent panics → job status flips to `failed`, process survives

### Frontend
- `useResearchJobs`: polling interval active when job running, stops when done, options cache invalidated on done, toasts fire once per transition
- `useTriggerResearch`: 202 path invalidates research-jobs query, 409 shows toast
- `OptionsExplorer`: button disabled when job running, status bar shows correct copy for success/failed last run, history toggle works

---

## File Changes Summary

**Backend (new):**
- `internal/researchjob/model.go`
- `internal/researchjob/repository.go`
- `internal/researchjob/handler.go`
- `internal/researchjob/runner.go`
- `internal/db/migrations/025_research_jobs.up.sql`
- `internal/db/migrations/025_research_jobs.down.sql`

**Backend (modified):**
- `internal/research/agent.go` — bump LLM timeout from 60s to 5 min
- `cmd/server/routes.go` — replace old research trigger route, register 3 researchjob routes, add stale job cleanup, wire `researchJobRepo`/`researchJobHandler` deps
- `cmd/server/main.go` — instantiate `researchjob.Repository` and `researchjob.Handler`

**Frontend (new):**
- `src/hooks/useResearchJobs.ts`
- `src/api/researchJobs.ts` — Zod schemas + fetch wrappers for all 3 endpoints

**Frontend (modified):**
- `src/hooks/useResearch.ts` — rename export to `useTriggerResearch`, replace blocking await with fire-and-forget
- `src/pages/OptionsExplorer.tsx` — wire `useResearchJobs`, status bar, history toggle, update `useResearch` → `useTriggerResearch`
- `src/pages/OptionsExplorer.module.css` — status bar and history list styles
- `src/i18n/en.json`, `src/i18n/fr.json` — new research job keys
