package pricedigest

import "time"

// PriceChange represents a single shopping item whose price changed in the last 24 hours.
type PriceChange struct {
	ItemID       string  `json:"itemId"`
	ItemName     string  `json:"itemName"`
	Merchant     string  `json:"merchant"`
	MissionName  string  `json:"missionName"`
	MissionSlug  string  `json:"missionSlug"`
	OldPrice     float64 `json:"oldPrice"`
	NewPrice     float64 `json:"newPrice"`
	ChangeAmount float64 `json:"changeAmount"`
	ChangePct    float64 `json:"changePct"`
	Status       string  `json:"status"`
	IsDropping   bool    `json:"isDropping"`
}

// Digest is the response returned by GET /api/price-digest.
type Digest struct {
	GeneratedAt       time.Time     `json:"generatedAt"`
	TotalItemsChecked int           `json:"totalItemsChecked"`
	PriceDrops        []PriceChange `json:"priceDrops"`
	PriceRises        []PriceChange `json:"priceRises"`
	StableItems       int           `json:"stableItems"`
	BiggestDrop       *PriceChange  `json:"biggestDrop"`
	BiggestRise       *PriceChange  `json:"biggestRise"`
	Summary           string        `json:"summary"`
}
