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
// French market availability and include pros/cons + savings percent (Phase 79).
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
	prompt := fmt.Sprintf(
		`Vous êtes un expert en consommation française. Pour '%s' à %.2f EUR (catégorie: %s), suggérez 3-5 alternatives moins chères disponibles en France, de qualité comparable. Donnez des alternatives réelles et disponibles sur le marché français actuel.

Utilisez l'outil suggest_substitutes pour retourner les résultats structurés. Pour chaque alternative:
- name: nom du produit
- brand: marque
- price: prix en EUR (doit être inférieur à %.2f)
- currency: "EUR"
- retailer: détaillant français (Fnac, Darty, Amazon France, Cdiscount, Leclerc, Boulanger)
- reason: pourquoi c'est une bonne alternative (1-2 phrases)
- advantage: "cheaper", "better_rated", "eco_friendly", ou "local"
- savingsPercent: pourcentage d'économies vs le produit original (0-100)
- pros: liste de 2-3 avantages en français
- cons: liste de 1-2 inconvénients en français
- whyConsider: 1 phrase expliquant pourquoi l'envisager
- url: URL de recherche optionnelle`,
		name, budget, category, budget)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{substituteTool()},
		MaxTokens: 1536,
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

// substituteTool returns the tool schema for structured substitute suggestions (Phase 79 enriched).
func substituteTool() llm.Tool {
	return llm.Tool{
		Name:        "suggest_substitutes",
		Description: "Submit 3-5 cheaper alternative product suggestions available in the French market. Call this once with all suggestions.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"substitutes"},
			"properties": map[string]any{
				"substitutes": map[string]any{
					"type":        "array",
					"description": "Array of 3-5 alternative products cheaper than the original, available in France",
					"minItems":    3,
					"maxItems":    5,
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
								"description": "Why this is a good substitute (1-2 sentences in French)",
							},
							"advantage": map[string]any{
								"type":        "string",
								"description": "Primary advantage: cheaper, better_rated, eco_friendly, or local",
								"enum":        []string{"cheaper", "better_rated", "eco_friendly", "local"},
							},
							"savingsPercent": map[string]any{
								"type":        "number",
								"description": "Percentage saved vs original price (0-100)",
								"minimum":     0,
								"maximum":     100,
							},
							"pros": map[string]any{
								"type":        "array",
								"description": "2-3 advantages in French",
								"maxItems":    3,
								"items":       map[string]any{"type": "string"},
							},
							"cons": map[string]any{
								"type":        "array",
								"description": "1-2 disadvantages in French",
								"maxItems":    2,
								"items":       map[string]any{"type": "string"},
							},
							"whyConsider": map[string]any{
								"type":        "string",
								"description": "One sentence in French explaining why to consider this alternative",
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
