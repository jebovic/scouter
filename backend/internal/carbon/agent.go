package carbon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// CarbonAgentInterface is the narrow interface consumed by Handler so tests
// can inject a fake without depending on the real LLM.
type CarbonAgentInterface interface {
	Estimate(ctx context.Context, itemName, category string) (*CarbonEstimate, error)
}

// Agent uses the LLM to estimate the carbon footprint of a product.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Estimate calls the LLM to produce a CO2 footprint estimation with a French
// label, eco tips, and confidence level.
func (a *Agent) Estimate(ctx context.Context, itemName, category string) (*CarbonEstimate, error) {
	llmReq := a.buildRequest(itemName, category)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "carbon-estimate",
	})

	resp, err := a.provider.Complete(llmCtx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("carbon estimate llm call: %w", err)
	}

	payload, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "carbon-estimate-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, llmReq, "carbon_estimate")
		if err != nil {
			return nil, fmt.Errorf("carbon estimate json fallback: %w", err)
		}
		payload, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse carbon estimate response: %w", err)
		}
	}

	return buildEstimate(itemName, category, payload), nil
}

// toolPayload is the shape of data returned by the carbon_estimate tool.
type toolPayload struct {
	KgCO2      float64  `json:"kg_co2"`
	Label      string   `json:"label"`
	Tips       []string `json:"tips"`
	Confidence string   `json:"confidence"`
}

func (a *Agent) buildRequest(itemName, category string) llm.CompletionRequest {
	catStr := category
	if catStr == "" {
		catStr = "non spécifiée"
	}
	prompt := fmt.Sprintf(
		`Estime l'empreinte carbone totale du produit suivant en tenant compte de son cycle de vie complet (fabrication, transport, utilisation, fin de vie).

Produit : %s
Catégorie : %s

Utilise tes connaissances sur les émissions typiques de ce type de produit pour donner une estimation réaliste en kg CO₂ équivalent.`,
		itemName, catStr,
	)

	return llm.CompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "Tu es un expert en analyse d'impact carbone des produits de consommation. Estime l'empreinte carbone totale (fabrication, transport, utilisation, fin de vie) pour le produit donné.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools:     []llm.Tool{carbonTool()},
		MaxTokens: 512,
	}
}

// parseToolResponse extracts the carbon_estimate tool call payload.
func parseToolResponse(resp llm.CompletionResponse) (toolPayload, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "carbon_estimate" {
			continue
		}
		return unmarshalPayload(tc.Input)
	}
	return toolPayload{}, fmt.Errorf("LLM did not call carbon_estimate")
}

func unmarshalPayload(input map[string]any) (toolPayload, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return toolPayload{}, fmt.Errorf("re-marshal tool input: %w", err)
	}
	var p toolPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return toolPayload{}, fmt.Errorf("unmarshal carbon payload: %w", err)
	}
	return p, nil
}

// buildEstimate assembles a CarbonEstimate from the LLM payload.
func buildEstimate(itemName, category string, p toolPayload) *CarbonEstimate {
	kg := p.KgCO2
	if kg < 0 {
		kg = 0
	}

	label := p.Label
	if label == "" {
		label = labelFromKg(kg)
	}

	confidence := p.Confidence
	if confidence != "high" && confidence != "medium" && confidence != "low" {
		confidence = "medium"
	}

	tips := p.Tips
	if tips == nil {
		tips = []string{}
	}

	return &CarbonEstimate{
		ItemName:   itemName,
		Category:   category,
		KgCO2:      kg,
		Label:      label,
		Tips:       tips,
		Confidence: confidence,
	}
}

// labelFromKg returns a French label for a CO2 value in kg.
func labelFromKg(kg float64) string {
	switch {
	case kg <= 1:
		return "Très faible"
	case kg <= 5:
		return "Faible"
	case kg <= 20:
		return "Moyen"
	case kg <= 100:
		return "Élevé"
	default:
		return "Très élevé"
	}
}

// carbonTool returns the LLM tool schema for the carbon_estimate function.
func carbonTool() llm.Tool {
	return llm.Tool{
		Name:        "carbon_estimate",
		Description: "Estimate the total CO2 carbon footprint of a consumer product over its full lifecycle (manufacturing, transport, usage, end of life).",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"kg_co2", "label", "tips", "confidence"},
			"properties": map[string]any{
				"kg_co2": map[string]any{
					"type":        "number",
					"description": "Estimated CO2 equivalent in kilograms for the full product lifecycle",
					"minimum":     0,
				},
				"label": map[string]any{
					"type":        "string",
					"description": "French impact label: one of 'Très faible', 'Faible', 'Moyen', 'Élevé', 'Très élevé'",
					"enum":        []string{"Très faible", "Faible", "Moyen", "Élevé", "Très élevé"},
				},
				"tips": map[string]any{
					"type":        "array",
					"description": "2 to 3 practical eco tips in French to reduce the product's carbon impact",
					"items": map[string]any{
						"type": "string",
					},
					"minItems": 2,
					"maxItems": 3,
				},
				"confidence": map[string]any{
					"type":        "string",
					"description": "Confidence level of the estimate",
					"enum":        []string{"high", "medium", "low"},
				},
			},
		},
	}
}
