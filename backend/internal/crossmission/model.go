package crossmission

import "time"

// MissionComparison holds computed analytics for one mission.
type MissionComparison struct {
	MissionID     string  `json:"missionId"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Budget        float64 `json:"budget"`
	ItemCount     int     `json:"itemCount"`
	TotalPrice    float64 `json:"totalPrice"`
	TotalEstimate float64 `json:"totalEstimate"`
	BoughtCount   int     `json:"boughtCount"`
	SavingsAmount float64 `json:"savingsAmount"`
	SavingsPct    float64 `json:"savingsPct"`
	BudgetUsedPct float64 `json:"budgetUsedPct"`
	Efficiency    string  `json:"efficiency"`
}

// CrossMissionResponse is the envelope returned by the handler.
type CrossMissionResponse struct {
	Missions               []MissionComparison `json:"missions"`
	BestMission            *MissionComparison  `json:"bestMission"`
	TotalSavingsAllMissions float64            `json:"totalSavingsAllMissions"`
	GeneratedAt            time.Time           `json:"generatedAt"`
}
