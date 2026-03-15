package budgetrec

// Priority indicates the urgency of a budget recommendation.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// BudgetRecommendation is a single actionable suggestion for a spending category.
type BudgetRecommendation struct {
	Category          string   `json:"category"`
	CurrentBudget     float64  `json:"currentBudget"`
	RecommendedBudget float64  `json:"recommendedBudget"`
	Reason            string   `json:"reason"`
	Priority          Priority `json:"priority"`
	SaveAmount        float64  `json:"saveAmount"`
}

// BudgetAnalysis is the full result returned by the analysis endpoint.
type BudgetAnalysis struct {
	MissionID        string                 `json:"missionId"`
	TotalBudget      float64                `json:"totalBudget"`
	TotalSpent       float64                `json:"totalSpent"`
	Recommendations  []BudgetRecommendation `json:"recommendations"`
	PotentialSavings float64                `json:"potentialSavings"`
	HealthScore      int                    `json:"healthScore"` // 0–100
}
