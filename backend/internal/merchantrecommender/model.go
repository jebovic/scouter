package merchantrecommender

type MerchantScore struct {
	Merchant       string  `json:"merchant"`
	Score          float64 `json:"score"`
	AvgPrice       float64 `json:"avgPrice"`
	PriceCount     int     `json:"priceCount"`
	Recommendation string  `json:"recommendation"`
}

type Response struct {
	ItemID   string          `json:"itemId"`
	ItemName string          `json:"itemName"`
	Rankings []MerchantScore `json:"rankings"`
}
