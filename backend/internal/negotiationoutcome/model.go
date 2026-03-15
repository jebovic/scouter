package negotiationoutcome

// Outcome holds the negotiation result for a single shopping item.
type Outcome struct {
	ItemID          string   `json:"itemId"`
	ItemName        string   `json:"itemName"`
	Merchant        string   `json:"merchant"`
	AskingPrice     float64  `json:"askingPrice"`
	NegotiatedPrice *float64 `json:"negotiatedPrice,omitempty"`
	Saved           float64  `json:"saved"`
	SuccessRate     float64  `json:"successRate"` // 0-1
	TipForMerchant  string   `json:"tipForMerchant"`
	Verdict         string   `json:"verdict"` // "success","partial","failed","pending"
	Attempts        int      `json:"attempts"`
}

// Summary aggregates negotiation outcomes for a mission.
type Summary struct {
	Outcomes      []Outcome `json:"outcomes"`
	TotalSaved    float64   `json:"totalSaved"`
	AvgSuccess    float64   `json:"avgSuccessRate"`
	BestMerchant  string    `json:"bestMerchant"`
	TotalAttempts int       `json:"totalAttempts"`
}
