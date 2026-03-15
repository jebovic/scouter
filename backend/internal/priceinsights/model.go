package priceinsights

import "time"

// PriceInsights holds the best-time-to-buy analysis for a shopping item.
type PriceInsights struct {
	ItemID            string     `json:"itemId"`
	DataPoints        int        `json:"dataPoints"`
	BestDayOfWeek     string     `json:"bestDayOfWeek"`
	WorstDayOfWeek    string     `json:"worstDayOfWeek"`
	BestTimeOfMonth   string     `json:"bestTimeOfMonth"`
	PriceByDay        []DayStats `json:"priceByDay"`
	BuyRecommendation string     `json:"buyRecommendation"`
	Confidence        string     `json:"confidence"` // "low", "medium", "high"
	GeneratedAt       time.Time  `json:"generatedAt"`
}

// DayStats holds average price statistics for a given day of the week.
type DayStats struct {
	Day      string  `json:"day"`
	AvgPrice float64 `json:"avgPrice"`
	Count    int     `json:"count"`
}
