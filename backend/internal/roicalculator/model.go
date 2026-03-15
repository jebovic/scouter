package roicalculator

import "time"

// MissionROI holds the computed return-on-research-investment for a mission.
type MissionROI struct {
	MissionID          string    `json:"missionId"`
	TotalSpend         float64   `json:"totalSpend"`
	TotalEstimate      float64   `json:"totalEstimate"`
	TotalTargetSavings float64   `json:"totalTargetSavings"`
	SavingsVsEstimate  float64   `json:"savingsVsEstimate"`
	SavingsPercent     float64   `json:"savingsPercent"`
	PriceChecks        int       `json:"priceChecks"`
	DaysTracking       int       `json:"daysTracking"`
	AvgChecksPerDay    float64   `json:"avgChecksPerDay"`
	ROIScore           float64   `json:"roiScore"`
	Verdict            string    `json:"verdict"`
	GeneratedAt        time.Time `json:"generatedAt"`
}

// verdict returns a French verdict string based on the ROI score.
func verdictFor(score float64) string {
	switch {
	case score >= 80:
		return "Excellent travail de prospection !"
	case score >= 50:
		return "Bonne recherche de prix"
	case score >= 20:
		return "Quelques économies réalisées"
	case score >= 0:
		return "Continuez à surveiller les prix"
	default:
		return "Les prix ont augmenté depuis l'estimation"
	}
}
