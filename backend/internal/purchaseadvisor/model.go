// Package purchaseadvisor provides AI-powered personalized purchase recommendations
// using LLM tool-use to deliver structured buy/wait/skip/negotiate decisions.
package purchaseadvisor

// Decision constants represent the advisor's recommended action.
const (
	DecisionBuyNow    = "buy_now"
	DecisionWait      = "wait"
	DecisionSkip      = "skip"
	DecisionNegotiate = "negotiate"
)

// PurchaseAdvice is the structured AI recommendation for a shopping item.
type PurchaseAdvice struct {
	ItemID                    string  `json:"itemId"`
	ItemName                  string  `json:"itemName"`
	Decision                  string  `json:"decision"`    // "buy_now"|"wait"|"skip"|"negotiate"
	Confidence                int     `json:"confidence"`  // 0-100
	Reasoning                 string  `json:"reasoning"`   // French explanation
	AlternativeSuggestion     string  `json:"alternativeSuggestion"`
	BestTimeToAct             string  `json:"bestTimeToAct"`
	EstimatedSavingsByWaiting float64 `json:"estimatedSavingsByWaiting"`
}

// AdvisorRequest contains all context needed to generate a purchase recommendation.
type AdvisorRequest struct {
	ItemID       string    `json:"itemId"`
	ItemName     string    `json:"itemName"`
	Category     string    `json:"category"`
	CurrentPrice float64   `json:"currentPrice"`
	TargetPrice  float64   `json:"targetPrice"`
	Budget       float64   `json:"budget"`
	Status       string    `json:"status"`
	PriceHistory []float64 `json:"priceHistory"`
}
