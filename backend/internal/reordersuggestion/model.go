// Package reordersuggestion provides smart purchase order recommendations
// for shopping items within a mission. Suggestions are purely deterministic
// (no LLM calls) and cached in-process for 10 minutes.
package reordersuggestion

// Suggestion holds the buy-now recommendation for a single shopping item.
type Suggestion struct {
	ItemID    string  `json:"itemId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	BuyScore  int     `json:"buyScore"`
	Priority  string  `json:"priority"`  // "immédiat" | "cette semaine" | "ce mois" | "flexible"
	Reasoning string  `json:"reasoning"`
	Merchant  string  `json:"merchant"`
}

// MerchantBatch groups items from the same merchant to suggest combined ordering.
type MerchantBatch struct {
	Merchant  string   `json:"merchant"`
	Items     []string `json:"items"`
	BatchNote string   `json:"batchNote"`
}

// Response is the JSON envelope for GET /api/missions/{missionId}/reorder-suggestions.
type Response struct {
	MissionID       string          `json:"missionId"`
	Suggestions     []Suggestion    `json:"suggestions"`
	MerchantBatches []MerchantBatch `json:"merchantBatches"`
	TotalToBuy      float64         `json:"totalToBuy"`
}

// rawItem is the DB row for a shopping item.
type rawItem struct {
	id           string
	name         string
	price        float64
	targetPrice  *float64
	status       string
	costCategory string
	merchant     string
}
