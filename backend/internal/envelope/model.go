// Package envelope implements the Envelope Budgeting feature (Phase 35).
// Envelopes represent named monthly budget allocations (e.g. "Groceries €300/mo").
package envelope

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel validation errors.
var (
	ErrNameRequired  = errors.New("name is required")
	ErrInvalidAmount = errors.New("monthlyAmount must be greater than 0")
)

// Envelope is a named monthly budget bucket.
type Envelope struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	MonthlyAmount float64   `json:"monthlyAmount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
}

// EnvelopeSummary aggregates an envelope with computed mission stats.
type EnvelopeSummary struct {
	Envelope     Envelope `json:"envelope"`
	MissionCount int      `json:"missionCount"`
	TotalBudget  float64  `json:"totalBudget"`
	TotalSpent   float64  `json:"totalSpent"`
}

// CreateRequest is the JSON payload for POST /api/envelopes.
type CreateRequest struct {
	Name          string  `json:"name"`
	MonthlyAmount float64 `json:"monthlyAmount"`
	Currency      string  `json:"currency"`
}

// UpdateRequest is the JSON payload for PATCH /api/envelopes/:id.
// All fields are optional — only non-nil fields are applied.
type UpdateRequest struct {
	Name          *string  `json:"name,omitempty"`
	MonthlyAmount *float64 `json:"monthlyAmount,omitempty"`
	Currency      *string  `json:"currency,omitempty"`
}
