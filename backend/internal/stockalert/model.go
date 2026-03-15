package stockalert

// StockStatus represents a simulated stock availability response for a shopping item.
type StockStatus struct {
	ItemID             string  `json:"itemId"`
	Name               string  `json:"name"`
	StockStatus        string  `json:"stockStatus"`
	StockLabel         string  `json:"stockLabel"`
	EstimatedUnitsLeft *int    `json:"estimatedUnitsLeft"`
	RestockDays        *int    `json:"restockDays"`
	Alert              string  `json:"alert"`
	Urgency            string  `json:"urgency"`
}
