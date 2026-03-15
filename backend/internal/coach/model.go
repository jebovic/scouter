// Package coach implements the CoachAgent, which uses LLM analysis to provide
// actionable tips and suggestions for a mission based on its current state.
package coach

import "time"

// CoachTip represents a single actionable suggestion for a mission.
type CoachTip struct {
	ID       string `json:"id"`       // deterministic: sha256(missionID+content)[0:8]
	Category string `json:"category"` // "research" | "budget" | "timing" | "decision"
	Priority string `json:"priority"` // "high" | "medium" | "low"
	Title    string `json:"title"`
	Body     string `json:"body"`
	Action   string `json:"action,omitempty"` // optional CTA e.g. "Launch research"
}

// CoachResponse contains tips and metadata for a mission.
type CoachResponse struct {
	MissionID   string    `json:"missionId"`
	Tips        []CoachTip `json:"tips"`
	GeneratedAt time.Time `json:"generatedAt"`
}
