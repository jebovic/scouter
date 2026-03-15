package travel

import "time"

// FlightOffer represents a flight option from Aviationstack.
type FlightOffer struct {
	Airline   string    `json:"airline"`
	FlightNum string    `json:"flightNum"`
	Departure string    `json:"departure"` // IATA code
	Arrival   string    `json:"arrival"`
	DepartAt  time.Time `json:"departAt"`
	ArriveAt  time.Time `json:"arriveAt"`
	PriceEUR  float64   `json:"priceEur"`
	Currency  string    `json:"currency"`
	Source    string    `json:"source"` // "aviationstack"
}

// TrainOffer represents a train journey from SNCF/Navitia.
type TrainOffer struct {
	Operator  string    `json:"operator"`
	TrainNum  string    `json:"trainNum"`
	Departure string    `json:"departure"` // station name
	Arrival   string    `json:"arrival"`
	DepartAt  time.Time `json:"departAt"`
	ArriveAt  time.Time `json:"arriveAt"`
	Duration  int       `json:"durationMin"` // minutes
	PriceEUR  float64   `json:"priceEur"`
	Source    string    `json:"source"` // "sncf"
}
