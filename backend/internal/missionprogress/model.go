package missionprogress

import "time"

// MissionProgress is a compact snapshot of how far through a mission the user is.
type MissionProgress struct {
	MissionID      string      `json:"missionId"`
	TotalItems     int         `json:"totalItems"`
	BoughtItems    int         `json:"boughtItems"`
	WatchingItems  int         `json:"watchingItems"`
	DeferredItems  int         `json:"deferredItems"`
	FlashSaleItems int         `json:"flashSaleItems"`
	CompletionPct  float64     `json:"completionPct"`  // bought/total*100
	BudgetUsed     float64     `json:"budgetUsed"`     // sum of bought-item prices
	BudgetTotal    float64     `json:"budgetTotal"`    // mission budget
	BudgetPct      float64     `json:"budgetPct"`      // budgetUsed/budgetTotal*100
	TopSaving      *ItemSaving `json:"topSaving,omitempty"` // biggest deal
	GeneratedAt    time.Time   `json:"generatedAt"`
}

// ItemSaving describes the best price-vs-target saving found in the mission.
type ItemSaving struct {
	ItemName     string  `json:"itemName"`
	CurrentPrice float64 `json:"currentPrice"`
	TargetPrice  float64 `json:"targetPrice"`
	Saving       float64 `json:"saving"`
}
