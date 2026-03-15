package wishlistshare

import "time"

// WishlistItem is a single item in the shareable wishlist card.
type WishlistItem struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
	Merchant string  `json:"merchant"`
}

// WishlistShareCard is the data structure returned for a shareable wishlist card.
type WishlistShareCard struct {
	MissionID   string         `json:"missionId"`
	MissionName string         `json:"missionName"`
	TotalBudget float64        `json:"totalBudget"`
	Currency    string         `json:"currency"`
	Items       []WishlistItem `json:"items"`
	ShareURL    string         `json:"shareUrl"`
	GeneratedAt time.Time      `json:"generatedAt"`
}
