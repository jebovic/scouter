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
	result []option.Option
	err    error
	panics bool
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
