package elasticity

// ElasticityResult is the response for GET /api/missions/{missionId}/items/{itemId}/elasticity.
type ElasticityResult struct {
	ItemID               string   `json:"itemId"`
	Elasticity           *float64 `json:"elasticity"`
	Classification       string   `json:"classification"`
	Insight              string   `json:"insight"`
	OptimalBuyThreshold  *float64 `json:"optimalBuyThreshold"`
	PricePoints          int      `json:"pricePoints"`
}
