package decisionmatrix

// ItemScores holds the individual criterion scores for a shopping item.
type ItemScores struct {
	Prix         int `json:"prix"`
	Urgence      int `json:"urgence"`
	Valeur       int `json:"valeur"`
	Disponibilite int `json:"disponibilite"`
	Confiance    int `json:"confiance"`
}

// RankedItem is a shopping item with its decision matrix scores and rank.
type RankedItem struct {
	ItemID         string     `json:"itemId"`
	Name           string     `json:"name"`
	Scores         ItemScores `json:"scores"`
	TotalScore     int        `json:"totalScore"`
	Recommendation string     `json:"recommendation"`
	Rank           int        `json:"rank"`
}

// MatrixResponse is the full decision matrix response for a mission.
type MatrixResponse struct {
	Items             []RankedItem `json:"items"`
	TopRecommendation *RankedItem  `json:"topRecommendation"`
}

// recommendation returns a French recommendation string based on the total score.
func recommendation(score int) string {
	switch {
	case score >= 80:
		return "Achetez maintenant"
	case score >= 60:
		return "Très bon moment pour acheter"
	case score >= 40:
		return "À surveiller de près"
	case score >= 20:
		return "Attendez"
	default:
		return "Différez l'achat"
	}
}
