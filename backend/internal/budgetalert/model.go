package budgetalert

// BudgetAlert represents a budget threshold alert for a single mission.
type BudgetAlert struct {
	MissionID       string   `json:"missionId"`
	MissionName     string   `json:"missionName"`
	TotalBudget     float64  `json:"totalBudget"`
	TotalSpent      float64  `json:"totalSpent"`
	Percentage      float64  `json:"percentage"`
	AlertLevel      string   `json:"alertLevel"` // "info" | "warning" | "critical"
	Message         string   `json:"message"`
	Recommendations []string `json:"recommendations"`
}

// BudgetAlertList is the response envelope for the GET /api/budget-alerts endpoint.
type BudgetAlertList struct {
	Alerts          []BudgetAlert `json:"alerts"`
	MissionsChecked int           `json:"missionsChecked"`
}
