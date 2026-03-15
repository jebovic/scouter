package currency

// Rates holds exchange rates with EUR as base currency.
type Rates struct {
	Base  string             `json:"base"` // always "EUR"
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

// ConvertResponse is the response body for the /convert endpoint.
type ConvertResponse struct {
	Result float64 `json:"result"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}
