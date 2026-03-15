package spendingvelocity

type VelocityReport struct {
	MissionID         string  `json:"missionId"`
	TotalBudget       float64 `json:"totalBudget"`
	TotalSpent        float64 `json:"totalSpent"`
	RemainingBudget   float64 `json:"remainingBudget"`
	DaysActive        int     `json:"daysActive"`
	DailyVelocity     float64 `json:"dailyVelocity"`
	ProjectedDaysLeft int     `json:"projectedDaysLeft"`
	PaceStatus        string  `json:"paceStatus"`
	PaceLabel         string  `json:"paceLabel"`
	Advice            string  `json:"advice"`
	BudgetUsedPercent float64 `json:"budgetUsedPercent"`
}
