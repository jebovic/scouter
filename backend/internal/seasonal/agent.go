package seasonal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// SeasonalAgentInterface is the narrow interface consumed by Handler so tests
// can inject a fake without depending on the real LLM.
type SeasonalAgentInterface interface {
	Predict(ctx context.Context, itemName, category string) (*Prediction, error)
}

// Agent uses the LLM to predict the best time to buy a product on the French market.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Predict calls the LLM to produce a seasonal discount prediction with French
// market context, best months, expected discounts, and relevant shopping events.
func (a *Agent) Predict(ctx context.Context, itemName, category string) (*Prediction, error) {
	llmReq := a.buildRequest(itemName, category)

	llmCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "seasonal-prediction",
	})

	resp, err := a.provider.Complete(llmCtx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("seasonal prediction llm call: %w", err)
	}

	payload, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "seasonal-prediction-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, llmReq, "seasonal_prediction")
		if err != nil {
			return nil, fmt.Errorf("seasonal prediction json fallback: %w", err)
		}
		payload, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse seasonal prediction response: %w", err)
		}
	}

	return buildPrediction(itemName, category, payload), nil
}

// toolPayload is the shape of data returned by the seasonal_prediction tool.
type toolPayload struct {
	BestMonths     []string `json:"best_months"`
	AvgDiscountPct float64  `json:"avg_discount_pct"`
	Events         []string `json:"events"`
	NextEventDays  int      `json:"next_event_days"`
	Rationale      string   `json:"rationale"`
	Confidence     string   `json:"confidence"`
}

func (a *Agent) buildRequest(itemName, category string) llm.CompletionRequest {
	catStr := category
	if catStr == "" {
		catStr = "non spécifiée"
	}
	prompt := fmt.Sprintf(
		`Analyse le meilleur moment pour acheter le produit suivant sur le marché français.
Tiens compte des événements commerciaux majeurs : Soldes d'hiver (janvier), Soldes d'été (juillet),
Black Friday (novembre), Cyber Monday (novembre), Amazon Prime Day (juillet), Rentrée scolaire (août-septembre),
Fête des Mères (mai), Noël (décembre), Saint-Valentin (février).

Produit : %s
Catégorie : %s

Indique les meilleurs mois pour acheter, la remise moyenne attendue, les événements pertinents,
le nombre de jours jusqu'à la prochaine opportunité d'achat, et une explication en français.`,
		itemName, catStr,
	)

	return llm.CompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "Tu es un expert en analyse des tendances de prix et des promotions sur le marché français. Aide les consommateurs à trouver le meilleur moment pour acheter leurs produits.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools:     []llm.Tool{seasonalTool()},
		MaxTokens: 512,
	}
}

// parseToolResponse extracts the seasonal_prediction tool call payload.
func parseToolResponse(resp llm.CompletionResponse) (toolPayload, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "seasonal_prediction" {
			continue
		}
		return unmarshalPayload(tc.Input)
	}
	return toolPayload{}, fmt.Errorf("LLM did not call seasonal_prediction")
}

func unmarshalPayload(input map[string]any) (toolPayload, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return toolPayload{}, fmt.Errorf("re-marshal tool input: %w", err)
	}
	var p toolPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return toolPayload{}, fmt.Errorf("unmarshal seasonal payload: %w", err)
	}
	return p, nil
}

// buildPrediction assembles a Prediction from the LLM payload.
func buildPrediction(itemName, category string, p toolPayload) *Prediction {
	discount := p.AvgDiscountPct
	if discount < 0 {
		discount = 0
	}
	if discount > 100 {
		discount = 100
	}

	nextDays := p.NextEventDays
	if nextDays < 0 {
		nextDays = 0
	}

	confidence := p.Confidence
	if confidence != "high" && confidence != "medium" && confidence != "low" {
		confidence = "medium"
	}

	bestMonths := p.BestMonths
	if bestMonths == nil {
		bestMonths = []string{}
	}

	events := p.Events
	if events == nil {
		events = []string{}
	}

	rationale := p.Rationale
	if rationale == "" {
		rationale = "Aucune information disponible."
	}

	return &Prediction{
		ItemName:       itemName,
		Category:       category,
		BestMonths:     bestMonths,
		AvgDiscountPct: discount,
		Events:         events,
		NextEventDays:  nextDays,
		Rationale:      rationale,
		Confidence:     confidence,
	}
}

// seasonalTool returns the LLM tool schema for the seasonal_prediction function.
func seasonalTool() llm.Tool {
	return llm.Tool{
		Name:        "seasonal_prediction",
		Description: "Predict the best months and shopping events to buy a product on the French market, with expected discount percentage and rationale.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"best_months", "avg_discount_pct", "events", "next_event_days", "rationale", "confidence"},
			"properties": map[string]any{
				"best_months": map[string]any{
					"type":        "array",
					"description": "List of best months to buy in French (e.g. 'Novembre', 'Décembre')",
					"items": map[string]any{
						"type": "string",
					},
					"minItems": 1,
				},
				"avg_discount_pct": map[string]any{
					"type":        "number",
					"description": "Average expected discount percentage during the best buying periods (0-100)",
					"minimum":     0,
					"maximum":     100,
				},
				"events": map[string]any{
					"type":        "array",
					"description": "French market shopping events relevant to this product (e.g. 'Black Friday', 'Soldes d\\'hiver')",
					"items": map[string]any{
						"type": "string",
					},
				},
				"next_event_days": map[string]any{
					"type":        "integer",
					"description": "Estimated number of days until the next best buying opportunity",
					"minimum":     0,
				},
				"rationale": map[string]any{
					"type":        "string",
					"description": "1-2 sentences in French explaining why these months/events are best for this product",
				},
				"confidence": map[string]any{
					"type":        "string",
					"description": "Confidence level of the prediction",
					"enum":        []string{"high", "medium", "low"},
				},
			},
		},
	}
}
