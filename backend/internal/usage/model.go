// Package usage tracks per-call LLM token consumption for budget visibility.
package usage

import (
	"time"

	"github.com/google/uuid"
)

// Record represents a single LLM call logged to the llm_usage table.
type Record struct {
	ID           uuid.UUID  `json:"id"`
	Provider     string     `json:"provider"`
	Model        string     `json:"model"`
	Agent        string     `json:"agent"`
	MissionID    *uuid.UUID `json:"mission_id,omitempty"`
	InputTokens  int        `json:"input_tokens"`
	OutputTokens int        `json:"output_tokens"`
	WasFallback  bool       `json:"was_fallback"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Summary aggregates usage over a time window.
type Summary struct {
	Period       string         `json:"period"`
	ByProvider   []ProviderStat `json:"by_provider"`
	TotalInput   int            `json:"total_input_tokens"`
	TotalOutput  int            `json:"total_output_tokens"`
	FallbackCalls int           `json:"fallback_calls"`
}

// ProviderStat holds aggregated token counts for one provider.
type ProviderStat struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	Calls        int    `json:"calls"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}
