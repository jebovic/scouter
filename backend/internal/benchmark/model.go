package benchmark

// PriceRange holds the low/average/high market price estimates for a product.
type PriceRange struct {
	Low     float64 `json:"low"`
	Average float64 `json:"average"`
	High    float64 `json:"high"`
}

// BenchmarkResult is the full market benchmark response for a shopping item.
type BenchmarkResult struct {
	ItemID       string     `json:"itemId"`
	ItemName     string     `json:"itemName"`
	CurrentPrice float64    `json:"currentPrice"`
	Currency     string     `json:"currency"`
	MarketRange  PriceRange `json:"marketRange"`
	// Verdict is one of "good_deal", "average", or "overpriced".
	Verdict string `json:"verdict"`
	// DiffPercent is (currentPrice - marketAverage) / marketAverage * 100.
	// Negative means below average (good deal), positive means above average.
	DiffPercent float64 `json:"diffPercent"`
	Explanation string  `json:"explanation"`
	// CachedAt is the Unix timestamp (seconds) when the result was computed.
	CachedAt int64 `json:"cachedAt"`
}
