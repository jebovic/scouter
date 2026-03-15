// Package persona implements the PersonaAgent, which uses LLM tool-use to analyze
// the user's purchase history and generate a spending persona profile. It owns
// prompt construction, tool schema definition, response parsing, and persistence.
package persona

import (
	"errors"
	"time"
)

// ErrNotFound is returned when no persona record exists yet.
var ErrNotFound = errors.New("no persona found")

// Persona holds the generated spending archetype and associated insights.
type Persona struct {
	ID        string    `json:"id"`
	Archetype string    `json:"archetype"`
	Traits    []string  `json:"traits"`
	Tips      []string  `json:"tips"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
}
