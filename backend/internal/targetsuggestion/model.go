package targetsuggestion

import "time"

// TargetSuggestion holds auto-suggested target prices for a shopping item.
type TargetSuggestion struct {
	ItemID           string    `json:"itemId"`
	CurrentPrice     float64   `json:"currentPrice"`
	Conservative     float64   `json:"conservative"`     // 10% below current
	Moderate         float64   `json:"moderate"`         // historical avg - 5%
	Aggressive       float64   `json:"aggressive"`       // historical min
	Recommended      float64   `json:"recommended"`      // best pick based on volatility
	RecommendedLabel string    `json:"recommendedLabel"` // "conservateur", "modéré", "agressif"
	Reasoning        string    `json:"reasoning"`        // French explanation
	HistoricalAvg    float64   `json:"historicalAvg"`
	HistoricalMin    float64   `json:"historicalMin"`
	Volatility       string    `json:"volatility"` // "low", "medium", "high"
	DataPoints       int       `json:"dataPoints"`
	GeneratedAt      time.Time `json:"generatedAt"`
}

// ApplyRequest is the payload for applying a suggestion tier.
type ApplyRequest struct {
	Target string `json:"target"` // "conservative", "moderate", "aggressive"
}
