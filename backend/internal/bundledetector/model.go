// Package bundledetector detects synergistic item pairs within a mission's
// shopping list and suggests bundle savings opportunities. Detection is purely
// deterministic (FNV hash + category synergy rules) — no LLM calls.
package bundledetector

// BundleItem is a single item participating in a bundle deal.
type BundleItem struct {
	ItemID   string  `json:"itemId"`
	ItemName string  `json:"itemName"`
	Price    float64 `json:"price"`
}

// BundleDeal represents a detected synergistic bundle between two or more items.
type BundleDeal struct {
	Items         []BundleItem `json:"items"`
	BundleName    string       `json:"bundleName"`
	TotalPrice    float64      `json:"totalPrice"`
	EstimatedSave float64      `json:"estimatedSave"` // absolute € amount
	SavePercent   float64      `json:"savePercent"`   // % discount if bought together
	Tip           string       `json:"tip"`
	Category      string       `json:"category"`
	Confidence    string       `json:"confidence"` // "high" | "medium" | "low"
}

// BundleResponse is the JSON envelope for GET /api/missions/{id}/bundle-deals.
type BundleResponse struct {
	MissionID string       `json:"missionId"`
	Bundles   []BundleDeal `json:"bundles"`
	TotalSave float64      `json:"totalPotentialSave"`
	Message   string       `json:"message"`
}

// rawItem holds the DB row for a shopping item.
type rawItem struct {
	id    string
	name  string
	price float64
}
