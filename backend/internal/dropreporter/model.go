package dropreporter

import "time"

// DropPrediction holds the price-drop probability forecast for a shopping item.
type DropPrediction struct {
	ItemID              string    `json:"itemId"`
	Probability         float64   `json:"probability"`          // 0.0–1.0
	PredictedDropPct    float64   `json:"predictedDropPercent"` // estimated % drop
	Horizon             int       `json:"horizon"`              // days: 7, 14, or 30
	Confidence          string    `json:"confidence"`           // "faible", "moyen", "élevé"
	Signals             []string  `json:"signals"`
	GeneratedAt         time.Time `json:"generatedAt"`
}
