# Async Research Jobs Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the blocking research HTTP call with a fire-and-forget job system so the UI never freezes while research runs, and show ongoing + past jobs inline in the Options screen.

**Architecture:** POST /research inserts a `research_jobs` row (status=pending) and immediately returns a job ID; a goroutine picks up the job, transitions it to running/done/failed, and the frontend polls GET /research/jobs every 3 seconds until completion, then auto-refreshes the options list.

**Tech Stack:** Go (chi, pgx/v5, golang-migrate), React 19, TanStack Query v5, Zod, react-i18next

**Spec:** `docs/superpowers/specs/2026-03-16-async-research-jobs-design.md`

---

## Chunk 1: Backend — migration, package, routes

### Task 1: DB migration

**Files:**
- Create: `backend/internal/db/migrations/025_research_jobs.up.sql`
- Create: `backend/internal/db/migrations/025_research_jobs.down.sql`

- [ ] **Step 1: Write up migration**

```sql
-- backend/internal/db/migrations/025_research_jobs.up.sql
CREATE TABLE research_jobs (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  mission_id    UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
  status        TEXT NOT NULL DEFAULT 'pending',
  feedback      TEXT,
  error         TEXT,
  options_count INT,
  started_at    TIMESTAMPTZ,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON research_jobs(mission_id, created_at DESC);
```

- [ ] **Step 2: Write down migration**

```sql
-- backend/internal/db/migrations/025_research_jobs.down.sql
DROP TABLE IF EXISTS research_jobs;
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/migrations/025_research_jobs.up.sql backend/internal/db/migrations/025_research_jobs.down.sql
git commit -m "feat(db): add research_jobs table migration 025"
```

---

### Task 2: researchjob model

**Files:**
- Create: `backend/internal/researchjob/model.go`

- [ ] **Step 1: Write model**

```go
// backend/internal/researchjob/model.go
package researchjob

import (
	"time"

	"github.com/google/uuid"
)

// Status values for ResearchJob.
const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

// ResearchJob tracks an async research execution for a mission.
type ResearchJob struct {
	ID           uuid.UUID  `json:"id"`
	MissionID    uuid.UUID  `json:"missionId"`
	Status       string     `json:"status"`
	Feedback     *string    `json:"feedback"`
	Error        *string    `json:"error"`
	OptionsCount *int       `json:"optionsCount"`
	StartedAt    *time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/researchjob/model.go
git commit -m "feat(researchjob): add ResearchJob model"
```

---

### Task 3: researchjob repository

**Files:**
- Create: `backend/internal/researchjob/repository.go`
- Create: `backend/internal/researchjob/repository_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// backend/internal/researchjob/repository_test.go
package researchjob_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/researchjob"
)

// These are integration tests — they require a real DB.
// Run with: cd backend && go test ./internal/researchjob/... -v -tags integration
// For unit testing the repository behaviour, we use a mock repo below.

// MockRepository implements Repository for unit tests.
type MockRepository struct {
	jobs         []researchjob.ResearchJob
	failStaleErr error
}

func (m *MockRepository) Create(ctx context.Context, missionID uuid.UUID, feedback string) (*researchjob.ResearchJob, error) {
	job := researchjob.ResearchJob{
		ID:        uuid.New(),
		MissionID: missionID,
		Status:    researchjob.StatusPending,
	}
	if feedback != "" {
		job.Feedback = &feedback
	}
	m.jobs = append(m.jobs, job)
	return &job, nil
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, optionsCount int) (int64, error) {
	for i, j := range m.jobs {
		if j.ID == id {
			m.jobs[i].Status = status
			return 1, nil
		}
	}
	return 0, nil // mission deleted — 0 rows
}

func (m *MockRepository) GetByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]researchjob.ResearchJob, error) {
	var out []researchjob.ResearchJob
	for _, j := range m.jobs {
		if j.MissionID == missionID {
			out = append(out, j)
		}
	}
	return out, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*researchjob.ResearchJob, error) {
	for _, j := range m.jobs {
		if j.ID == id {
			return &j, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) HasActiveJob(ctx context.Context, missionID uuid.UUID) (bool, error) {
	for _, j := range m.jobs {
		if j.MissionID == missionID && (j.Status == researchjob.StatusPending || j.Status == researchjob.StatusRunning) {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockRepository) FailStaleJobs(ctx context.Context) error {
	return m.failStaleErr
}

func TestMockRepository_Create(t *testing.T) {
	repo := &MockRepository{}
	missionID := uuid.New()

	job, err := repo.Create(context.Background(), missionID, "some feedback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.MissionID != missionID {
		t.Errorf("want missionID %v, got %v", missionID, job.MissionID)
	}
	if job.Status != researchjob.StatusPending {
		t.Errorf("want status pending, got %v", job.Status)
	}
	if job.Feedback == nil || *job.Feedback != "some feedback" {
		t.Errorf("want feedback 'some feedback', got %v", job.Feedback)
	}
}

func TestMockRepository_UpdateStatus_MissionDeleted(t *testing.T) {
	repo := &MockRepository{}
	// UpdateStatus for unknown job returns 0 rows (mission deleted scenario)
	n, err := repo.UpdateStatus(context.Background(), uuid.New(), researchjob.StatusDone, "", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("want 0 rows affected, got %d", n)
	}
}

func TestMockRepository_HasActiveJob(t *testing.T) {
	repo := &MockRepository{}
	missionID := uuid.New()

	has, _ := repo.HasActiveJob(context.Background(), missionID)
	if has {
		t.Fatal("expected no active job")
	}

	repo.Create(context.Background(), missionID, "")

	has, _ = repo.HasActiveJob(context.Background(), missionID)
	if !has {
		t.Fatal("expected active job after create")
	}
}
```

- [ ] **Step 2: Run tests — expect compile fail (repository not defined yet)**

```bash
cd backend && go test ./internal/researchjob/... -v 2>&1 | head -30
```

Expected: compile error about missing package/types.

- [ ] **Step 3: Write repository**

```go
// backend/internal/researchjob/repository.go
package researchjob

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for research jobs.
type Repository interface {
	Create(ctx context.Context, missionID uuid.UUID, feedback string) (*ResearchJob, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, optionsCount int) (int64, error)
	GetByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]ResearchJob, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchJob, error)
	HasActiveJob(ctx context.Context, missionID uuid.UUID) (bool, error)
	FailStaleJobs(ctx context.Context) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed research job repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const jobCols = `id, mission_id, status, feedback, error, options_count, started_at, completed_at, created_at`

func (r *pgRepository) Create(ctx context.Context, missionID uuid.UUID, feedback string) (*ResearchJob, error) {
	var fb *string
	if feedback != "" {
		fb = &feedback
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO research_jobs (mission_id, feedback)
		VALUES ($1, $2)
		RETURNING `+jobCols,
		missionID, fb)
	return scanJob(row)
}

// UpdateStatus updates a job's status and sets timing/result fields.
// Returns (rowsAffected, error). rowsAffected == 0 means the job row no longer
// exists (mission was deleted via CASCADE).
func (r *pgRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, optionsCount int) (int64, error) {
	var tag interface{ RowsAffected() int64 }
	var err error

	switch status {
	case StatusRunning:
		now := time.Now()
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, started_at = $3
			WHERE id = $1`, id, status, now)
	case StatusDone:
		now := time.Now()
		var cnt *int
		if optionsCount > 0 {
			cnt = &optionsCount
		}
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, options_count = $3, completed_at = $4
			WHERE id = $1`, id, status, cnt, now)
	case StatusFailed:
		now := time.Now()
		var errPtr *string
		if errMsg != "" {
			errPtr = &errMsg
		}
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, error = $3, completed_at = $4
			WHERE id = $1`, id, status, errPtr, now)
	default:
		return 0, fmt.Errorf("unknown status: %s", status)
	}

	if err != nil {
		return 0, fmt.Errorf("update research job status: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *pgRepository) GetByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]ResearchJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+jobCols+`
		FROM research_jobs
		WHERE mission_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, missionID, limit)
	if err != nil {
		return nil, fmt.Errorf("get research jobs by mission: %w", err)
	}
	defer rows.Close()

	var jobs []ResearchJob
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *j)
	}
	return jobs, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*ResearchJob, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+jobCols+`
		FROM research_jobs
		WHERE id = $1`, id)
	j, err := scanJob(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

func (r *pgRepository) HasActiveJob(ctx context.Context, missionID uuid.UUID) (bool, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM research_jobs
		WHERE mission_id = $1 AND status IN ('pending', 'running')
		LIMIT 1`, missionID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check active research job: %w", err)
	}
	return true, nil
}

func (r *pgRepository) FailStaleJobs(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE research_jobs
		SET status = 'failed', error = 'server restarted', completed_at = now()
		WHERE status IN ('pending', 'running')`)
	if err != nil {
		return fmt.Errorf("fail stale research jobs: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(s scanner) (*ResearchJob, error) {
	var j ResearchJob
	err := s.Scan(
		&j.ID, &j.MissionID, &j.Status, &j.Feedback,
		&j.Error, &j.OptionsCount, &j.StartedAt, &j.CompletedAt, &j.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan research job: %w", err)
	}
	return &j, nil
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd backend && go test ./internal/researchjob/... -v -run TestMock
```

Expected: PASS for all 3 TestMock* tests.

- [ ] **Step 5: Verify build**

```bash
cd backend && go build ./internal/researchjob/...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/researchjob/
git commit -m "feat(researchjob): add Repository interface and pgRepository"
```

---

### Task 4: researchjob runner (goroutine)

**Files:**
- Create: `backend/internal/researchjob/runner.go`
- Create: `backend/internal/researchjob/runner_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// backend/internal/researchjob/runner_test.go
package researchjob_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/researchjob"
)

// MockAgent implements the AgentRunner interface for testing.
type MockAgent struct {
	result  []option.Option
	err     error
	panics  bool
}

func (m *MockAgent) Run(ctx context.Context, ms mission.Mission, fb *researchjob.FeedbackInput) ([]option.Option, error) {
	if m.panics {
		panic("test panic")
	}
	return m.result, m.err
}

func TestRunResearchJob_Success(t *testing.T) {
	repo := &MockRepository{}
	missionID := uuid.New()
	job, _ := repo.Create(context.Background(), missionID, "")

	agent := &MockAgent{
		result: []option.Option{{ID: uuid.New(), Name: "MacBook Pro"}, {ID: uuid.New(), Name: "Dell XPS"}},
	}

	logger := slog.Default()
	ms := mission.Mission{ID: missionID}

	researchjob.RunResearchJob(job.ID, ms, "", agent, repo, logger)

	// Wait briefly for goroutine (RunResearchJob is synchronous in tests)
	jobs, _ := repo.GetByMission(context.Background(), missionID, 10)
	if len(jobs) == 0 {
		t.Fatal("no jobs found")
	}
	if jobs[0].Status != researchjob.StatusDone {
		t.Errorf("want status done, got %s", jobs[0].Status)
	}
	if jobs[0].OptionsCount == nil || *jobs[0].OptionsCount != 2 {
		t.Errorf("want options_count 2, got %v", jobs[0].OptionsCount)
	}
}

func TestRunResearchJob_AgentError(t *testing.T) {
	repo := &MockRepository{}
	missionID := uuid.New()
	job, _ := repo.Create(context.Background(), missionID, "")

	agent := &MockAgent{err: errors.New("LLM timeout")}
	logger := slog.Default()
	ms := mission.Mission{ID: missionID}

	researchjob.RunResearchJob(job.ID, ms, "", agent, repo, logger)

	jobs, _ := repo.GetByMission(context.Background(), missionID, 10)
	if jobs[0].Status != researchjob.StatusFailed {
		t.Errorf("want status failed, got %s", jobs[0].Status)
	}
	if jobs[0].Error == nil || *jobs[0].Error != "LLM timeout" {
		t.Errorf("want error 'LLM timeout', got %v", jobs[0].Error)
	}
}

func TestRunResearchJob_PanicRecovery(t *testing.T) {
	repo := &MockRepository{}
	missionID := uuid.New()
	job, _ := repo.Create(context.Background(), missionID, "")

	agent := &MockAgent{panics: true}
	logger := slog.Default()
	ms := mission.Mission{ID: missionID}

	// Must not panic the caller
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RunResearchJob leaked a panic: %v", r)
		}
	}()

	researchjob.RunResearchJob(job.ID, ms, "", agent, repo, logger)

	// Give goroutine a moment if async
	time.Sleep(50 * time.Millisecond)

	jobs, _ := repo.GetByMission(context.Background(), missionID, 10)
	if jobs[0].Status != researchjob.StatusFailed {
		t.Errorf("want status failed after panic, got %s", jobs[0].Status)
	}
}
```

- [ ] **Step 2: Run tests — expect compile fail**

```bash
cd backend && go test ./internal/researchjob/... -v -run TestRunResearchJob 2>&1 | head -30
```

Expected: compile errors (RunResearchJob, AgentRunner, FeedbackInput not defined).

- [ ] **Step 3: Write runner**

```go
// backend/internal/researchjob/runner.go
package researchjob

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// FeedbackInput matches research.FeedbackInput for decoupling.
type FeedbackInput struct {
	Feedback string `json:"feedback"`
}

// AgentRunner is the interface the research agent must satisfy.
type AgentRunner interface {
	Run(ctx context.Context, m mission.Mission, fb *FeedbackInput) ([]option.Option, error)
}

// RunResearchJob executes research for a job, updating its status in the DB.
// It is safe to call as a goroutine — panics are recovered and written as failures.
// The function is intentionally synchronous so it can be tested directly.
func RunResearchJob(jobID uuid.UUID, ms mission.Mission, feedback string, agent AgentRunner, repo Repository, logger *slog.Logger) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in research job", "job_id", jobID, "panic", r)
			if _, err := repo.UpdateStatus(ctx, jobID, StatusFailed, "internal panic", 0); err != nil {
				logger.Error("failed to mark panicked job as failed", "job_id", jobID, "err", err)
			}
		}
	}()

	// Transition to running — agent.Run() owns the LLM timeout (5 min).
	if _, err := repo.UpdateStatus(ctx, jobID, StatusRunning, "", 0); err != nil {
		logger.Error("failed to mark job as running", "job_id", jobID, "err", err)
		return
	}

	var fb *FeedbackInput
	if feedback != "" {
		fb = &FeedbackInput{Feedback: feedback}
	}

	// agent.Run uses context.Background() with its own 5-minute LLM timeout.
	results, err := agent.Run(ctx, ms, fb)
	if err != nil {
		logger.Error("research job failed", "job_id", jobID, "err", err)
		n, updateErr := repo.UpdateStatus(ctx, jobID, StatusFailed, err.Error(), 0)
		if updateErr != nil {
			logger.Error("failed to mark job as failed", "job_id", jobID, "err", updateErr)
		} else if n == 0 {
			logger.Warn("research job row not found (mission deleted?)", "job_id", jobID)
		}
		return
	}

	n, updateErr := repo.UpdateStatus(ctx, jobID, StatusDone, "", len(results))
	if updateErr != nil {
		logger.Error("failed to mark job as done", "job_id", jobID, "err", updateErr)
	} else if n == 0 {
		logger.Warn("research job row not found after completion (mission deleted?)", "job_id", jobID)
	}
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd backend && go test ./internal/researchjob/... -v -run TestRunResearchJob
```

Expected: PASS for all 3 tests.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/researchjob/runner.go backend/internal/researchjob/runner_test.go
git commit -m "feat(researchjob): add RunResearchJob goroutine with panic recovery"
```

---

### Task 5: researchjob handler

**Files:**
- Create: `backend/internal/researchjob/handler.go`
- Create: `backend/internal/researchjob/handler_test.go`

- [ ] **Step 1: Write failing handler tests**

```go
// backend/internal/researchjob/handler_test.go
package researchjob_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/researchjob"
)

// stubMissionService returns a fixed mission.
type stubMissionService struct {
	m *mission.Mission
}

func (s *stubMissionService) GetByID(_ interface{}, id interface{}) (*mission.Mission, error) {
	return s.m, nil
}

// newTestRouter wires a Handler with mock deps onto a chi router.
func newTestRouter(repo researchjob.Repository, ms *mission.Mission) chi.Router {
	svc := &stubMissionService{m: ms}
	// MockAgent never panics, returns 0 options
	agent := &MockAgent{result: nil}
	h := researchjob.NewHandler(repo, svc, agent)
	r := chi.NewRouter()
	r.Route("/api/missions/{missionID}", func(r chi.Router) {
		r.Mount("/research", h.Routes())
	})
	return r
}

func TestHandler_Trigger_202(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID, Name: "Test Mission"}
	repo := &MockRepository{}
	router := newTestRouter(repo, ms)

	body := strings.NewReader(`{"feedback":"cheaper please"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID.String()+"/research", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["job_id"] == nil {
		t.Error("expected job_id in response")
	}
	if resp["status"] != "pending" {
		t.Errorf("want status pending, got %v", resp["status"])
	}
}

func TestHandler_Trigger_409_Concurrent(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID}
	repo := &MockRepository{}
	// Pre-insert a running job
	repo.Create(nil, missionID, "")

	router := newTestRouter(repo, ms)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID.String()+"/research", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
}

func TestHandler_List_200(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID}
	repo := &MockRepository{}
	repo.Create(nil, missionID, "feedback")

	router := newTestRouter(repo, ms)
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID.String()+"/research/jobs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["jobs"] == nil {
		t.Error("expected jobs array in response")
	}
}
```

- [ ] **Step 2: Run tests — expect compile fail**

```bash
cd backend && go test ./internal/researchjob/... -v -run TestHandler 2>&1 | head -20
```

- [ ] **Step 3: Write handler**

Note: The handler's `AgentRunner` uses `*FeedbackInput` from this package, but the actual `research.Agent.Run` takes `*research.FeedbackInput`. The adapter is wired in `routes.go` via a thin wrapper (see Task 6).

```go
// backend/internal/researchjob/handler.go
package researchjob

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

// MissionGetter is the subset of mission.Service the handler needs.
type MissionGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*mission.Mission, error)
}

// Handler exposes research job endpoints.
type Handler struct {
	repo       Repository
	missionSvc MissionGetter
	agent      AgentRunner
	logger     *slog.Logger
}

// NewHandler constructs a Handler.
func NewHandler(repo Repository, missionSvc MissionGetter, agent AgentRunner) *Handler {
	return &Handler{
		repo:       repo,
		missionSvc: missionSvc,
		agent:      agent,
		logger:     slog.Default(),
	}
}

// Routes registers the three research job endpoints under the parent subrouter.
// Mount at "/research" under /api/missions/{missionID}.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.trigger)
	r.Get("/jobs", h.list)
	r.Get("/jobs/{jobID}", h.get)
	return r
}

// trigger: POST /api/missions/{missionID}/research
func (h *Handler) trigger(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	ms, err := h.missionSvc.GetByID(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ms == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	active, err := h.repo.HasActiveJob(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if active {
		httputil.WriteError(w, http.StatusConflict, "research already running")
		return
	}

	var feedback string
	var input FeedbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err == nil && input.Feedback != "" {
		feedback = input.Feedback
	}

	job, err := h.repo.Create(r.Context(), missionID, feedback)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	go RunResearchJob(job.ID, *ms, feedback, h.agent, h.repo, h.logger)

	httputil.WriteJSON(w, http.StatusAccepted, map[string]any{
		"job_id": job.ID,
		"status": StatusPending,
	})
}

// list: GET /api/missions/{missionID}/research/jobs
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	jobs, err := h.repo.GetByMission(r.Context(), missionID, 10)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if jobs == nil {
		jobs = []ResearchJob{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

// get: GET /api/missions/{missionID}/research/jobs/{jobID}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.repo.GetByID(r.Context(), jobID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if job == nil {
		httputil.WriteError(w, http.StatusNotFound, "job not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, job)
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd backend && go test ./internal/researchjob/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/researchjob/handler.go backend/internal/researchjob/handler_test.go
git commit -m "feat(researchjob): add HTTP handler with trigger/list/get endpoints"
```

---

### Task 6: Wire routes + bump agent timeout

**Files:**
- Modify: `backend/internal/research/agent.go` (lines ~112–115: change timeout)
- Modify: `backend/cmd/server/routes.go` (add researchjob import + wiring)
- Modify: `backend/cmd/server/main.go` (add researchjob repo/handler instantiation)

The research agent's `AgentRunner` interface takes `*research.FeedbackInput`, but `researchjob.AgentRunner` takes `*researchjob.FeedbackInput`. Bridge this with a thin adapter in routes.go.

- [ ] **Step 1: Bump LLM timeout in agent.go**

In `backend/internal/research/agent.go`, find:
```go
llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
```
Change to:
```go
llmCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
```

- [ ] **Step 2: Verify agent still builds**

```bash
cd backend && go build ./internal/research/...
```

- [ ] **Step 3: Add researchjob adapter in routes.go**

In `routes.go`, add the import `"github.com/jibei/scouter/internal/researchjob"` and add a `researchAgentAdapter` type near the other local types in the file (or in a small file `cmd/server/adapter.go`):

```go
// researchAgentAdapter adapts research.Agent to researchjob.AgentRunner.
// research.FeedbackInput and researchjob.FeedbackInput are identical structs.
type researchAgentAdapter struct {
	agent *research.Agent
}

func (a *researchAgentAdapter) Run(ctx context.Context, m mission.Mission, fb *researchjob.FeedbackInput) ([]option.Option, error) {
	var rFb *research.FeedbackInput
	if fb != nil {
		rFb = &research.FeedbackInput{Feedback: fb.Feedback}
	}
	return a.agent.Run(ctx, m, rFb)
}
```

- [ ] **Step 4: Wire researchjob handler in routes**

In `routes.go`, in the `routeDeps` struct, add:
```go
researchJobRepo    researchjob.Repository
researchJobHandler *researchjob.Handler
```

In `registerRoutes`, find where `researchH.Routes()` is mounted (search for `research` in the mission subrouter block). Replace:
```go
r.Mount("/research", researchH.Routes())
```
With:
```go
r.Mount("/research", deps.researchJobHandler.Routes())
```

(The old `researchH` can be removed from `routeDeps` once all its callers are replaced — just the one route mount.)

- [ ] **Step 5: Add stale job cleanup in routes.go startup**

In `registerRoutes` (or the init section at the top of the function), after the DB pool is available:
```go
if err := deps.researchJobRepo.FailStaleJobs(context.Background()); err != nil {
    slog.Warn("failed to clean stale research jobs", "err", err)
}
```

- [ ] **Step 6: Instantiate researchjob deps in main.go**

In `main.go`, after the research agent is created, add:
```go
researchJobRepo := researchjob.NewRepository(pool)
researchJobHandler := researchjob.NewHandler(
    researchJobRepo,
    missionSvc,
    &researchAgentAdapter{agent: researchAgent},
)
```

Pass them into `routeDeps`:
```go
deps := routeDeps{
    // ... existing fields ...
    researchJobRepo:    researchJobRepo,
    researchJobHandler: researchJobHandler,
}
```

- [ ] **Step 7: Build backend**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 8: Run all backend tests**

```bash
cd backend && go test ./... 2>&1 | tail -30
```

Expected: all pass.

- [ ] **Step 9: Commit**

```bash
git add backend/internal/research/agent.go backend/cmd/server/routes.go backend/cmd/server/main.go
git commit -m "feat(researchjob): wire async research handler into routes, bump LLM timeout to 5min"
```

---

## Chunk 2: Frontend

### Task 7: API layer

**Files:**
- Create: `frontend/src/api/researchJobs.ts`
- Modify: `frontend/src/api/missions.ts` (update `triggerResearch` response schema)

- [ ] **Step 1: Write researchJobs.ts with Zod schemas**

```typescript
// frontend/src/api/researchJobs.ts
import { z } from 'zod'
import { apiFetch } from './client'

// ── Schemas ──────────────────────────────────────────────────────────────────

export const ResearchJobSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  status: z.enum(['pending', 'running', 'done', 'failed']),
  feedback: z.string().nullable(),
  error: z.string().nullable(),
  optionsCount: z.number().int().nullable(),
  startedAt: z.string().nullable(),
  completedAt: z.string().nullable(),
  createdAt: z.string(),
})

export type ResearchJob = z.infer<typeof ResearchJobSchema>

export const TriggerResearchResponseSchema = z.object({
  job_id: z.string().uuid(),
  status: z.literal('pending'),
})

export type TriggerResearchResponse = z.infer<typeof TriggerResearchResponseSchema>

export const ResearchJobsListSchema = z.object({
  jobs: z.array(ResearchJobSchema),
})

// ── API functions ─────────────────────────────────────────────────────────────

export async function triggerResearchJob(
  missionId: string,
  feedback?: string
): Promise<TriggerResearchResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/research`, {
    method: 'POST',
    body: feedback ? JSON.stringify({ feedback }) : undefined,
  })
  return TriggerResearchResponseSchema.parse(data)
}

export async function listResearchJobs(missionId: string): Promise<ResearchJob[]> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/research/jobs`)
  const { jobs } = ResearchJobsListSchema.parse(data)
  return jobs
}

export async function getResearchJob(missionId: string, jobId: string): Promise<ResearchJob> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/research/jobs/${jobId}`)
  return ResearchJobSchema.parse(data)
}
```

- [ ] **Step 2: Update missions.ts — remove triggerResearch (it moved to researchJobs.ts)**

In `frontend/src/api/missions.ts`, remove the `triggerResearch` function and `AgentResultSchema` (they are superseded by `researchJobs.ts`). If `triggerPricing` still uses `AgentResultSchema`, keep the schema for that alone.

- [ ] **Step 3: Update api/index.ts to export researchJobs**

Add to `frontend/src/api/index.ts` (or wherever the barrel export is):
```typescript
export * from './researchJobs'
```

- [ ] **Step 4: Build frontend to catch import errors**

```bash
cd frontend && npm run typecheck 2>&1 | head -40
```

Fix any import errors from removing `triggerResearch` from `missions.ts`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/researchJobs.ts frontend/src/api/missions.ts frontend/src/api/index.ts
git commit -m "feat(frontend): add researchJobs API layer, remove blocking triggerResearch"
```

---

### Task 8: hooks — useResearchJobs + useTriggerResearch

**Files:**
- Create: `frontend/src/hooks/useResearchJobs.ts`
- Modify: `frontend/src/hooks/useResearch.ts`
- Create: `frontend/src/hooks/useResearchJobs.test.ts`

- [ ] **Step 1: Write failing test for useResearchJobs**

```typescript
// frontend/src/hooks/useResearchJobs.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import * as api from '../api/researchJobs'

vi.mock('../api/researchJobs')
vi.mock('../components/scouter', () => ({
  useToast: () => ({ toast: vi.fn() }),
}))

const missionId = 'mission-123'

function makeWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children)
}

describe('useResearchJobs', () => {
  beforeEach(() => vi.clearAllMocks())

  it('returns jobs from API', async () => {
    vi.mocked(api.listResearchJobs).mockResolvedValue([
      {
        id: 'job-1', missionId, status: 'done',
        feedback: null, error: null, optionsCount: 3,
        startedAt: null, completedAt: null, createdAt: '2026-03-16T10:00:00Z',
      },
    ])

    const { result } = renderHook(() => {
      const { useResearchJobs } = require('./useResearchJobs')
      return useResearchJobs(missionId)
    }, { wrapper: makeWrapper() })

    await waitFor(() => expect(result.current.jobs).toHaveLength(1))
    expect(result.current.jobs[0].status).toBe('done')
  })

  it('does not poll when all jobs are terminal', async () => {
    vi.mocked(api.listResearchJobs).mockResolvedValue([
      {
        id: 'job-1', missionId, status: 'done',
        feedback: null, error: null, optionsCount: 2,
        startedAt: null, completedAt: null, createdAt: '2026-03-16T10:00:00Z',
      },
    ])

    const { result } = renderHook(() => {
      const { useResearchJobs } = require('./useResearchJobs')
      return useResearchJobs(missionId)
    }, { wrapper: makeWrapper() })

    await waitFor(() => expect(result.current.isPolling).toBe(false))
  })
})
```

- [ ] **Step 2: Run test — expect fail**

```bash
cd frontend && npx vitest run src/hooks/useResearchJobs.test.ts 2>&1 | tail -20
```

- [ ] **Step 3: Write useResearchJobs hook**

```typescript
// frontend/src/hooks/useResearchJobs.ts
import { useRef, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { listResearchJobs, type ResearchJob } from '../api/researchJobs'
import { useToast } from '../components/scouter'

const TERMINAL_STATUSES = new Set(['done', 'failed'])

export function useResearchJobs(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  // Track per-job previous status to detect transitions and fire toasts once.
  const prevStatusRef = useRef<Record<string, string>>({})

  const query = useQuery({
    queryKey: ['research-jobs', missionId],
    queryFn: () => listResearchJobs(missionId),
    enabled: !!missionId,
    // Poll every 3s while any job is active; stop when all are terminal.
    refetchInterval: (q) => {
      const jobs = q.state.data ?? []
      const hasActive = jobs.some((j) => !TERMINAL_STATUSES.has(j.status))
      return hasActive ? 3000 : false
    },
  })

  const jobs: ResearchJob[] = query.data ?? []
  const isPolling = jobs.some((j) => !TERMINAL_STATUSES.has(j.status))

  // Detect status transitions and fire toasts + invalidations.
  useEffect(() => {
    const prev = prevStatusRef.current
    for (const job of jobs) {
      const oldStatus = prev[job.id]
      if (oldStatus && oldStatus !== job.status) {
        if (job.status === 'done') {
          const count = job.optionsCount ?? 0
          toast(`Research complete — ${count} options found`, 'success')
          qc.invalidateQueries({ queryKey: ['options', missionId] })
          qc.invalidateQueries({ queryKey: ['agent-runs', missionId] })
        } else if (job.status === 'failed') {
          toast(`Research failed: ${job.error ?? 'Unknown error'}`, 'error')
        }
      }
      prev[job.id] = job.status
    }
    prevStatusRef.current = prev
  }, [jobs, missionId, qc, toast])

  return { jobs, isPolling, isLoading: query.isLoading }
}
```

- [ ] **Step 4: Update useResearch.ts — rename to useTriggerResearch, fire-and-forget**

Replace the contents of `frontend/src/hooks/useResearch.ts`:

```typescript
// frontend/src/hooks/useResearch.ts
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { triggerResearchJob } from '../api/researchJobs'
import { useToast } from '../components/scouter'

/**
 * useTriggerResearch fires a research job and returns immediately.
 * Use useResearchJobs to track the job's progress and update the options list.
 */
export function useTriggerResearch(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()

  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: (feedback?: string) => triggerResearchJob(missionId, feedback),
    onSuccess: () => {
      // Invalidate research-jobs so the polling hook picks up the new job.
      qc.invalidateQueries({ queryKey: ['research-jobs', missionId] })
    },
    onError: (err: unknown) => {
      const msg = err instanceof Error ? err.message : 'Unknown error'
      // 409 = already running, surface it distinctly
      if (msg.includes('409') || msg.toLowerCase().includes('already running')) {
        toast('Research is already running', 'info')
      } else {
        toast(`Failed to start research: ${msg}`, 'error')
      }
    },
  })

  return { triggerResearch: mutateAsync, isPending }
}

// Re-export under old name for any consumers not yet updated.
export { useTriggerResearch as useResearch }
```

- [ ] **Step 5: Export new hook from hooks index**

In `frontend/src/hooks/index.ts`, add:
```typescript
export { useResearchJobs } from './useResearchJobs'
export { useTriggerResearch } from './useResearch'
```

- [ ] **Step 6: Run tests**

```bash
cd frontend && npx vitest run src/hooks/useResearchJobs.test.ts 2>&1 | tail -20
```

Expected: PASS.

- [ ] **Step 7: Typecheck**

```bash
cd frontend && npm run typecheck 2>&1 | head -30
```

- [ ] **Step 8: Commit**

```bash
git add frontend/src/hooks/useResearchJobs.ts frontend/src/hooks/useResearch.ts frontend/src/hooks/index.ts frontend/src/hooks/useResearchJobs.test.ts
git commit -m "feat(frontend): add useResearchJobs polling hook, convert useResearch to fire-and-forget"
```

---

### Task 9: OptionsExplorer UI + i18n

**Files:**
- Modify: `frontend/src/pages/OptionsExplorer.tsx`
- Modify: `frontend/src/pages/OptionsExplorer.module.css`
- Modify: `frontend/src/i18n/en.json`
- Modify: `frontend/src/i18n/fr.json`

- [ ] **Step 1: Add i18n keys to en.json**

Find the `"research"` key in `frontend/src/i18n/en.json`. Add or merge the following keys (do not overwrite existing keys):
```json
"researchJob": {
  "running": "Research running…",
  "lastRunSuccess": "Last run: {{date}} · {{count}} options found",
  "lastRunFailed": "Last run: {{date}} · Failed",
  "showHistory": "show history",
  "hideHistory": "hide history",
  "complete": "Research complete — {{count}} options found",
  "alreadyRunning": "Research is already running",
  "historyStatus": {
    "done": "✓",
    "failed": "✗",
    "running": "🔄",
    "pending": "⏳"
  }
}
```

- [ ] **Step 2: Add i18n keys to fr.json**

Same structure, French values:
```json
"researchJob": {
  "running": "Recherche en cours…",
  "lastRunSuccess": "Dernière recherche : {{date}} · {{count}} options trouvées",
  "lastRunFailed": "Dernière recherche : {{date}} · Échec",
  "showHistory": "voir l'historique",
  "hideHistory": "masquer l'historique",
  "complete": "Recherche terminée — {{count}} options trouvées",
  "alreadyRunning": "Une recherche est déjà en cours",
  "historyStatus": {
    "done": "✓",
    "failed": "✗",
    "running": "🔄",
    "pending": "⏳"
  }
}
```

- [ ] **Step 3: Add CSS for status bar and history list**

Append to `frontend/src/pages/OptionsExplorer.module.css`:

```css
/* Research job status bar */
.researchStatus {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  min-height: 24px;
}

.researchBadge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--warning-dim, rgba(245, 158, 11, 0.15));
  color: var(--warning, #f59e0b);
  border-radius: 20px;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 500;
}

.researchLastRun {
  font-size: 12px;
  color: var(--text-dim);
  margin-bottom: 12px;
}

.researchHistoryToggle {
  color: var(--accent, #4a9eff);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 12px;
  padding: 0;
  margin-left: 6px;
}

.researchHistoryToggle:hover {
  text-decoration: underline;
}

.researchHistory {
  background: var(--surface-2, #1e2330);
  border: 1px solid var(--border, #2a3040);
  border-radius: 8px;
  padding: 8px 12px;
  margin-bottom: 12px;
  font-size: 12px;
}

.researchHistoryRow {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  color: var(--text-dim);
  border-bottom: 1px solid var(--border, #2a3040);
}

.researchHistoryRow:last-child {
  border-bottom: none;
}

.researchHistoryStatus {
  margin-right: 6px;
}
```

- [ ] **Step 4: Update OptionsExplorer.tsx**

In `OptionsExplorer.tsx`:

**a) Update imports** — replace `useResearch` import with `useTriggerResearch` and add `useResearchJobs`:

```typescript
import {
  // ... existing imports ...
  useTriggerResearch,    // replaces useResearch
  useResearchJobs,       // new
  // ... rest of imports
} from '../hooks'
```

**b) Replace hook usage** — find:
```typescript
const { triggerResearch, isPending: researchPending } = useResearch(mission?.id ?? '')
```
Replace with:
```typescript
const { triggerResearch, isPending: triggerPending } = useTriggerResearch(mission?.id ?? '')
const { jobs, isPolling } = useResearchJobs(mission?.id ?? '')
const researchPending = triggerPending || isPolling
```

**c) Add history toggle state** — add after the other `useState` declarations:
```typescript
const [showJobHistory, setShowJobHistory] = useState(false)
```

**d) Add ResearchStatusBar component inline** — insert below the research button in the header JSX. Find the research button section and add after it:

```tsx
{/* Research job status bar */}
{isPolling && (
  <span className={styles.researchBadge}>
    🔄 {t('researchJob.running')}
  </span>
)}

{jobs.length > 0 && (
  <div className={styles.researchLastRun}>
    {(() => {
      const last = jobs[0]
      if (last.status === 'done') {
        return t('researchJob.lastRunSuccess', {
          date: new Date(last.createdAt).toLocaleDateString(),
          count: last.optionsCount ?? 0,
        })
      }
      if (last.status === 'failed') {
        return t('researchJob.lastRunFailed', {
          date: new Date(last.createdAt).toLocaleDateString(),
        })
      }
      return null
    })()}
    {jobs.length > 0 && (
      <button
        className={styles.researchHistoryToggle}
        onClick={() => setShowJobHistory((v) => !v)}
      >
        {showJobHistory ? t('researchJob.hideHistory') : t('researchJob.showHistory')} ↓
      </button>
    )}
  </div>
)}

{showJobHistory && jobs.length > 0 && (
  <div className={styles.researchHistory}>
    {jobs.map((job) => (
      <div key={job.id} className={styles.researchHistoryRow}>
        <span>
          <span className={styles.researchHistoryStatus}>
            {t(`researchJob.historyStatus.${job.status}`)}
          </span>
          {new Date(job.createdAt).toLocaleString()}
          {job.status === 'done' && job.optionsCount != null && ` · ${job.optionsCount} options`}
          {job.status === 'failed' && job.error && ` · ${job.error}`}
        </span>
      </div>
    ))}
  </div>
)}
```

- [ ] **Step 5: Build frontend**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

Expected: no errors.

- [ ] **Step 6: Typecheck**

```bash
cd frontend && npm run typecheck 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 7: Run frontend tests**

```bash
cd frontend && npx vitest run 2>&1 | tail -20
```

Expected: all pass (fix any snapshot/import failures caused by the hook rename).

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/OptionsExplorer.tsx frontend/src/pages/OptionsExplorer.module.css frontend/src/i18n/en.json frontend/src/i18n/fr.json
git commit -m "feat(frontend): async research job status bar, history list, polling in OptionsExplorer"
```

---

## Final verification

- [ ] **Backend tests green**

```bash
cd backend && go test ./... 2>&1 | grep -E "FAIL|ok"
```

- [ ] **Frontend build + typecheck + tests green**

```bash
cd frontend && npm run typecheck && npm run build && npx vitest run 2>&1 | tail -10
```

- [ ] **Smoke test against running stack**

```bash
make dev  # starts docker-compose stack
# In browser: open a mission's Options screen
# Click "Run Research" — should return instantly (button re-enables, badge appears)
# Wait for polling to detect job completion (~5s on fast LLM)
# Options list should refresh automatically
```

- [ ] **Final commit tag**

```bash
git tag v0.2.0-async-research
```
