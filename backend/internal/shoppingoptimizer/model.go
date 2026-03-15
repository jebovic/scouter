// Package shoppingoptimizer implements a deterministic shopping list prioritizer
// that ranks items by deal quality and urgency without LLM calls.
package shoppingoptimizer

// OptimizedItem is a shopping item annotated with priority metadata.
type OptimizedItem struct {
	ItemID           string   `json:"itemId"`
	ItemName         string   `json:"itemName"`
	Merchant         string   `json:"merchant"`
	Price            float64  `json:"price"`
	PriorityScore    float64  `json:"priorityScore"`
	Reasons          []string `json:"reasons"`
	UrgencyLevel     string   `json:"urgencyLevel"` // "now" | "soon" | "wait"
	SavingsPotential float64  `json:"savingsPotential"`
}

// OptimizeResponse is the full result returned by the optimizer endpoint.
type OptimizeResponse struct {
	Items          []OptimizedItem `json:"items"`
	TotalBudget    float64         `json:"totalBudget"`
	TotalOptimized float64         `json:"totalOptimized"`
	Summary        string          `json:"summary"`
}
