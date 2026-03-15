package regretanalyzer

import (
	"github.com/google/uuid"
)

type RegretSignal struct {
	Type              string  `json:"type"`              // "overpaid" | "impulse" | "timing" | "ok"
	Severity          string  `json:"severity"`          // "high" | "medium" | "low" | "none"
	Description       string  `json:"description"`
	PotentialSaving   float64 `json:"potentialSaving"`
}

type Response struct {
	MissionID   uuid.UUID      `json:"missionId"`
	TotalSpent  float64        `json:"totalSpent"`
	Signals     []RegretSignal `json:"signals"`
	RegretScore int            `json:"regretScore"`  // 0-100, higher = more regret
	Summary     string         `json:"summary"`
}
