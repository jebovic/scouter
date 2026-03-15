package cashback

import (
	"time"

	"github.com/google/uuid"
)

// Entry represents a single cashback record logged by a user.
type Entry struct {
	ID          uuid.UUID  `json:"id"`
	UserRef     string     `json:"userRef"`
	ItemName    string     `json:"itemName"`
	Merchant    string     `json:"merchant"`
	Amount      float64    `json:"amount"`
	PercentRate float64    `json:"percentRate"`
	Status      string     `json:"status"`
	Notes       string     `json:"notes,omitempty"`
	ClaimedAt   *time.Time `json:"claimedAt,omitempty"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// CreateRequest is the payload for POST /api/cashback.
type CreateRequest struct {
	ItemName    string  `json:"itemName"`
	Merchant    string  `json:"merchant"`
	Amount      float64 `json:"amount"`
	PercentRate float64 `json:"percentRate"`
	Status      string  `json:"status"`
	Notes       string  `json:"notes"`
}

// UpdateRequest is the payload for PUT /api/cashback/{id}.
type UpdateRequest struct {
	Status string     `json:"status"`
	PaidAt *time.Time `json:"paidAt,omitempty"`
	Notes  string     `json:"notes"`
}

// Summary aggregates cashback totals for a user.
type Summary struct {
	TotalPending   float64 `json:"totalPending"`
	TotalConfirmed float64 `json:"totalConfirmed"`
	TotalPaid      float64 `json:"totalPaid"`
	EntryCount     int     `json:"entryCount"`
}

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusPaid      = "paid"
)
