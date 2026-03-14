// Package pricing implements the PricingAgent, which uses LLM tool-use to find
// the best prices and merchants for researched options. It owns prompt construction,
// tool schema definition, response parsing, and shopping item persistence.
// The llm.Provider is transport-only.
package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/usage"
)

// Agent uses the LLM to find prices and merchants for mission options.
type Agent struct {
	provider     llm.Provider
	shoppingRepo shopping.Repository
	usageSvc     *usage.Service
}

// NewAgent creates a new PricingAgent.
func NewAgent(provider llm.Provider, shoppingRepo shopping.Repository, usageSvc *usage.Service) *Agent {
	return &Agent{provider: provider, shoppingRepo: shoppingRepo, usageSvc: usageSvc}
}

// Run performs LLM-based price research for the given options, persists shopping items, and returns them.
// Clears existing shopping items for the mission before persisting new ones (idempotent re-runs).
func (a *Agent) Run(ctx context.Context, m mission.Mission, options []option.Option) ([]shopping.Item, error) {
	if len(options) == 0 {
		return []shopping.Item{}, nil
	}

	if err := a.shoppingRepo.DeleteByMission(ctx, m.ID); err != nil {
		return nil, fmt.Errorf("clear existing shopping items: %w", err)
	}

	req := a.buildRequest(m, options)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("pricing llm call: %w", err)
	}
	a.usageSvc.Log(ctx, resp.Usage, "pricing", m.ID, resp.WasFallback)

	rawItems, err := parseToolResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("parse pricing response: %w", err)
	}

	results := make([]shopping.Item, 0, len(rawItems))
	for _, item := range rawItems {
		item.MissionID = m.ID
		created, err := a.shoppingRepo.Create(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("persist shopping item %q: %w", item.Name, err)
		}
		results = append(results, *created)
	}

	return results, nil
}

func (a *Agent) buildRequest(m mission.Mission, options []option.Option) llm.CompletionRequest {
	var sb strings.Builder
	for i, o := range options {
		priceHint := ""
		if o.PriceRange != nil {
			priceHint = fmt.Sprintf(" (estimated %.0f–%.0f %s)", o.PriceRange.Min, o.PriceRange.Max, m.Currency)
		}
		fmt.Fprintf(&sb, "%d. %s [%s]%s\n", i+1, o.Name, o.Badge, priceHint)
	}

	prompt := fmt.Sprintf(`You are a price intelligence specialist helping optimize a purchase budget.

Mission: %s
Budget: %.2f %s
Currency: %s

Options to price:
%s
For each option, find the best available prices across merchants. Use the submit_pricing_items tool to return structured results.

Guidelines:
- Include 1-3 merchant options per product where relevant (prioritize fewer vendors to consolidate shipping)
- Assign status badges: "buy" (best deal now), "flash-sale" (limited-time deal), "preorder" (not yet available),
  "defer" (wait for better price), "watch" (monitor for deals), "crisis" (over budget), "recommended" (best overall)
- Calculate TCO where relevant (include shipping, warranties, accessories)
- Flag seasonal sales or price drop patterns in notes
- Group items by cost_category matching the mission's cost categories where possible`,
		m.Name, m.Budget, m.Currency, m.Currency, sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{pricingTool()},
		MaxTokens: 4096,
	}
}

// parseToolResponse extracts shopping items from the first submit_pricing_items tool call.
func parseToolResponse(resp llm.CompletionResponse) ([]shopping.Item, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_pricing_items" {
			continue
		}
		return unmarshalItems(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call submit_pricing_items")
}

func unmarshalItems(input map[string]any) ([]shopping.Item, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Items []struct {
			Name             string   `json:"name"`
			Merchant         string   `json:"merchant"`
			CostCategory     string   `json:"cost_category"`
			Price            float64  `json:"price"`
			OriginalEstimate *float64 `json:"original_estimate"`
			Status           string   `json:"status"`
			Note             string   `json:"note"`
			URL              string   `json:"url"`
		} `json:"items"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal pricing payload: %w", err)
	}

	items := make([]shopping.Item, 0, len(payload.Items))
	for _, p := range payload.Items {
		items = append(items, shopping.Item{
			Name:             p.Name,
			Merchant:         p.Merchant,
			CostCategory:     p.CostCategory,
			Price:            p.Price,
			OriginalEstimate: p.OriginalEstimate,
			Status:           p.Status,
			Note:             p.Note,
			URL:              p.URL,
		})
	}

	return items, nil
}

// pricingTool returns the tool schema for structured shopping item submission.
func pricingTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_pricing_items",
		Description: "Submit structured pricing results. Call this once with all priced shopping items.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"items"},
			"properties": map[string]any{
				"items": map[string]any{
					"type":        "array",
					"description": "List of priced shopping items across all options",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"name", "merchant", "cost_category", "price", "status"},
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Product or item name (can include variant/sku details)",
							},
							"merchant": map[string]any{
								"type":        "string",
								"description": "Merchant or platform name (e.g. Amazon, Fnac, Cdiscount)",
							},
							"cost_category": map[string]any{
								"type":        "string",
								"description": "Cost category grouping (e.g. hardware, transport, accommodation)",
							},
							"price": map[string]any{
								"type":        "number",
								"description": "Current price in the mission currency",
							},
							"original_estimate": map[string]any{
								"type":        "number",
								"description": "Original budget estimate for this item if different from current price",
							},
							"status": map[string]any{
								"type": "string",
								"enum": []string{"buy", "flash-sale", "preorder", "defer", "watch", "crisis"},
							},
							"note": map[string]any{
								"type":        "string",
								"description": "Price context, deal expiry, TCO notes, or seasonal patterns",
							},
							"url": map[string]any{
								"type":        "string",
								"description": "Direct product or listing URL",
							},
						},
					},
				},
			},
		},
	}
}
