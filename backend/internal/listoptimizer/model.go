package listoptimizer

type ItemRef struct {
	ItemID   string  `json:"itemId"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
	Priority int     `json:"priority"` // 1=highest
}

type MerchantGroup struct {
	Merchant   string    `json:"merchant"`
	Items      []ItemRef `json:"items"`
	TotalPrice float64   `json:"totalPrice"`
	TripSaving float64   `json:"tripSaving"` // savings vs buying separately
}

type OptimizedList struct {
	MissionID      string          `json:"missionId"`
	MerchantGroups []MerchantGroup `json:"merchantGroups"`
	TotalItems     int             `json:"totalItems"`
	TotalPrice     float64         `json:"totalPrice"`
	EstimatedTrips int             `json:"estimatedTrips"`
	Summary        string          `json:"summary"`
}
