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

// ShoppingPersona is the global AI-generated spending archetype with a French label,
// behavioral insights, and personalized tips.
type ShoppingPersona struct {
	Archetype    string         `json:"archetype"`    // French label e.g. "Le Chasseur de Deals"
	Icon         string         `json:"icon"`         // emoji e.g. "🎯"
	Description  string         `json:"description"`  // 2-3 sentences, French
	Strengths    []string       `json:"strengths"`    // 3 French bullet points
	Blindspots   []string       `json:"blindspots"`   // 2 French blind spots
	Tips         []string       `json:"tips"`         // 3 personalized French tips
	ScoreProfile map[string]int `json:"scoreProfile"` // e.g. {"frugalité": 85, "impulsivité": 20}
	CachedAt     int64          `json:"cachedAt"`     // unix timestamp
}
