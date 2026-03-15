package pricestreak

import "time"

// PriceStreak holds the consecutive price-drop (or rise) streak analysis
// for a shopping item.
type PriceStreak struct {
	ItemID           string    `json:"itemId"`
	CurrentStreak    int       `json:"currentStreak"`    // consecutive price changes (+ or -)
	StreakType       string    `json:"streakType"`       // "dropping", "rising", "stable"
	TotalDropPct     float64   `json:"totalDropPct"`     // cumulative % change during streak
	TotalDropAmt     float64   `json:"totalDropAmt"`     // cumulative € change
	StreakStart       time.Time `json:"streakStart"`
	StreakStartPrice  float64   `json:"streakStartPrice"`
	Recommendation   string    `json:"recommendation"` // French advice
	Urgency          string    `json:"urgency"`        // "low", "medium", "high", "urgent"
	GeneratedAt      time.Time `json:"generatedAt"`
}
