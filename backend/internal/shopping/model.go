package shopping

import (
	"time"

	"github.com/google/uuid"
)

// Item is a concrete thing to buy/book/commit to within a mission.
type Item struct {
	ID               uuid.UUID `json:"id"`
	MissionID        uuid.UUID `json:"missionId"`
	Name             string    `json:"name"`
	Merchant         string    `json:"merchant"`
	CostCategory     string    `json:"costCategory"`
	Price            float64   `json:"price"`
	OriginalEstimate *float64  `json:"originalEstimate,omitempty"`
	TargetPrice      *float64  `json:"targetPrice,omitempty"`
	Status           string    `json:"status"` // buy | flash-sale | preorder | defer | watch | crisis
	Note             string    `json:"note,omitempty"`
	URL              string    `json:"url,omitempty"`
	Pinned           bool      `json:"pinned"`
	CreatedAt        time.Time `json:"createdAt"`
}

// PriceSnapshot records a price observation at a point in time.
type PriceSnapshot struct {
	ID         uuid.UUID `json:"id"`
	ItemID     uuid.UUID `json:"itemId"`
	Price      float64   `json:"price"`
	RecordedAt time.Time `json:"recordedAt"`
	Note       string    `json:"note,omitempty"`
}

// CreateRequest is the payload for adding a shopping item.
type CreateRequest struct {
	Name             string   `json:"name"                          validate:"required,min=1"`
	Merchant         string   `json:"merchant"`
	CostCategory     string   `json:"costCategory"`
	Price            float64  `json:"price"                         validate:"gte=0"`
	OriginalEstimate *float64 `json:"originalEstimate,omitempty"    validate:"omitempty,gte=0"`
	TargetPrice      *float64 `json:"targetPrice,omitempty"         validate:"omitempty,gte=0"`
	Status           string   `json:"status"                        validate:"required,oneof=buy flash-sale preorder defer watch crisis recommended"`
	Note             string   `json:"note,omitempty"`
	URL              string   `json:"url,omitempty"`
}

// UpdateRequest is the payload for updating a shopping item.
type UpdateRequest struct {
	Price       *float64 `json:"price,omitempty"       validate:"omitempty,gte=0"`
	TargetPrice *float64 `json:"targetPrice,omitempty" validate:"omitempty,gte=0"`
	Status      *string  `json:"status,omitempty"      validate:"omitempty,oneof=buy flash-sale preorder defer watch crisis recommended"`
	Note        *string  `json:"note,omitempty"`
	Merchant    *string  `json:"merchant,omitempty"`
}

// PriceSnapshotRequest records a new price snapshot.
type PriceSnapshotRequest struct {
	Price float64 `json:"price" validate:"gte=0"`
	Note  string  `json:"note,omitempty"`
}

