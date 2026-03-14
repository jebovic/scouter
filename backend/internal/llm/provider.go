// Package llm defines the transport-only LLM provider interface.
// Domain logic (prompt construction, tool schemas, response parsing)
// lives in internal/research (ResearchAgent) and internal/pricing (PricingAgent).
package llm

import (
	"context"
	"fmt"
)

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"` // "user" | "assistant" | "system"
	Content string `json:"content"`
}

// Tool defines a function the LLM can call (tool use / function calling).
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// CompletionRequest is a model-agnostic LLM request.
type CompletionRequest struct {
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// ToolCall represents a single tool invocation returned by the LLM.
type ToolCall struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// Usage holds token consumption for a single LLM call.
type Usage struct {
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Provider     string `json:"provider"` // "ollama" | "anthropic"
	Model        string `json:"model"`
}

// CompletionResponse is a model-agnostic LLM response.
type CompletionResponse struct {
	Content     string     `json:"content"`
	ToolCalls   []ToolCall `json:"tool_calls,omitempty"`
	Usage       Usage      `json:"usage"`
	WasFallback bool       `json:"was_fallback,omitempty"` // set by RoutingProvider when secondary was used
	ModelName   string     `json:"model_name,omitempty"`   // which model actually ran
	Degraded    bool       `json:"degraded,omitempty"`     // true if a non-first-choice model was used
	Attempts    int        `json:"attempts,omitempty"`     // how many models were tried
}

// StatusError is returned when the upstream API responds with a non-2xx HTTP status.
// The RoutingProvider uses this to distinguish HTTP 5xx (infra → fallback)
// from HTTP 4xx (model/auth error → return to caller, no fallback).
type StatusError struct {
	Code int
}

func (e *StatusError) Error() string { return fmt.Sprintf("upstream HTTP %d", e.Code) }

// Provider is the transport-only LLM interface.
// Implementations must not contain domain logic.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

// EmbedProvider is the interface for embedding providers (Phase 11 — pgvector).
// Do not implement yet; this stub prevents provider.go changes in Phase 11.
type EmbedProvider interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}
