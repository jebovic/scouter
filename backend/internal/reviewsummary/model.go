// Package reviewsummary generates deterministic structured review summaries
// for shopping items, seeded by the item ID hash. No LLM call is needed —
// the generator produces realistic, varied data without external dependencies.
package reviewsummary

import "time"

// ReviewSummary is the structured output returned by the Handler.
type ReviewSummary struct {
	ItemID         string    `json:"itemId"`
	ItemName       string    `json:"itemName"`
	OverallScore   float64   `json:"overallScore"`   // 0–10
	TotalReviews   int       `json:"totalReviews"`
	Pros           []string  `json:"pros"`
	Cons           []string  `json:"cons"`
	Verdict        string    `json:"verdict"`
	RecommendedFor []string  `json:"recommendedFor"`
	AvoidIf        []string  `json:"avoidIf"`
	LastUpdated    time.Time `json:"lastUpdated"`
}
