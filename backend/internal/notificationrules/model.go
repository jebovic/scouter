package notificationrules

import "time"

// Alert represents a single triggered smart alert for a shopping item.
type Alert struct {
	ItemID      string    `json:"itemId"`
	ItemName    string    `json:"itemName"`
	RuleID      string    `json:"ruleId"`
	Severity    string    `json:"severity"`    // "critical" | "warning"
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggeredAt"`
}

// AlertsResponse is the response envelope for GET /api/missions/{missionId}/smart-alerts.
type AlertsResponse struct {
	MissionID     string  `json:"missionId"`
	Alerts        []Alert `json:"alerts"`
	CriticalCount int     `json:"criticalCount"`
	WarningCount  int     `json:"warningCount"`
}

// shoppingItem is an internal struct for queried shopping items.
type shoppingItem struct {
	id           string
	name         string
	price        float64
	targetPrice  *float64
	status       string
	costCategory string
}

// pricePoint is an internal struct for a price history entry.
type pricePoint struct {
	price float64
}
