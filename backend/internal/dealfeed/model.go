package dealfeed

import "time"

// PromoOffer represents a single French market promotional offer.
type PromoOffer struct {
	Title         string     `json:"title"`
	Retailer      string     `json:"retailer"`
	OriginalPrice float64    `json:"originalPrice"`
	PromoPrice    float64    `json:"promoPrice"`
	Discount      int        `json:"discount"` // percentage
	URL           string     `json:"url"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	Category      string     `json:"category"`
	IsFlash       bool       `json:"isFlash"`
	Source        string     `json:"source"` // "ai_generated" or "api"
}

// PromoFeedResult is the top-level response returned by the promo feed endpoint.
type PromoFeedResult struct {
	Offers    []PromoOffer `json:"offers"`
	Query     string       `json:"query"`
	Category  string       `json:"category"`
	FetchedAt time.Time    `json:"fetchedAt"`
}
