package option

import (
	"time"

	"github.com/google/uuid"
)

// Option is a researched alternative within a mission.
type Option struct {
	ID         uuid.UUID   `json:"id"`
	MissionID  uuid.UUID   `json:"missionId"`
	Name       string      `json:"name"`
	Category   string      `json:"category"`
	Badge      string      `json:"badge"` // recommended | alternative | rejected | watch
	Attributes []Attribute `json:"attributes"`
	PriceRange *PriceRange `json:"priceRange,omitempty"`
	Notes      string      `json:"notes,omitempty"`
	Warnings   []string    `json:"warnings"`
	URL        string      `json:"url,omitempty"`
	CreatedAt  time.Time   `json:"createdAt"`
}

// Attribute is a typed key-value pair describing an aspect of an option.
type Attribute struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Value any     `json:"value"`
	Type  string  `json:"type"` // text | price | score | boolean
	Max   *int    `json:"max,omitempty"`
	Pass  *bool   `json:"pass,omitempty"`
}

// PriceRange captures the min/max/best observed prices for an option.
type PriceRange struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Best float64 `json:"best"`
}

// CreateRequest is the payload for adding an option.
type CreateRequest struct {
	Name       string      `json:"name"                validate:"required,min=1"`
	Category   string      `json:"category"`
	Badge      string      `json:"badge"               validate:"required,oneof=recommended alternative watch rejected"`
	Attributes []Attribute `json:"attributes"`
	PriceRange *PriceRange `json:"priceRange,omitempty"`
	Notes      string      `json:"notes,omitempty"`
	Warnings   []string    `json:"warnings"`
	URL        string      `json:"url,omitempty"`
}

// UpdateRequest is the payload for updating an option.
type UpdateRequest struct {
	Badge      *string     `json:"badge,omitempty"      validate:"omitempty,oneof=recommended alternative watch rejected"`
	Attributes []Attribute `json:"attributes,omitempty"`
	PriceRange *PriceRange `json:"priceRange,omitempty"`
	Notes      *string     `json:"notes,omitempty"`
	Warnings   []string    `json:"warnings,omitempty"`
}
