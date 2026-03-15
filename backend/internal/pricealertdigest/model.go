package pricealertdigest

// AlertEntry represents a single alert for an item.
type AlertEntry struct {
	ItemID    string  `json:"itemId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	AlertType string  `json:"alertType"` // "target_hit", "below_avg", "trending_down"
	Message   string  `json:"message"`   // French message
	Savings   float64 `json:"savings"`   // potential savings vs avg or target
}

// PriceAlertDigest is the response returned by GET /api/missions/{missionId}/price-alert-digest.
type PriceAlertDigest struct {
	MissionID   string       `json:"missionId"`
	AlertCount  int          `json:"alertCount"`
	Alerts      []AlertEntry `json:"alerts"`
	Summary     string       `json:"summary"`    // French summary
	GeneratedAt string       `json:"generatedAt"` // RFC3339
}
