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
