package competitorprice

import "time"

// CompetitorPrices aggregates simulated competitor price offers for a shopping item.
type CompetitorPrices struct {
	ItemID      string            `json:"itemId"`
	ItemName    string            `json:"itemName"`
	BasePrice   float64           `json:"basePrice"`
	Competitors []CompetitorOffer `json:"competitors"`
	BestOffer   *CompetitorOffer  `json:"bestOffer,omitempty"`
	Savings     float64           `json:"savings"`
	GeneratedAt time.Time         `json:"generatedAt"`
}

// CompetitorOffer represents a single platform's price for the item.
type CompetitorOffer struct {
	Platform     string  `json:"platform"`
	Price        float64 `json:"price"`
	PriceDiff    float64 `json:"priceDiff"`
	PriceDiffPct float64 `json:"priceDiffPct"`
	URL          string  `json:"url"`
	IsBest       bool    `json:"isBest"`
	Availability string  `json:"availability"`
}
