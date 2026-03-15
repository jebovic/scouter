// Package rebalancer implements the Smart Budget Rebalancer AI agent (Phase 74).
// It analyses envelope allocations vs. actual mission spending and suggests
// monthly budget adjustments via the LLM rebalance_budget tool.
package rebalancer

// EnvelopeAdjustment holds the suggested reallocation for a single envelope.
type EnvelopeAdjustment struct {
	EnvelopeID       string  `json:"envelopeId"`
	EnvelopeName     string  `json:"envelopeName"`
	CurrentMonthly   float64 `json:"currentMonthly"`
	SuggestedMonthly float64 `json:"suggestedMonthly"`
	// DeltaPercent is positive for an increase, negative for a decrease.
	DeltaPercent float64 `json:"deltaPercent"`
	// Reason is a French-language explanation of the suggested change.
	Reason string `json:"reason"`
}

// RebalancePlan is the complete rebalancing suggestion returned by the agent.
type RebalancePlan struct {
	Adjustments    []EnvelopeAdjustment `json:"adjustments"`
	// Summary is a French-language executive summary of the plan.
	Summary        string               `json:"summary"`
	TotalCurrent   float64              `json:"totalCurrent"`
	TotalSuggested float64              `json:"totalSuggested"`
	// CachedAt is the Unix timestamp when this plan was computed.
	CachedAt       int64                `json:"cachedAt"`
}
