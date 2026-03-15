package comparisonscore

type ItemScore struct {
	ItemID         string  `json:"itemId"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
	PriceScore     float64 `json:"priceScore"`    // 0-100, higher = better price vs budget
	UrgencyScore   float64 `json:"urgencyScore"`  // 0-100 based on status
	GapScore       float64 `json:"gapScore"`      // 0-100, how close to target price
	CompositeScore float64 `json:"compositeScore"` // weighted average
	Rank           int     `json:"rank"`           // 1=best overall
	Verdict        string  `json:"verdict"`        // French verdict
}

type ComparisonReport struct {
	MissionID  string      `json:"missionId"`
	ItemScores []ItemScore `json:"itemScores"` // sorted by compositeScore desc
	TopPick    string      `json:"topPick"`    // name of #1 item
	Summary    string      `json:"summary"`    // French summary
}
