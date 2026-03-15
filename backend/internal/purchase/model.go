package purchase

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PurchaseRecord represents a recorded purchase against a mission.
type PurchaseRecord struct {
	ID           uuid.UUID  `json:"id"`
	MissionID    uuid.UUID  `json:"missionId"`
	OptionID     *uuid.UUID `json:"optionId,omitempty"`
	PurchasedAt  time.Time  `json:"purchasedAt"`
	FinalPrice   float64    `json:"finalPrice"`
	Merchant     string     `json:"merchant"`
	Satisfaction *int       `json:"satisfaction,omitempty"`
	Review       *string    `json:"review,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// CreateRequest is the payload for recording a purchase.
type CreateRequest struct {
	OptionID     *uuid.UUID `json:"optionId"`
	PurchasedAt  string     `json:"purchasedAt"` // YYYY-MM-DD
	FinalPrice   float64    `json:"finalPrice"`
	Merchant     string     `json:"merchant"`
	Satisfaction *int       `json:"satisfaction"`
	Review       *string    `json:"review"`
}

// UpdateRequest allows editing satisfaction/review/merchant/finalPrice after recording.
type UpdateRequest struct {
	Satisfaction *int     `json:"satisfaction"`
	Review       *string  `json:"review"`
	Merchant     *string  `json:"merchant"`
	FinalPrice   *float64 `json:"finalPrice"`
}

// Sentinel errors.
var (
	ErrAlreadyExists = errors.New("purchase record already exists for this mission")
	ErrNotFound      = errors.New("purchase record not found")
)
