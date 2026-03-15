package health

import "time"

// SignalScore represents a single scored dimension of mission health.
type SignalScore struct {
	Name        string  `json:"name"`
	Score       int     `json:"score"`   // 0–100
	Weight      float64 `json:"weight"`  // 0.0–1.0
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

// HealthReport is the composite mission health assessment.
type HealthReport struct {
	MissionID    string        `json:"missionId"`
	OverallScore int           `json:"overallScore"` // 0–100, weighted average
	Grade        string        `json:"grade"`        // "A" | "B" | "C" | "D" | "F"
	Signals      []SignalScore `json:"signals"`
	Summary      string        `json:"summary"` // 1–2 sentence AI-generated summary in French
	Advice       string        `json:"advice"`  // 1 sentence actionable advice in French
	GeneratedAt  time.Time     `json:"generatedAt"`
}
