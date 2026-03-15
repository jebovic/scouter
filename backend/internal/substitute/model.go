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
}

// SubstituteResponse is the full response returned by the handler.
type SubstituteResponse struct {
	OptionID    string       `json:"optionId"`
	ProductName string       `json:"productName"`
	Substitutes []Substitute `json:"substitutes"`
	GeneratedAt time.Time    `json:"generatedAt"`
}
