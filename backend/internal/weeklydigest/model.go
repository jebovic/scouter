package weeklydigest

import "time"

// DigestAlert represents a single price alert item in the weekly digest.
type DigestAlert struct {
	ItemID      string   `json:"itemId"`
	ItemName    string   `json:"itemName"`
	MissionName string   `json:"missionName"`
	MissionSlug string   `json:"missionSlug"`
	OldPrice    float64  `json:"oldPrice"`
	NewPrice    float64  `json:"newPrice"`
	TargetPrice *float64 `json:"targetPrice,omitempty"`
	ChangePct   float64  `json:"changePct"`
	AlertType   string   `json:"alertType"` // "big_drop", "new_low", "target_hit"
}

// WeeklyDigest is the structured response for GET /api/weekly-digest.
type WeeklyDigest struct {
	WeekStart   time.Time     `json:"weekStart"`
	WeekEnd     time.Time     `json:"weekEnd"`
	TotalAlerts int           `json:"totalAlerts"`
	BigDrops    []DigestAlert `json:"bigDrops"`   // >5% price drop
	NewLows     []DigestAlert `json:"newLows"`    // new all-time low for item
	TargetHits  []DigestAlert `json:"targetHits"` // price hit/below target
	Summary     string        `json:"summary"`    // French summary text
	GeneratedAt time.Time     `json:"generatedAt"`
}
