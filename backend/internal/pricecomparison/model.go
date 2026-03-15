package pricecomparison

type RetailerPrice struct {
	Retailer    string  `json:"retailer"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	IsBestPrice bool    `json:"isBestPrice"`
	DiffPercent float64 `json:"diffPercent"` // vs current item price
}

type Response struct {
	ItemID    string          `json:"itemId"`
	ItemName  string          `json:"itemName"`
	BasePrice float64         `json:"basePrice"`
	Retailers []RetailerPrice `json:"retailers"`
	BestDeal  *RetailerPrice  `json:"bestDeal"`
	MaxSaving float64         `json:"maxSaving"` // basePrice - bestDeal.Price if positive
}
