package ecoscore

// EcoScore is the mission-level sustainability summary.
type EcoScore struct {
	MissionID               string         `json:"missionId"`
	TotalCO2kg              float64        `json:"totalCO2kg"`
	EcoGrade                string         `json:"ecoGrade"` // A|B|C|D|E
	GradeColor              string         `json:"gradeColor"`
	ItemScores              []ItemEcoScore `json:"itemScores"`
	SustainableShopping     []string       `json:"sustainableShopping"`
	EcoTips                 []string       `json:"ecoTips"`
	CompensationTreesNeeded int            `json:"compensationTreesNeeded"`
}

// ItemEcoScore is the CO2 estimate for a single shopping item.
type ItemEcoScore struct {
	ItemID   string  `json:"itemId"`
	ItemName string  `json:"itemName"`
	Category string  `json:"category"`
	CO2kg    float64 `json:"co2kg"`
	Grade    string  `json:"grade"`
}
