package pricealertrule

import "time"

// RuleType defines the kind of alert condition.
type RuleType string

const (
	RuleDropPercent RuleType = "drop_percent"
	RuleFixedBelow  RuleType = "fixed_below"
	RuleWeeklyLow   RuleType = "weekly_low"
)

// AlertRule represents a single price alert rule for a shopping item.
type AlertRule struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"itemId"`
	Type      RuleType  `json:"type"`
	Value     float64   `json:"value"`     // percentage for drop_percent, price for fixed_below, ignored for weekly_low
	BasePrice float64   `json:"basePrice"` // price when rule was created (for drop_percent)
	CreatedAt time.Time `json:"createdAt"`
	Triggered bool      `json:"triggered"`
}

// CheckResult is the outcome of evaluating a single rule against a current price.
type CheckResult struct {
	ItemID    string `json:"itemId"`
	RuleID    string `json:"ruleId"`
	Triggered bool   `json:"triggered"`
	Message   string `json:"message"`
}
