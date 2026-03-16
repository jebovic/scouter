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
