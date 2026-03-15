package loyaltypoints

import "time"

// LoyaltyProgram describes a retailer loyalty or cashback scheme.
type LoyaltyProgram struct {
	Name           string  `json:"name"`
	Retailer       string  `json:"retailer"`
	PointsRate     float64 `json:"pointsRate"`     // points earned per euro spent (0 if cashback only)
	CashbackRate   float64 `json:"cashbackRate"`   // percentage of purchase returned (0 if points only)
	MinRedemption  float64 `json:"minRedemption"`  // minimum points/cashback amount before redemption
	Notes          string  `json:"notes"`
}

// UserPoints represents a user's accumulated points in a given program.
type UserPoints struct {
	ProgramName    string     `json:"programName"`
	Retailer       string     `json:"retailer"`
	Points         float64    `json:"points"`
	EstimatedValue float64    `json:"estimatedValue"` // euro value of the accumulated points
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

// CalculateRequest is the body for POST /api/loyalty/calculate.
type CalculateRequest struct {
	Retailer    string  `json:"retailer"`
	Price       float64 `json:"price"`
	ProgramName string  `json:"programName"`
}

// CalculateResponse is returned by POST /api/loyalty/calculate.
type CalculateResponse struct {
	PointsEarned      float64 `json:"pointsEarned"`
	EstimatedCashback float64 `json:"estimatedCashback"` // euro value
	Notes             string  `json:"notes"`
}
