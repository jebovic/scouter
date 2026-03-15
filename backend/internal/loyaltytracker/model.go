package loyaltytracker

// loyaltyProgram holds static French merchant loyalty program data.
type loyaltyProgram struct {
	Program     string
	PointRate   float64
	Tier        string
	CashbackPct float64
}

// defaultProgram is returned for merchants not in the registry.
var defaultProgram = loyaltyProgram{
	Program:     "Aucun programme",
	PointRate:   0,
	Tier:        "",
	CashbackPct: 0,
}

// merchantRegistry maps well-known French merchant names to their loyalty programs.
var merchantRegistry = map[string]loyaltyProgram{
	"Fnac": {
		Program:     "Fnac+",
		PointRate:   0.01,
		Tier:        "Adhérent",
		CashbackPct: 2.0,
	},
	"Darty": {
		Program:     "Club Darty",
		PointRate:   0.01,
		Tier:        "Standard",
		CashbackPct: 1.5,
	},
	"Amazon": {
		Program:     "Prime",
		PointRate:   0.005,
		Tier:        "Prime",
		CashbackPct: 0.5,
	},
	"Cdiscount": {
		Program:     "Cdiscount à volonté",
		PointRate:   0.02,
		Tier:        "Standard",
		CashbackPct: 3.0,
	},
	"Boulanger": {
		Program:     "Club Boulanger",
		PointRate:   0.015,
		Tier:        "Standard",
		CashbackPct: 2.5,
	},
	"Leclerc": {
		Program:     "Carte E.Leclerc",
		PointRate:   0.02,
		Tier:        "Standard",
		CashbackPct: 3.0,
	},
	"Carrefour": {
		Program:     "Carrefour +",
		PointRate:   0.01,
		Tier:        "Standard",
		CashbackPct: 2.0,
	},
	"La Redoute": {
		Program:     "Club La Redoute",
		PointRate:   0.01,
		Tier:        "Standard",
		CashbackPct: 1.0,
	},
}

// MerchantSummary holds the loyalty data for a single merchant.
type MerchantSummary struct {
	Merchant          string  `json:"merchant"`
	Total             float64 `json:"total"`
	Program           string  `json:"program"`
	Tier              string  `json:"tier,omitempty"`
	EstimatedCashback float64 `json:"estimatedCashback"`
	EstimatedPoints   int64   `json:"estimatedPoints"`
}

// LoyaltySummary is the API response for GET /api/missions/{missionId}/loyalty-summary.
type LoyaltySummary struct {
	Merchants             []MerchantSummary `json:"merchants"`
	TotalEstimatedCashback float64          `json:"totalEstimatedCashback"`
}
