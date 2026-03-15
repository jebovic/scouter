// Package priceforecast implements AI-driven price trajectory forecasting for
// shopping items. It uses LLM tool-use to generate a 30-day price forecast
// with confidence bounds and a French-language summary.
package priceforecast

import "time"

// DataPoint is a historical price observation used as LLM input.
type DataPoint struct {
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
}

// ForecastPoint is a single predicted price observation with confidence bounds.
type ForecastPoint struct {
	Date  time.Time `json:"date"`
	Low   float64   `json:"low"`  // lower confidence bound
	Mid   float64   `json:"mid"`  // median prediction
	High  float64   `json:"high"` // upper confidence bound
}

// PriceForecastResult is the full AI price forecast response for a shopping item.
type PriceForecastResult struct {
	ItemID        string          `json:"item_id"`
	ItemName      string          `json:"item_name"`
	Currency      string          `json:"currency"`
	Historical    []DataPoint     `json:"historical"`
	Forecast      []ForecastPoint `json:"forecast"`     // next 30 days
	Confidence    float64         `json:"confidence"`   // 0-1
	Summary       string          `json:"summary"`      // French text
	BestBuyWindow string          `json:"best_buy_window"` // e.g. "dans 2-3 semaines"
	GeneratedAt   time.Time       `json:"generated_at"`
}

// ForecastRequest carries the inputs required by the ForecastAgent.
type ForecastRequest struct {
	ItemID     string
	ItemName   string
	Currency   string
	Historical []DataPoint
	HorizonDays int
}
