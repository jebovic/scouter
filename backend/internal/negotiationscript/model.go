package negotiationscript

// Script holds the four contextual French negotiation sections.
type Script struct {
	Opening     string `json:"opening"`
	Argument    string `json:"argument"`
	CounterOffer string `json:"counterOffer"`
	Closing     string `json:"closing"`
}

// Response is the API response for the negotiation script endpoint.
type Response struct {
	ItemID           string   `json:"itemId"`
	Name             string   `json:"name"`
	Script           Script   `json:"script"`
	Tips             []string `json:"tips"`
	EstimatedSavings float64  `json:"estimatedSavings"`
}
