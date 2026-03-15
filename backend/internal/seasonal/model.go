// Package seasonal provides AI-driven seasonal discount prediction for the French market.
// It uses LLM tool-use to predict the best months/seasons to buy a product,
// expected discount percentage, and relevant shopping events.
package seasonal

// Prediction holds the seasonal discount prediction for a product.
type Prediction struct {
	ItemName       string   `json:"itemName"`
	Category       string   `json:"category"`
	BestMonths     []string `json:"bestMonths"`    // e.g. ["Novembre", "Décembre"]
	AvgDiscountPct float64  `json:"avgDiscountPct"` // e.g. 25.0
	Events         []string `json:"events"`         // e.g. ["Black Friday", "Soldes d'hiver"]
	NextEventDays  int      `json:"nextEventDays"`  // days until next best buying opportunity
	Rationale      string   `json:"rationale"`      // 1-2 sentences in French
	Confidence     string   `json:"confidence"`     // "high", "medium", "low"
}

// PredictRequest carries the inputs required by the seasonal prediction endpoint.
type PredictRequest struct {
	ItemName string `json:"itemName"`
	Category string `json:"category"`
}
