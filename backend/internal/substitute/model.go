// Package substitute implements the AI Product Substitute Finder agent.
// It suggests cheaper or better alternative products for any shopping option,
// focused on French market availability.
package substitute

import "time"

// Substitute represents a single alternative product suggestion.
type Substitute struct {
	Name      string  `json:"name"`
	Brand     string  `json:"brand"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Retailer  string  `json:"retailer"`  // French retailer name
	Reason    string  `json:"reason"`    // Why it's a good substitute
	Advantage string  `json:"advantage"` // "cheaper" | "better_rated" | "eco_friendly" | "local"
	URL       string  `json:"url,omitempty"`

	// Phase 79 enrichment fields — populated when available.
	SavingsPercent float64  `json:"savingsPercent,omitempty"` // savings vs original price
	Pros           []string `json:"pros,omitempty"`           // up to 3 bullet points (French)
	Cons           []string `json:"cons,omitempty"`           // up to 2 bullet points (French)
	WhyConsider    string   `json:"whyConsider,omitempty"`    // 1 sentence rationale (French)
}

// SubstituteResponse is the full response returned by the handler.
type SubstituteResponse struct {
	OptionID      string       `json:"optionId"`
	ProductName   string       `json:"productName"`
	OriginalPrice float64      `json:"originalPrice,omitempty"`
	Substitutes   []Substitute `json:"substitutes"`
	GeneratedAt   time.Time    `json:"generatedAt"`
	CachedAt      int64        `json:"cachedAt"` // unix timestamp, 0 when freshly generated
}
