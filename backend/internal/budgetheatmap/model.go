package budgetheatmap

import "time"

// HeatmapCell holds aggregated price_history data for one (category, month) pair.
type HeatmapCell struct {
	Category  string  `json:"category"`
	Month     string  `json:"month"`     // "2026-01"
	Total     float64 `json:"total"`
	Count     int     `json:"count"`
	Intensity float64 `json:"intensity"` // 0-1 normalised against MaxTotal
}

// BudgetHeatmap is the full response returned by GET /api/analytics/budget-heatmap.
type BudgetHeatmap struct {
	Cells       []HeatmapCell `json:"cells"`
	Categories  []string      `json:"categories"`
	Months      []string      `json:"months"`
	MaxTotal    float64       `json:"maxTotal"`
	GeneratedAt time.Time     `json:"generatedAt"`
}
