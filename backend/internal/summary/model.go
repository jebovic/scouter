// Package summary implements the AI Shopping Summary Report (Phase 43).
// It generates a concise French-language markdown report via the LLM,
// summarising research findings, top recommendations, and purchase timing advice.
// The report is cached in memory (sync.RWMutex + TTL) — no DB migration needed.
package summary

import (
	"time"

	"github.com/google/uuid"
)

// ShoppingSummary is the generated markdown report for a mission.
type ShoppingSummary struct {
	ID          uuid.UUID `json:"id"`
	MissionID   uuid.UUID `json:"missionId"`
	Content     string    `json:"content"`     // Markdown text
	GeneratedAt time.Time `json:"generatedAt"`
}
