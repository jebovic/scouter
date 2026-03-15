package negotiation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Agent uses the LLM to generate a negotiation script for a purchase option.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new NegotiationAgent.
// Panics if provider is nil.
func NewAgent(provider llm.Provider) *Agent {
	if provider == nil {
		panic("negotiation.NewAgent: provider must not be nil")
	}
	return &Agent{provider: provider}
}

// Run validates the input, calls the LLM, and returns a NegotiationScript.
func (a *Agent) Run(ctx context.Context, input NegotiationInput) (NegotiationScript, error) {
	if err := validateInput(input); err != nil {
		return NegotiationScript{}, fmt.Errorf("invalid input: %w", err)
	}

	req := a.buildRequest(input)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "negotiation",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return NegotiationScript{}, fmt.Errorf("negotiation llm call: %w", err)
	}

	script, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh timeout.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "negotiation_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "generate_negotiation_script")
		if err != nil {
			return NegotiationScript{}, fmt.Errorf("negotiation json fallback: %w", err)
		}
		script, err = parseToolResponse(resp)
		if err != nil {
			return NegotiationScript{}, fmt.Errorf("parse negotiation response: %w", err)
		}
	}

	return script, nil
}

// validateInput checks that the NegotiationInput has valid fields.
func validateInput(input NegotiationInput) error {
	if strings.TrimSpace(input.MissionName) == "" {
		return fmt.Errorf("missionName must not be empty")
	}
	if strings.TrimSpace(input.OptionName) == "" {
		return fmt.Errorf("optionName must not be empty")
	}
	if input.OptionPrice <= 0 {
		return fmt.Errorf("optionPrice must be positive, got %.2f", input.OptionPrice)
	}
	if input.Budget < 0 {
		return fmt.Errorf("budget must not be negative, got %.2f", input.Budget)
	}
	return nil
}

// buildRequest constructs the LLM completion request.
func (a *Agent) buildRequest(input NegotiationInput) llm.CompletionRequest {
	var sb strings.Builder

	// Constraints section
	if len(input.Constraints) > 0 {
		sb.WriteString("\nConstraints:\n")
		for _, c := range input.Constraints {
			qualifier := "preferred"
			if c.Type == "hard" {
				qualifier = "required"
			}
			fmt.Fprintf(&sb, "- %s (%s): %s\n", c.Label, qualifier, c.Value)
		}
	}

	// Attributes section
	if len(input.Attributes) > 0 {
		sb.WriteString("\nKey product attributes:\n")
		for k, v := range input.Attributes {
			fmt.Fprintf(&sb, "- %s: %s\n", k, v)
		}
	}

	// Market prices section
	if len(input.MarketPrices) > 0 {
		sb.WriteString("\nMarket prices observed:\n")
		for _, mp := range input.MarketPrices {
			fmt.Fprintf(&sb, "- %s: %.2f %s\n", mp.Merchant, mp.Price, input.Currency)
		}
	}

	// Historical low section
	if input.HistoricalLow != nil {
		fmt.Fprintf(&sb, "\nHistorical low price: %.2f %s\n", *input.HistoricalLow, input.Currency)
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	prompt := fmt.Sprintf(`You are an expert negotiation coach helping a buyer get the best deal.

Mission: %s
Category: %s
Budget: %.2f %s
Product: %s
Listed Price: %.2f %s
%s
Generate a practical, realistic negotiation script. Use the generate_negotiation_script tool.

Guidelines:
- openingOffer should be the first offer to make (typically 5-15%% below target)
- walkAwayPrice is the maximum the buyer should pay (at or below budget when possible)
- suggestedDiscount is the realistic discount percentage to aim for (0-80%%)
- talkingPoints should be concrete, factual leverage points (competitor prices, specs, timing)
- counterOfferScript should be step-by-step dialogue the buyer can follow verbatim
- confidenceScore reflects how negotiable this purchase is (0=fixed price, 1=highly negotiable)`,
		input.MissionName, input.Category, input.Budget, currency,
		input.OptionName, input.OptionPrice, currency, sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{negotiationTool()},
		MaxTokens: 2048,
	}
}

// parseToolResponse extracts the NegotiationScript from the first generate_negotiation_script tool call.
func parseToolResponse(resp llm.CompletionResponse) (NegotiationScript, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "generate_negotiation_script" {
			continue
		}
		return unmarshalScript(tc.Input)
	}
	return NegotiationScript{}, fmt.Errorf("LLM did not call generate_negotiation_script")
}

func unmarshalScript(input map[string]any) (NegotiationScript, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return NegotiationScript{}, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var script NegotiationScript
	if err := json.Unmarshal(raw, &script); err != nil {
		return NegotiationScript{}, fmt.Errorf("unmarshal negotiation script: %w", err)
	}

	// Ensure slices are non-nil for consistent JSON output
	if script.TalkingPoints == nil {
		script.TalkingPoints = []string{}
	}
	if script.CounterOfferScript == nil {
		script.CounterOfferScript = []string{}
	}

	return script, nil
}

// negotiationTool returns the tool schema for structured negotiation script generation.
func negotiationTool() llm.Tool {
	return llm.Tool{
		Name:        "generate_negotiation_script",
		Description: "Generate a structured negotiation script to help the buyer get the best deal. Call this once with all negotiation guidance.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"openingOffer", "walkAwayPrice", "suggestedDiscount", "talkingPoints", "counterOfferScript", "confidenceScore"},
			"properties": map[string]any{
				"openingOffer": map[string]any{
					"type":        "number",
					"description": "The first price to offer the seller (in the mission currency)",
				},
				"walkAwayPrice": map[string]any{
					"type":        "number",
					"description": "The maximum price the buyer should accept before walking away",
				},
				"suggestedDiscount": map[string]any{
					"type":        "number",
					"description": "The realistic discount percentage to aim for (0 to 80)",
					"minimum":     0,
					"maximum":     80,
				},
				"talkingPoints": map[string]any{
					"type":        "array",
					"description": "Concrete, factual leverage points the buyer can raise during negotiation",
					"items":       map[string]any{"type": "string"},
					"minItems":    1,
				},
				"counterOfferScript": map[string]any{
					"type":        "array",
					"description": "Step-by-step dialogue the buyer should follow during the negotiation",
					"items":       map[string]any{"type": "string"},
					"minItems":    1,
				},
				"confidenceScore": map[string]any{
					"type":        "number",
					"description": "How negotiable this purchase is, from 0.0 (fixed price) to 1.0 (highly negotiable)",
					"minimum":     0,
					"maximum":     1,
				},
			},
		},
	}
}
