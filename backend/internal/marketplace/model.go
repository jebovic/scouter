package marketplace

type MarketplaceResult struct {
	Site      string  `json:"site"`
	URL       string  `json:"url"`
	Price     float64 `json:"price,omitempty"`
	Available bool    `json:"available"`
	Label     string  `json:"label"` // "Neuf", "Occasion", etc.
}

type MarketplaceResponse struct {
	ItemName string              `json:"itemName"`
	Results  []MarketplaceResult `json:"results"`
}
