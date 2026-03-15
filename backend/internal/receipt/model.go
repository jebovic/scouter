// Package receipt implements AI-powered receipt text analysis.
// The Agent owns prompt construction, tool schema, and LLM response parsing.
// The Repository persists and retrieves ReceiptAnalysis records.
// The Handler exposes HTTP endpoints under /api/missions/{slug}/receipts.
package receipt

import (
	"time"

	"github.com/google/uuid"
)

// ReceiptItem is a single line item extracted from a receipt.
type ReceiptItem struct {
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Total     float64 `json:"total"`
}

// ReceiptAnalysis is the persisted result of analysing a raw receipt text.
type ReceiptAnalysis struct {
	ID          uuid.UUID     `json:"id"`
	MissionSlug string        `json:"missionSlug"`
	RawText     string        `json:"rawText"`
	Items       []ReceiptItem `json:"items"`
	Total       float64       `json:"total"`
	Currency    string        `json:"currency"`
	AnalyzedAt  time.Time     `json:"analyzedAt"`
}

// AnalysisResult is the structured output produced by the Agent before DB persistence.
type AnalysisResult struct {
	Items    []ReceiptItem `json:"items"`
	Total    float64       `json:"total"`
	Currency string        `json:"currency"`
}
