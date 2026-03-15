package dealaggregator

// Deal represents a single deal from a French deal aggregator site.
type Deal struct {
	Source    string  `json:"source"`
	Title     string  `json:"title"`
	OrigPrice float64 `json:"origPrice"`
	DealPrice float64 `json:"dealPrice"`
	Discount  int     `json:"discount"` // percentage
	URL       string  `json:"url"`
	Hot       bool    `json:"hot"`
	Votes     int     `json:"votes"`
}

// Response is the API response for the deal aggregator endpoint.
type Response struct {
	ItemID   string  `json:"itemId"`
	Deals    []Deal  `json:"deals"`
	BestDeal *Deal   `json:"bestDeal,omitempty"`
}
