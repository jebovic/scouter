package timingrecommender

// Signal is a single scoring input to the composite timing score.
type Signal struct {
	Name   string  `json:"name"`
	Score  int     `json:"score"`   // 0–100 contribution
	Weight float64 `json:"weight"`
	Label  string  `json:"label"`   // human-readable French text
}

// Response is the API payload returned by GET /api/missions/{missionId}/items/{itemId}/timing-score.
type Response struct {
	ItemID         string   `json:"itemId"`
	ItemName       string   `json:"itemName"`
	TimingScore    int      `json:"timingScore"`    // 0–100 composite
	Recommendation string   `json:"recommendation"` // "buy_now", "watch", "wait"
	Verdict        string   `json:"verdict"`        // French verdict text
	Signals        []Signal `json:"signals"`
	BestBuyWindow  string   `json:"bestBuyWindow"`  // e.g. "Soldes d'été (juin)"
	UrgencyLevel   string   `json:"urgencyLevel"`   // "low", "medium", "high", "critical"
}
