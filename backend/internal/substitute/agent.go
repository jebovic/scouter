package substitute

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Substitutefinder is the narrow interface consumed by Handler so tests can
// inject a fake without depending on the real LLM.
type SubstituteAgent interface {
	FindSubstitutes(ctx context.Context, name, category string, budget float64) ([]Substitute, error)
}

// Agent uses the LLM to generate product substitute suggestions for the French market.
// It follows the same tool-use pattern as reviews.Agent and pricecomp.Agent.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new substitute Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// FindSubstitutes calls the LLM to produce a list of substitute products
// for the given product name, category, and budget. Results are focused on
// French market availability.
func (a *Agent) FindSubstitutes(ctx context.Context, name, category string, budget float64) ([]Substitute, error) {
	req := a.buildRequest(name, category, budget)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "substitutes",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("substitutes llm call: %w", err)
	}

	subs, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh context.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "substitutes_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "suggest_substitutes")
		if err != nil {
			return nil, fmt.Errorf("substitutes json fallback: %w", err)
		}
		subs, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse substitutes response: %w", err)
		}
	}

	return subs, nil
}

func (a *Agent) buildRequest(name, category string, budget float64) llm.CompletionRequest {
	prompt := fmt.Sprintf(`You are a French market product expert helping shoppers find alternatives.

Product: %s
Category: %s
Budget: %.2f EUR

Suggest 2-3 alternative products that are genuinely available in France (sold at major French retailers like Fnac, Darty, Amazon France, Cdiscount, Leclerc, Boulanger, or direct brand websites). Use the suggest_substitutes tool to return structured results.

Guidelines:
- Each substitute should be a real, purchasable product in France right now
- Prices should be realistic and in EUR
- Advantage must be one of: cheaper, better_rated, eco_friendly, local
- "local" means a French or European brand
- "cheaper" means meaningfully less expensive than the original budget
- "better_rated" means stronger consumer reviews / ratings
- "eco_friendly" means certified sustainable or refurbished
- Reason should be 1-2 sentences explaining why this is a good alternative
- URL is optional but should be a plausible search URL when provided`, name, category, budget)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{substituteTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts substitutes from the first suggest_substitutes tool call.
func parseToolResponse(resp llm.CompletionResponse) ([]Substitute, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "suggest_substitutes" {
			continue
		}
		return unmarshalSubstitutes(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call suggest_substitutes")
}

func unmarshalSubstitutes(input map[string]any) ([]Substitute, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Substitutes []Substitute `json:"substitutes"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal substitutes payload: %w", err)
	}

	if payload.Substitutes == nil {
		payload.Substitutes = []Substitute{}
	}
	return payload.Substitutes, nil
}

// substituteTool returns the tool schema for structured substitute suggestions.
func substituteTool() llm.Tool {
	return llm.Tool{
		Name:        "suggest_substitutes",
		Description: "Submit a list of alternative product suggestions available in the French market. Call this once with all suggestions.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"substitutes"},
			"properties": map[string]any{
				"substitutes": map[string]any{
					"type":        "array",
					"description": "Array of 2-3 alternative products available in France",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"name", "brand", "price", "currency", "retailer", "reason", "advantage"},
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Product name",
							},
							"brand": map[string]any{
								"type":        "string",
								"description": "Brand name",
							},
							"price": map[string]any{
								"type":        "number",
								"description": "Price in the specified currency",
								"minimum":     0,
							},
							"currency": map[string]any{
								"type":        "string",
								"description": "ISO 4217 currency code, typically EUR",
							},
							"retailer": map[string]any{
								"type":        "string",
								"description": "French retailer name (e.g. Fnac, Darty, Amazon France, Cdiscount)",
							},
							"reason": map[string]any{
								"type":        "string",
								"description": "Why this is a good substitute (1-2 sentences)",
							},
							"advantage": map[string]any{
								"type":        "string",
								"description": "Primary advantage: cheaper, better_rated, eco_friendly, or local",
								"enum":        []string{"cheaper", "better_rated", "eco_friendly", "local"},
							},
							"url": map[string]any{
								"type":        "string",
								"description": "Optional search/product URL",
							},
						},
					},
				},
			},
		},
	}
}
