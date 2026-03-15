// Package reviews implements the ReviewAgent, which uses LLM tool-use to
// generate a simulated review summary for a product option. It owns prompt
// construction, tool schema definition, and response parsing.
// The llm.Provider is transport-only.
package reviews

// ReviewSummary is the structured output returned by the ReviewAgent and cached
// by the Handler.
type ReviewSummary struct {
	OptionID     string   `json:"optionId"`
	Sentiment    float64  `json:"sentiment"`    // -1..1, positive = good
	PositivePct  int      `json:"positivePct"`
	NeutralPct   int      `json:"neutralPct"`
	NegativePct  int      `json:"negativePct"`
	Pros         []string `json:"pros"`
	Cons         []string `json:"cons"`
	TotalReviews int      `json:"totalReviews"`
	Source       string   `json:"source"`
	RefreshedAt  string   `json:"refreshedAt"` // RFC3339
}
