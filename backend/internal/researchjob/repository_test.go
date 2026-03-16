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
			if errMsg != "" {
				e := errMsg
				m.jobs[i].Error = &e
			}
			if optionsCount > 0 {
				c := optionsCount
				m.jobs[i].OptionsCount = &c
			}
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
