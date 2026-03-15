// Package prediction provides pure-Go linear regression price forecasting.
// It has no I/O dependencies and is designed for easy unit testing.
package prediction

// PricePrediction is a 30-day forward price forecast for a shopping item.
type PricePrediction struct {
	ItemID          string  `json:"itemId"`
	CurrentPrice    float64 `json:"currentPrice"`
	PredictedPrice  float64 `json:"predictedPrice"`  // 30-day forecast
	ConfidenceLevel string  `json:"confidenceLevel"` // "high" | "medium" | "low"
	Trend           string  `json:"trend"`           // "rising" | "falling" | "stable"
	TrendPct        float64 `json:"trendPct"`        // % change over last 30 days
	DataPoints      int     `json:"dataPoints"`      // number of price history points used
	ForecastDays    int     `json:"forecastDays"`    // always 30
}

// PricePoint is a single price observation identified by its day index
// (0 = earliest known snapshot, N = latest).
type PricePoint struct {
	DayIndex float64
	Price    float64
}
