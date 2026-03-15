package spendinganalytics

import "time"

// SpendingAnalytics holds cross-mission spending aggregates.
type SpendingAnalytics struct {
	TotalBudget    float64         `json:"totalBudget"`
	TotalSpending  float64         `json:"totalSpending"`
	BudgetUsagePct float64         `json:"budgetUsagePct"`
	ByCategory     []CategoryStats `json:"byCategory"`
	ByMerchant     []MerchantStats `json:"byMerchant"`
	TopItems       []TopItem       `json:"topItems"`
	MonthlySummary []MonthStats    `json:"monthlySummary"`
	GeneratedAt    time.Time       `json:"generatedAt"`
}

// CategoryStats aggregates shopping_items with status='buy' by cost_category.
type CategoryStats struct {
	Category  string  `json:"category"`
	Total     float64 `json:"total"`
	ItemCount int     `json:"itemCount"`
	AvgPrice  float64 `json:"avgPrice"`
	Pct       float64 `json:"pct"`
}

// MerchantStats aggregates shopping_items with status='buy' by merchant.
type MerchantStats struct {
	Merchant  string  `json:"merchant"`
	Total     float64 `json:"total"`
	ItemCount int     `json:"itemCount"`
}

// TopItem represents the most expensive buy-status shopping items.
type TopItem struct {
	ItemID      string  `json:"itemId"`
	ItemName    string  `json:"itemName"`
	MissionName string  `json:"missionName"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// MonthStats aggregates price_history entries by calendar month.
type MonthStats struct {
	Month      string  `json:"month"`      // "2026-01"
	AvgPrice   float64 `json:"avgPrice"`
	DataPoints int     `json:"dataPoints"`
}
