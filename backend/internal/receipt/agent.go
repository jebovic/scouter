package receipt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// AgentInterface is the contract that the Handler depends on.
// Tests inject a fake; production uses Agent.
type AgentInterface interface {
	Analyze(ctx context.Context, rawText string) (*AnalysisResult, error)
}

// Agent uses LLM tool-use to extract structured line items from raw receipt text.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new receipt Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Analyze calls the LLM with the raw receipt text and returns structured items.
func (a *Agent) Analyze(ctx context.Context, rawText string) (*AnalysisResult, error) {
	req := a.buildRequest(rawText)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "receipt_analyze",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("receipt llm call: %w", err)
	}

	result, err := parseToolResponse(resp)
	if err != nil {
		// Fallback: retry in JSON mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "receipt_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "analyze_receipt")
		if err != nil {
			return nil, fmt.Errorf("receipt json fallback: %w", err)
		}
		result, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse receipt response: %w", err)
		}
	}

	return result, nil
}

func (a *Agent) buildRequest(rawText string) llm.CompletionRequest {
	prompt := fmt.Sprintf(
		"Vous êtes un assistant d'analyse de ticket de caisse. Analysez le texte brut fourni et extrayez les articles achetés avec leur quantité, prix unitaire et total. Répondez uniquement avec l'outil analyze_receipt. Si le texte ne ressemble pas à un ticket de caisse, retournez items=[] et total=0.\n\nTexte brut du ticket :\n%s",
		rawText,
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{receiptTool()},
		MaxTokens: 2048,
	}
}

// parseToolResponse extracts the AnalysisResult from the analyze_receipt tool call.
func parseToolResponse(resp llm.CompletionResponse) (*AnalysisResult, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "analyze_receipt" {
			continue
		}
		return unmarshalResult(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call analyze_receipt")
}

func unmarshalResult(input map[string]any) (*AnalysisResult, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Items []struct {
			Name      string  `json:"name"`
			Quantity  float64 `json:"quantity"`
			UnitPrice float64 `json:"unitPrice"`
			Total     float64 `json:"total"`
		} `json:"items"`
		Total    float64 `json:"total"`
		Currency string  `json:"currency"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal receipt payload: %w", err)
	}

	items := make([]ReceiptItem, 0, len(payload.Items))
	for _, p := range payload.Items {
		items = append(items, ReceiptItem{
			Name:      p.Name,
			Quantity:  p.Quantity,
			UnitPrice: p.UnitPrice,
			Total:     p.Total,
		})
	}

	currency := payload.Currency
	if currency == "" {
		currency = "EUR"
	}

	return &AnalysisResult{
		Items:    items,
		Total:    payload.Total,
		Currency: currency,
	}, nil
}

// receiptTool returns the LLM tool schema for receipt analysis.
func receiptTool() llm.Tool {
	return llm.Tool{
		Name:        "analyze_receipt",
		Description: "Extrait les articles achetés d'un ticket de caisse sous forme structurée.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"items", "total", "currency"},
			"properties": map[string]any{
				"items": map[string]any{
					"type":        "array",
					"description": "Liste des articles du ticket",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"name", "quantity", "unitPrice", "total"},
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Nom du produit",
							},
							"quantity": map[string]any{
								"type":        "number",
								"description": "Quantité achetée",
							},
							"unitPrice": map[string]any{
								"type":        "number",
								"description": "Prix unitaire",
							},
							"total": map[string]any{
								"type":        "number",
								"description": "Total ligne (quantité × prix unitaire)",
							},
						},
					},
				},
				"total": map[string]any{
					"type":        "number",
					"description": "Total général du ticket",
				},
				"currency": map[string]any{
					"type":        "string",
					"description": "Code devise ISO 4217 (EUR, USD, GBP…)",
				},
			},
		},
	}
}
