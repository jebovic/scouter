package burnrate

// DataPoint represents a single cumulative spend data point.
type DataPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// Response is the burn rate analysis for a mission.
type Response struct {
	MissionID      string      `json:"missionId"`
	TotalBudget    float64     `json:"totalBudget"`
	TotalSpent     float64     `json:"totalSpent"`
	DailyRate      float64     `json:"dailyRate"`      // avg spend per day
	ProjectedTotal float64     `json:"projectedTotal"` // extrapolated to 30 days
	DaysElapsed    int         `json:"daysElapsed"`
	BurnPoints     []DataPoint `json:"burnPoints"` // cumulative spend over time
	Status         string      `json:"status"`     // "on_track", "over_budget", "at_risk"
}
