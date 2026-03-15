package wishlist

import "time"

type WishListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         *string   `json:"url,omitempty"`
	TargetPrice *float64  `json:"targetPrice,omitempty"`
	Currency    string    `json:"currency"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateWishListItemRequest struct {
	Name        string   `json:"name"`
	URL         *string  `json:"url,omitempty"`
	TargetPrice *float64 `json:"targetPrice,omitempty"`
	Currency    string   `json:"currency"`
	Notes       *string  `json:"notes,omitempty"`
}
