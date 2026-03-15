// Package optimizer implements the AI Shopping List Optimizer,
// which uses an LLM to suggest an ordered buying strategy for a mission's shopping items.
package optimizer

import "time"

// ShoppingAction represents a single step in the optimized buying plan.
type ShoppingAction struct {
	Priority   int      `json:"priority"`   // 1 = first to execute
	ItemIDs    []string `json:"itemIds"`    // IDs of items in this step
	ItemNames  []string `json:"itemNames"`  // human-readable names
	Merchant   string   `json:"merchant"`
	Action     string   `json:"action"`     // "Acheter maintenant" | "Attendre la promo" | "Comparer en magasin"
	Reason     string   `json:"reason"`     // French explanation
	EstSavings float64  `json:"estSavings"` // estimated savings (0 when N/A)
	Currency   string   `json:"currency"`
}

// OptimizedPlan is the full ordered action plan returned by the optimizer.
type OptimizedPlan struct {
	MissionID   string           `json:"missionId"`
	GeneratedAt time.Time        `json:"generatedAt"`
	TotalItems  int              `json:"totalItems"`
	Actions     []ShoppingAction `json:"actions"`
	Summary     string           `json:"summary"` // 1-2 sentence French summary
}
