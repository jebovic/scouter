// Package frenchmarket provides hardcoded French market price intelligence
// for common product categories. No LLM or external API calls are required —
// all data is compiled into the binary.
package frenchmarket

// RetailerPricing describes a French retailer relevant to a product category.
type RetailerPricing struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	TypicalDiscount string `json:"typicalDiscount"`
	LoyaltyProgram  string `json:"loyaltyProgram"`
}

// PriceRange holds the typical min/max retail price in EUR.
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// MarketInsight aggregates French market intelligence for a given product.
type MarketInsight struct {
	ItemName             string            `json:"itemName"`
	Category             string            `json:"category"`
	AveragePriceEUR      float64           `json:"averagePriceEUR"`
	PriceRange           PriceRange        `json:"priceRange"`
	BestRetailers        []RetailerPricing `json:"bestRetailers"`
	BestMonthToBuy       string            `json:"bestMonthToBuy"`
	FlashSaleFrequency   string            `json:"flashSaleFrequency"`
	TipsForFrenchBuyers  []string          `json:"tipsForFrenchBuyers"`
	MarketTrend          string            `json:"marketTrend"`
}
