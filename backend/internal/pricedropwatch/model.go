package pricedropwatch

// WatchItem represents a single item being watched for price drops
type WatchItem struct {
	ItemID          string  `json:"itemId"`
	Name            string  `json:"name"`
	CurrentPrice    float64 `json:"currentPrice"`
	DropProbability float64 `json:"dropProbability"` // 0-100 percentage
	Signal          string  `json:"signal"`          // high_volatility, recent_spike, no_history, stable
	ExpectedDrop    float64 `json:"expectedDrop"`    // percentage
	Reasoning       string  `json:"reasoning"`       // French explanation
}

// WatchlistResponse contains the watchlist summary and items
type WatchlistResponse struct {
	Summary   string      `json:"summary"`   // French summary
	WatchItems []WatchItem `json:"watchItems"`
}
