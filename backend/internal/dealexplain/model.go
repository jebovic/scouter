// Package dealexplain implements the AI Deal Explainer feature.
// It generates a 2-3 sentence French explanation of why a shopping item
// is (or is not) a good deal, plus short French tags.
package dealexplain

import "time"

// DealExplanation is the response returned by the explain endpoint.
type DealExplanation struct {
	ItemID      string    `json:"itemId"`
	Explanation string    `json:"explanation"`  // 2-3 sentences in French
	Tags        []string  `json:"tags"`          // e.g. ["prix bas", "tendance baisse"]
	GeneratedAt time.Time `json:"generatedAt"`
}

// ExplainInput carries the context needed by the LLM agent.
type ExplainInput struct {
	ItemID   string
	Name     string
	Merchant string
	Price    float64
	Score    int    // 0-100 deal score
	Trend    string // "rising" | "falling" | "stable"
}
