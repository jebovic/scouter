// Package search implements semantic search over options using pgvector cosine distance.
package search

import (
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/option"
)

// Result is a single item returned by semantic search or the similar-options endpoint.
type Result struct {
	ID          uuid.UUID          `json:"id"`
	MissionID   uuid.UUID          `json:"mission_id"`
	MissionName string             `json:"mission_name"`
	MissionSlug string             `json:"mission_slug"`
	Name        string             `json:"name"`
	Category    string             `json:"category"`
	Badge       string             `json:"badge"`
	PriceRange  *option.PriceRange `json:"price_range,omitempty"`
	Notes       string             `json:"notes"`
	Similarity  float64            `json:"similarity"` // 0..1, higher = closer
}
