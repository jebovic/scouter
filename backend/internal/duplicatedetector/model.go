package duplicatedetector

import "time"

// DuplicateItem is a compact view of a shopping item inside a duplicate group.
type DuplicateItem struct {
	ItemID   string  `json:"itemId"`
	ItemName string  `json:"itemName"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
	Merchant string  `json:"merchant"`
}

// DuplicateGroup represents a set of items that are likely duplicates of each other.
type DuplicateGroup struct {
	Items       []DuplicateItem `json:"items"`
	Similarity  float64         `json:"similarity"` // 0-1
	Reason      string          `json:"reason"`     // French explanation
	Recommended string          `json:"recommended"` // itemId to keep
}

// DuplicateReport is the full result returned by the detector for a mission.
type DuplicateReport struct {
	MissionID   string           `json:"missionId"`
	Groups      []DuplicateGroup `json:"groups"`
	TotalDupes  int              `json:"totalDupes"`
	GeneratedAt time.Time        `json:"generatedAt"`
}
