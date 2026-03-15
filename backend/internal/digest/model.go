package digest

import "time"

// PriceDropItem represents a single shopping item whose price decreased
// compared to its price N days ago.
type PriceDropItem struct {
	MissionID   string  `json:"missionId"`
	MissionName string  `json:"missionName"`
	MissionSlug string  `json:"missionSlug"`
	ItemID      string  `json:"itemId"`
	ItemName    string  `json:"itemName"`
	Merchant    string  `json:"merchant"`
	PriceBefore float64 `json:"priceBefore"`
	PriceNow    float64 `json:"priceNow"`
	DropPct     float64 `json:"dropPct"`
	Currency    string  `json:"currency"`
}

// WeeklyDigest is the structured response returned by GET /api/digest/weekly.
type WeeklyDigest struct {
	GeneratedAt  time.Time       `json:"generatedAt"`
	PeriodDays   int             `json:"periodDays"`
	TotalDrops   int             `json:"totalDrops"`
	TotalSavings float64         `json:"totalSavings"`
	Currency     string          `json:"currency"`
	Items        []PriceDropItem `json:"items"`
}
