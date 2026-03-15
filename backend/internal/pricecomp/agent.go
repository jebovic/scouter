package pricecomp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Agent uses the LLM to generate simulated price comparison data across French
// retailers for a product. It follows the same tool-use pattern as reviews.Agent.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new PriceCompAgent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Compare calls the LLM to produce a PriceComparison for productName.
// It uses tool_use (submit_price_comparison) and falls back to JSON-only mode
// when the model does not emit a tool call.
func (a *Agent) Compare(ctx context.Context, productName, optionID string) (*PriceComparison, error) {
	req := a.buildRequest(productName)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "pricecomp",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("pricecomp llm call: %w", err)
	}

	comparison, err := parseToolResponse(resp, productName, optionID)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh context.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "pricecomp_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "submit_price_comparison")
		if err != nil {
			return nil, fmt.Errorf("pricecomp json fallback: %w", err)
		}
		comparison, err = parseToolResponse(resp, productName, optionID)
		if err != nil {
			return nil, fmt.Errorf("parse pricecomp response: %w", err)
		}
	}

	return comparison, nil
}

func (a *Agent) buildRequest(productName string) llm.CompletionRequest {
	prompt := fmt.Sprintf(`You are a French retail price analyst with access to current pricing data.

Product: %s

Simulate realistic price comparison data for this product across major French retailers (Fnac, Cdiscount, Amazon FR, Darty, Boulanger, La Redoute). Use the submit_price_comparison tool to return structured results.

Guidelines:
- Prices must be in EUR and reflect realistic French market pricing for the product
- All 6 retailers must be included even if some are unavailable
- Available retailers should have a valid product URL (can be a plausible placeholder)
- Unavailable retailers should have available=false and may have an empty URL
- Prices should vary realistically across retailers (5-20%% spread is normal)
- Sort offers by price ascending, with available=true offers before available=false
- Set lowestIdx to the index of the cheapest available offer (0-based)`, productName)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{priceCompTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts a PriceComparison from the first submit_price_comparison tool call.
func parseToolResponse(resp llm.CompletionResponse, productName, optionID string) (*PriceComparison, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_price_comparison" {
			continue
		}
		return unmarshalComparison(tc.Input, productName, optionID)
	}
	return nil, fmt.Errorf("LLM did not call submit_price_comparison")
}

func unmarshalComparison(input map[string]any, productName, optionID string) (*PriceComparison, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Offers []struct {
			Retailer  string  `json:"retailer"`
			Price     float64 `json:"price"`
			Currency  string  `json:"currency"`
			Available bool    `json:"available"`
			URL       string  `json:"url"`
		} `json:"offers"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal pricecomp payload: %w", err)
	}

	if payload.Offers == nil {
		payload.Offers = []struct {
			Retailer  string  `json:"retailer"`
			Price     float64 `json:"price"`
			Currency  string  `json:"currency"`
			Available bool    `json:"available"`
			URL       string  `json:"url"`
		}{}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	offers := make([]RetailerOffer, 0, len(payload.Offers))
	for _, o := range payload.Offers {
		currency := o.Currency
		if currency == "" {
			currency = "EUR"
		}
		offers = append(offers, RetailerOffer{
			Retailer:  o.Retailer,
			Price:     o.Price,
			Currency:  currency,
			Available: o.Available,
			URL:       o.URL,
			UpdatedAt: now,
		})
	}

	// Sort: available first, then by price ascending within each group.
	sort.SliceStable(offers, func(i, j int) bool {
		if offers[i].Available != offers[j].Available {
			return offers[i].Available // available before unavailable
		}
		return offers[i].Price < offers[j].Price
	})

	// Find index of the cheapest available offer.
	lowestIdx := 0
	for idx, o := range offers {
		if o.Available {
			lowestIdx = idx
			break
		}
	}

	return &PriceComparison{
		OptionID:  optionID,
		Product:   productName,
		Offers:    offers,
		LowestIdx: lowestIdx,
		FetchedAt: now,
	}, nil
}

// priceCompTool returns the tool schema for structured price comparison submission.
func priceCompTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_price_comparison",
		Description: "Submit structured price comparison data for the product across French retailers. Call this once with all offers.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"offers"},
			"properties": map[string]any{
				"offers": map[string]any{
					"type":        "array",
					"description": "List of retailer offers. Must include Fnac, Cdiscount, Amazon FR, Darty, Boulanger, La Redoute.",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"retailer", "price", "currency", "available"},
						"properties": map[string]any{
							"retailer": map[string]any{
								"type":        "string",
								"description": "Name of the retailer, e.g. 'Fnac', 'Cdiscount', 'Amazon FR', 'Darty', 'Boulanger', 'La Redoute'",
							},
							"price": map[string]any{
								"type":        "number",
								"description": "Current price in EUR",
								"minimum":     0,
							},
							"currency": map[string]any{
								"type":        "string",
								"description": "ISO 4217 currency code, always 'EUR' for French market",
								"enum":        []string{"EUR"},
							},
							"available": map[string]any{
								"type":        "boolean",
								"description": "Whether the product is currently in stock and available to purchase",
							},
							"url": map[string]any{
								"type":        "string",
								"description": "Product page URL at this retailer (empty string if unavailable)",
							},
						},
					},
					"minItems": 1,
				},
			},
		},
	}
}
