package missionreport

// missionRow holds the mission fields needed for the report.
type missionRow struct {
	name   string
	budget float64
	slug   string
}

// itemRow holds the shopping item fields needed for the report.
type itemRow struct {
	name         string
	price        float64
	targetPrice  *float64
	status       string
	merchant     string
	costCategory string
}
