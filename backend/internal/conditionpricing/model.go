package conditionpricing

import "time"

// ConditionPrice holds pricing info for a single refurbished grade/condition.
type ConditionPrice struct {
	Condition      string  `json:"condition"`      // e.g. "Comme neuf"
	Grade          string  `json:"grade"`          // "A+", "A", "B", "C"
	EstPrice       float64 `json:"estPrice"`
	Discount       float64 `json:"discount"`       // % off vs new price
	Platform       string  `json:"platform"`       // e.g. "Back Market"
	Availability   string  `json:"availability"`   // "Disponible", "Stock limité", "Indisponible"
	WarrantyMonths int     `json:"warrantyMonths"`
	SearchUrl      string  `json:"searchUrl"`
}

// ConditionReport is the full response for a condition-pricing request.
type ConditionReport struct {
	ItemId         string           `json:"itemId"`
	ItemName       string           `json:"itemName"`
	NewPrice       float64          `json:"newPrice"`
	Conditions     []ConditionPrice `json:"conditions"`
	BestDeal       *ConditionPrice  `json:"bestDeal,omitempty"`
	Recommendation string           `json:"recommendation"` // French
	GeneratedAt    time.Time        `json:"generatedAt"`
}
