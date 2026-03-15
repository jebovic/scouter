package budgetplanner

import "time"

// PurchasePlan is the sequenced purchase plan for a mission.
type PurchasePlan struct {
	MissionID     string        `json:"missionId"`
	TotalBudget   float64       `json:"totalBudget"`
	TotalEstimate float64       `json:"totalEstimate"`
	BudgetStatus  string        `json:"budgetStatus"` // "under", "over", "tight"
	Sequence      []PlannedItem `json:"sequence"`
	Deferred      []PlannedItem `json:"deferred"`
	Savings       float64       `json:"savings"`
	GeneratedAt   time.Time     `json:"generatedAt"`
}

// PlannedItem is a single item placed in the purchase sequence or deferred list.
type PlannedItem struct {
	ItemID         string  `json:"itemId"`
	ItemName       string  `json:"itemName"`
	Price          float64 `json:"price"`
	Status         string  `json:"status"`
	Priority       int     `json:"priority"` // computed score
	Reason         string  `json:"reason"`   // French explanation
	CumulativeCost float64 `json:"cumulativeCost"`
}
