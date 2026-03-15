// Package timing implements the AI Purchase Timing Advisor, which uses LLM
// tool-use to recommend the best time to buy for a given mission.
// It owns prompt construction, tool schema definition, and response parsing.
// The llm.Provider is transport-only.
package timing

// Recommendation values returned in TimingAdvice.Recommendation.
const (
	RecommendBuyNow = "BUY_NOW"
	RecommendWait   = "WAIT"
	RecommendUnsure = "UNSURE"
)

// TimingAdvice is the structured output returned by the TimingAgent and cached
// by the Handler.
type TimingAdvice struct {
	MissionID      string  `json:"missionId"`
	Recommendation string  `json:"recommendation"`          // BUY_NOW | WAIT | UNSURE
	Rationale      string  `json:"rationale"`
	WaitUntil      string  `json:"waitUntil,omitempty"`     // ISO date if WAIT
	NextSaleEvent  string  `json:"nextSaleEvent,omitempty"` // e.g. "Soldes d'été 2026"
	Confidence     float64 `json:"confidence"`              // 0..1
	GeneratedAt    string  `json:"generatedAt"`             // RFC3339
}
