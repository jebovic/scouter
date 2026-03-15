// Package negotiate implements the AI Negotiation Coach for shopping items.
// It generates actionable haggling tips for French market shoppers.
package negotiate

// NegotiationTip is a single actionable negotiation tip.
type NegotiationTip struct {
	Title       string `json:"title"`
	Description string `json:"description"` // French, 1-2 sentences
	Difficulty  string `json:"difficulty"`  // "easy" | "medium" | "hard"
}

// NegotiationAdvice is the full AI-generated negotiation coach output for a shopping item.
type NegotiationAdvice struct {
	ItemID       string           `json:"itemId"`
	ItemName     string           `json:"itemName"`
	CurrentPrice float64          `json:"currentPrice"`
	TargetPrice  float64          `json:"targetPrice"` // 13% below current (currentPrice * 0.87)
	Tips         []NegotiationTip `json:"tips"`        // 4-6 tips
	Script       string           `json:"script"`      // Ready-to-use French phrase
	CachedAt     int64            `json:"cachedAt"`
}
