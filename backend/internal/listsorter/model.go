package listsorter

import "github.com/google/uuid"

// RankedItem is a shopping item enriched with a sort score and explanatory reason.
type RankedItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	Score       int       `json:"score"`
	Reason      string    `json:"reason"`
	Rank        int       `json:"rank"`
}

// SortedResponse is the JSON envelope returned by the sorted-items endpoint.
type SortedResponse struct {
	Items    []RankedItem `json:"items"`
	SortedBy string       `json:"sortedBy"`
}

// rawItem is the DB row fetched before scoring.
type rawItem struct {
	ID          uuid.UUID
	Name        string
	Price       float64
	Status      string
	TargetPrice *float64
	Merchant    string
	CostCategory string
}
