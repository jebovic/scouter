package dealexplain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// ExplainAgentInterface is the narrow interface consumed by Handler so tests
// can inject a fake without depending on the real LLM.
type ExplainAgentInterface interface {
	Explain(ctx context.Context, input ExplainInput) (*DealExplanation, error)
}

// Agent uses the LLM to generate a French deal explanation for a shopping item.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new deal-explain Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Explain calls the LLM to produce a French explanation and tags for the deal.
// On LLM failure it returns an error and the handler falls back to deterministic text.
func (a *Agent) Explain(ctx context.Context, input ExplainInput) (*DealExplanation, error) {
	req := a.buildRequest(input)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "deal-explain",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("deal-explain llm call: %w", err)
	}

	explanation, tags, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry with JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapText,
			Label:        "deal-explain-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "suggest_deal_tags")
		if err != nil {
			return nil, fmt.Errorf("deal-explain json fallback: %w", err)
		}
		explanation, tags, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse deal-explain response: %w", err)
		}
	}

	return &DealExplanation{
		ItemID:      input.ItemID,
		Explanation: explanation,
		Tags:        tags,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func (a *Agent) buildRequest(input ExplainInput) llm.CompletionRequest {
	trendFR := trendLabel(input.Trend)
	prompt := fmt.Sprintf(`Tu es un expert en achats pour le marché français. Analyse cette affaire et explique pourquoi elle est bonne ou pas.

Produit: %s
Marchand: %s
Prix actuel: %.2f EUR
Score d'affaire: %d/100 (0=mauvaise affaire, 100=excellente affaire)
Tendance des prix: %s

Utilise l'outil suggest_deal_tags pour retourner une explication de 2-3 phrases en français et 2-4 tags courts en français.`,
		input.Name, input.Merchant, input.Price, input.Score, trendFR)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{dealTagTool()},
		MaxTokens: 512,
	}
}

// parseToolResponse extracts explanation and tags from the suggest_deal_tags tool call.
func parseToolResponse(resp llm.CompletionResponse) (string, []string, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "suggest_deal_tags" {
			continue
		}
		return unmarshalDealTags(tc.Input)
	}
	return "", nil, fmt.Errorf("LLM did not call suggest_deal_tags")
}

func unmarshalDealTags(input map[string]any) (string, []string, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return "", nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Explanation string   `json:"explanation"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", nil, fmt.Errorf("unmarshal deal tags payload: %w", err)
	}

	if payload.Tags == nil {
		payload.Tags = []string{}
	}
	return payload.Explanation, payload.Tags, nil
}

func dealTagTool() llm.Tool {
	return llm.Tool{
		Name:        "suggest_deal_tags",
		Description: "Return explanation and tags for a deal",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"explanation", "tags"},
			"properties": map[string]any{
				"explanation": map[string]any{
					"type":        "string",
					"description": "2-3 sentence French explanation of why this is a good or bad deal",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "2-4 short French tags describing the deal",
					"items":       map[string]any{"type": "string"},
				},
			},
		},
	}
}

// trendLabel converts trend code to French label for the prompt.
func trendLabel(trend string) string {
	switch trend {
	case "falling":
		return "en baisse"
	case "rising":
		return "en hausse"
	default:
		return "stable"
	}
}
