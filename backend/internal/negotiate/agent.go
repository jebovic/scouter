package negotiate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// AgentInterface is the narrow interface consumed by Handler so tests can inject a fake.
type AgentInterface interface {
	Negotiate(ctx context.Context, itemID, itemName, merchant string, currentPrice float64) (*NegotiationAdvice, error)
}

// Agent uses the LLM to generate negotiation tips for a shopping item.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new negotiate Agent backed by the given LLM provider.
// Panics if provider is nil.
func NewAgent(provider llm.Provider) *Agent {
	if provider == nil {
		panic("negotiate.NewAgent: provider must not be nil")
	}
	return &Agent{provider: provider}
}

// Negotiate calls the LLM with a French-market focused prompt and returns
// NegotiationAdvice for the given shopping item. The target price is
// computed as currentPrice * 0.87 (13% discount target).
func (a *Agent) Negotiate(ctx context.Context, itemID, itemName, merchant string, currentPrice float64) (*NegotiationAdvice, error) {
	targetPrice := currentPrice * 0.87

	req := a.buildRequest(itemName, merchant, currentPrice, targetPrice)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "negotiate-price",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("negotiate llm call: %w", err)
	}

	advice, err := parseToolResponse(resp, itemID, itemName, currentPrice, targetPrice)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapText,
			Label:        "negotiate-price-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "negotiate_price")
		if err != nil {
			return nil, fmt.Errorf("negotiate json fallback: %w", err)
		}
		advice, err = parseToolResponse(resp, itemID, itemName, currentPrice, targetPrice)
		if err != nil {
			return nil, fmt.Errorf("parse negotiate response: %w", err)
		}
	}

	return advice, nil
}

// toolPayload is the shape of data returned by the negotiate_price tool.
type toolPayload struct {
	Tips   []NegotiationTip `json:"tips"`
	Script string           `json:"script"`
}

func (a *Agent) buildRequest(itemName, merchant string, currentPrice, targetPrice float64) llm.CompletionRequest {
	prompt := fmt.Sprintf(
		`Vous êtes un expert en négociation et consommation française. Pour le produit '%s' vendu %.2f€ chez '%s', donnez 4-6 conseils de négociation pratiques et un script prêt à l'emploi en français. Adaptez les conseils au contexte français (grande surface, e-commerce, etc.). L'objectif de prix est %.2f€ (réduction de 13%%).`,
		itemName, currentPrice, merchant, targetPrice,
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{negotiateTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts NegotiationAdvice from the negotiate_price tool call.
func parseToolResponse(resp llm.CompletionResponse, itemID, itemName string, currentPrice, targetPrice float64) (*NegotiationAdvice, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "negotiate_price" {
			continue
		}
		payload, err := unmarshalPayload(tc.Input)
		if err != nil {
			return nil, err
		}
		// Ensure non-nil tips slice.
		if payload.Tips == nil {
			payload.Tips = []NegotiationTip{}
		}
		return &NegotiationAdvice{
			ItemID:       itemID,
			ItemName:     itemName,
			CurrentPrice: currentPrice,
			TargetPrice:  targetPrice,
			Tips:         payload.Tips,
			Script:       payload.Script,
			CachedAt:     time.Now().Unix(),
		}, nil
	}
	return nil, fmt.Errorf("LLM did not call negotiate_price")
}

func unmarshalPayload(input map[string]any) (toolPayload, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return toolPayload{}, fmt.Errorf("re-marshal tool input: %w", err)
	}
	var p toolPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return toolPayload{}, fmt.Errorf("unmarshal negotiate payload: %w", err)
	}
	return p, nil
}

// negotiateTool returns the LLM tool schema for the negotiate_price function.
func negotiateTool() llm.Tool {
	return llm.Tool{
		Name:        "negotiate_price",
		Description: "Génère 4 à 6 conseils de négociation pratiques et un script prêt à l'emploi en français pour aider un acheteur à obtenir un meilleur prix.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"tips", "script"},
			"properties": map[string]any{
				"tips": map[string]any{
					"type":        "array",
					"description": "Liste de 4 à 6 conseils de négociation pratiques",
					"minItems":    4,
					"maxItems":    6,
					"items": map[string]any{
						"type":     "object",
						"required": []string{"title", "description", "difficulty"},
						"properties": map[string]any{
							"title": map[string]any{
								"type":        "string",
								"description": "Titre court du conseil (en français)",
							},
							"description": map[string]any{
								"type":        "string",
								"description": "Description du conseil en 1 à 2 phrases (en français)",
							},
							"difficulty": map[string]any{
								"type":        "string",
								"description": "Difficulté d'application du conseil",
								"enum":        []string{"easy", "medium", "hard"},
							},
						},
					},
				},
				"script": map[string]any{
					"type":        "string",
					"description": "Phrase ou script prêt à l'emploi en français à utiliser face au vendeur",
				},
			},
		},
	}
}
