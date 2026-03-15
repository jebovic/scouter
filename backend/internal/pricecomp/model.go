// Package pricecomp implements the PriceCompAgent, which uses LLM tool-use to
// simulate price comparison data across French retailers for a product option.
// It owns prompt construction, tool schema definition, and response parsing.
// The llm.Provider is transport-only.
package pricecomp

// RetailerOffer represents a single retailer's offer for a product.
type RetailerOffer struct {
	Retailer  string  `json:"retailer"`  // e.g. "Fnac", "Cdiscount", "Amazon FR"
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`  // "EUR"
	Available bool    `json:"available"`
	URL       string  `json:"url"`       // product URL (may be empty/placeholder)
	UpdatedAt string  `json:"updatedAt"` // RFC3339
}

// PriceComparison is the structured output returned by the PriceCompAgent and
// cached by the Handler.
type PriceComparison struct {
	OptionID  string          `json:"optionId"`
	Product   string          `json:"product"`
	Offers    []RetailerOffer `json:"offers"`    // sorted by price asc, available first
	LowestIdx int             `json:"lowestIdx"` // index of cheapest available offer
	FetchedAt string          `json:"fetchedAt"` // RFC3339
}
