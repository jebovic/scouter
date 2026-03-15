package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Reviewer is the narrow interface consumed by Handler so tests can inject
// a fake without depending on the real LLM.
type Reviewer interface {
	Summarise(ctx context.Context, optionName string) (*ReviewSummary, error)
}

// Agent uses the LLM to generate a simulated review summary for a product.
// It follows the same tool-use pattern as research.Agent and pricing.Agent.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new ReviewAgent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Summarise calls the LLM to produce a ReviewSummary for optionName.
// It uses tool_use (submit_review_summary) and falls back to JSON-only mode
// when the model does not emit a tool call.
func (a *Agent) Summarise(ctx context.Context, optionName string) (*ReviewSummary, error) {
	req := a.buildRequest(optionName)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "reviews",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("reviews llm call: %w", err)
	}

	summary, err := parseToolResponse(resp, optionName)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh context.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "reviews_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "submit_review_summary")
		if err != nil {
			return nil, fmt.Errorf("reviews json fallback: %w", err)
		}
		summary, err = parseToolResponse(resp, optionName)
		if err != nil {
			return nil, fmt.Errorf("parse reviews response: %w", err)
		}
	}

	return summary, nil
}

func (a *Agent) buildRequest(optionName string) llm.CompletionRequest {
	prompt := fmt.Sprintf(`You are a consumer review analyst with access to aggregated public review data.

Product: %s

Simulate realistic aggregated review data for this product as if sourced from major retail and review platforms (Amazon, BestBuy, Trustpilot, Reddit, etc.). Use the submit_review_summary tool to return structured results.

Guidelines:
- Base sentiment and percentages on your knowledge of the product's real-world reception
- Pros and cons should reflect the most commonly mentioned points across reviews
- TotalReviews should be a plausible number (hundreds to tens of thousands)
- Sentiment is a float in [-1.0, 1.0]: -1 = entirely negative, 0 = neutral, 1 = entirely positive
- PositivePct + NeutralPct + NegativePct must equal 100`, optionName)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{reviewTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts a ReviewSummary from the first submit_review_summary tool call.
func parseToolResponse(resp llm.CompletionResponse, optionName string) (*ReviewSummary, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_review_summary" {
			continue
		}
		return unmarshalSummary(tc.Input, optionName)
	}
	return nil, fmt.Errorf("LLM did not call submit_review_summary")
}

func unmarshalSummary(input map[string]any, _ string) (*ReviewSummary, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Sentiment    float64  `json:"sentiment"`
		PositivePct  int      `json:"positive_pct"`
		NeutralPct   int      `json:"neutral_pct"`
		NegativePct  int      `json:"negative_pct"`
		Pros         []string `json:"pros"`
		Cons         []string `json:"cons"`
		TotalReviews int      `json:"total_reviews"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal review summary payload: %w", err)
	}

	if payload.Pros == nil {
		payload.Pros = []string{}
	}
	if payload.Cons == nil {
		payload.Cons = []string{}
	}

	return &ReviewSummary{
		Sentiment:    payload.Sentiment,
		PositivePct:  payload.PositivePct,
		NeutralPct:   payload.NeutralPct,
		NegativePct:  payload.NegativePct,
		Pros:         payload.Pros,
		Cons:         payload.Cons,
		TotalReviews: payload.TotalReviews,
		Source:       "llm-simulated",
		RefreshedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// reviewTool returns the tool schema for structured review summary submission.
func reviewTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_review_summary",
		Description: "Submit a structured aggregated review summary for the product. Call this once with all data.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"sentiment", "positive_pct", "neutral_pct", "negative_pct", "pros", "cons", "total_reviews"},
			"properties": map[string]any{
				"sentiment": map[string]any{
					"type":        "number",
					"description": "Overall sentiment score from -1.0 (very negative) to 1.0 (very positive)",
					"minimum":     -1.0,
					"maximum":     1.0,
				},
				"positive_pct": map[string]any{
					"type":        "integer",
					"description": "Percentage of positive reviews (0-100)",
					"minimum":     0,
					"maximum":     100,
				},
				"neutral_pct": map[string]any{
					"type":        "integer",
					"description": "Percentage of neutral reviews (0-100)",
					"minimum":     0,
					"maximum":     100,
				},
				"negative_pct": map[string]any{
					"type":        "integer",
					"description": "Percentage of negative reviews (0-100). positive_pct + neutral_pct + negative_pct must equal 100.",
					"minimum":     0,
					"maximum":     100,
				},
				"pros": map[string]any{
					"type":        "array",
					"description": "Most commonly praised aspects (3-6 bullet points)",
					"items":       map[string]any{"type": "string"},
				},
				"cons": map[string]any{
					"type":        "array",
					"description": "Most commonly criticised aspects (2-5 bullet points)",
					"items":       map[string]any{"type": "string"},
				},
				"total_reviews": map[string]any{
					"type":        "integer",
					"description": "Estimated total number of reviews across all sources",
					"minimum":     0,
				},
			},
		},
	}
}
