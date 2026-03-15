package priceanalytics

// TrendDirection describes the direction of price movement.
type TrendDirection string

const (
	TrendRising  TrendDirection = "rising"
	TrendFalling TrendDirection = "falling"
	TrendStable  TrendDirection = "stable"
)

// PriceStats holds computed analytics for a shopping item's price history.
// Fields used by the per-item endpoint (GetStats) keep their original names
// for backward compatibility; mission-level fields are added inline.
type PriceStats struct {
	// Per-item endpoint fields (GetStats).
	ItemID      string  `json:"itemId"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Average     float64 `json:"average"`
	Volatility  float64 `json:"volatility"`  // standard deviation
	DataPoints  int     `json:"dataPoints"`
	TrendSlope  float64 `json:"trendSlope"`  // positive = rising, negative = falling
	PriceDrop7d float64 `json:"priceDrop7d"` // change vs 7 days ago, negative = cheaper

	// Mission-level analytics fields (GetAnalytics).
	ItemName      string         `json:"itemName,omitempty"`
	Current       float64        `json:"current,omitempty"`
	VolatilityPct float64        `json:"volatilityPct,omitempty"` // stddev/avg*100
	Trend         TrendDirection `json:"trend,omitempty"`
}

// MissionPriceAnalytics aggregates price analytics across all items for a mission.
type MissionPriceAnalytics struct {
	MissionID    string       `json:"missionId"`
	Items        []PriceStats `json:"items"`
	BestDealItem string       `json:"bestDealItem"` // item with the most negative (current-avg)/avg
	MostVolatile string       `json:"mostVolatile"` // item with highest volatilityPct
}
