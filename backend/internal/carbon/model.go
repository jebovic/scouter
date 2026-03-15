// Package carbon provides AI-driven CO2 footprint estimation for shopping items.
// It uses LLM tool-use to estimate total product lifecycle emissions (manufacturing,
// transport, usage, and end-of-life) and returns French-language eco tips.
package carbon

// CarbonEstimate holds the CO2 footprint estimation for a single product.
type CarbonEstimate struct {
	ItemName   string   `json:"itemName"`
	Category   string   `json:"category"`
	KgCO2      float64  `json:"kgCO2"`
	Label      string   `json:"label"`      // "Très faible", "Faible", "Moyen", "Élevé", "Très élevé"
	Tips       []string `json:"tips"`       // 2-3 eco tips in French
	Confidence string   `json:"confidence"` // "high", "medium", "low"
}

// EstimateRequest carries the inputs required by the carbon estimation endpoint.
type EstimateRequest struct {
	ItemName string `json:"itemName"`
	Category string `json:"category"`
}
