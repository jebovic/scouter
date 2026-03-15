package activityfeed

import "time"

// Activity represents a single event in the activity feed.
type Activity struct {
	Type        string    `json:"type"` // "price_drop", "price_rise", "item_added", "mission_created"
	MissionID   string    `json:"missionId"`
	MissionName string    `json:"missionName"`
	MissionSlug string    `json:"missionSlug"`
	ItemID      string    `json:"itemId,omitempty"`
	ItemName    string    `json:"itemName,omitempty"`
	Description string    `json:"description"`
	Delta       float64   `json:"delta,omitempty"`
	OccurredAt  time.Time `json:"occurredAt"`
}

// ActivityFeed is the full response returned by GET /api/activity-feed.
type ActivityFeed struct {
	Activities    []Activity `json:"activities"`
	TotalMissions int        `json:"totalMissions"`
	TotalItems    int        `json:"totalItems"`
	ActiveAlerts  int        `json:"activeAlerts"`
	TotalBudget   float64    `json:"totalBudget"`
	GeneratedAt   time.Time  `json:"generatedAt"`
}
