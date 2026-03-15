package expensecategorizer

type CategoryBreakdown struct {
	Category   string  `json:"category"`
	Label      string  `json:"label"`      // French label
	TotalSpent float64 `json:"totalSpent"`
	ItemCount  int     `json:"itemCount"`
	Percentage float64 `json:"percentage"` // % of total spent
	Icon       string  `json:"icon"`
}

type ExpenseSummary struct {
	MissionID   string              `json:"missionId"`
	TotalItems  int                 `json:"totalItems"`
	TotalSpent  float64             `json:"totalSpent"`
	Breakdowns  []CategoryBreakdown `json:"breakdowns"` // sorted by totalSpent desc
	TopCategory string              `json:"topCategory"`
	Insight     string              `json:"insight"` // French insight string
}
