package wishlist

import "time"

type WishListItem struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	URL                *string    `json:"url,omitempty"`
	TargetPrice        *float64   `json:"targetPrice,omitempty"`
	Currency           string     `json:"currency"`
	Notes              *string    `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	LastCheckedAt      *time.Time `json:"lastCheckedAt,omitempty"`
	LastCheckedPrice   *float64   `json:"lastCheckedPrice,omitempty"`
	AlertEnabled       bool       `json:"alertEnabled"`
}

type CreateWishListItemRequest struct {
	Name        string   `json:"name"`
	URL         *string  `json:"url,omitempty"`
	TargetPrice *float64 `json:"targetPrice,omitempty"`
	Currency    string   `json:"currency"`
	Notes       *string  `json:"notes,omitempty"`
}

// UpdateWishListItemRequest is the payload for PATCH /{id}/alert.
type UpdateWishListItemRequest struct {
	AlertEnabled *bool `json:"alertEnabled"`
}
