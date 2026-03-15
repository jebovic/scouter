package priceforecast

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// ForecastAgentInterface is the narrow interface consumed by Handler so tests
// can inject a fake without depending on the real LLM.
type ForecastAgentInterface interface {
	Forecast(ctx context.Context, req ForecastRequest) (*PriceForecastResult, error)
}

// Agent uses the LLM to generate a 30-day price forecast with confidence
// intervals for a shopping item.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new ForecastAgent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Forecast calls the LLM to produce a price trajectory forecast with
// confidence bounds, a French summary, and the best buy window.
func (a *Agent) Forecast(ctx context.Context, req ForecastRequest) (*PriceForecastResult, error) {
	llmReq := a.buildRequest(req)

	llmCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "price-forecast",
	})

	resp, err := a.provider.Complete(llmCtx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("price forecast llm call: %w", err)
	}

	raw, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "price-forecast-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, llmReq, "generate_price_forecast")
		if err != nil {
			return nil, fmt.Errorf("price forecast json fallback: %w", err)
		}
		raw, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse price forecast response: %w", err)
		}
	}

	return buildResult(req, raw), nil
}

// toolPayload is the shape of data returned by the generate_price_forecast tool.
type toolPayload struct {
	ForecastPoints []struct {
		Date  string  `json:"date"`
		Low   float64 `json:"low"`
		Mid   float64 `json:"mid"`
		High  float64 `json:"high"`
	} `json:"forecastPoints"`
	Confidence    float64 `json:"confidence"`
	Summary       string  `json:"summary"`
	BestBuyWindow string  `json:"bestBuyWindow"`
}

func (a *Agent) buildRequest(req ForecastRequest) llm.CompletionRequest {
	histLines := make([]string, 0, len(req.Historical))
	for _, pt := range req.Historical {
		histLines = append(histLines, fmt.Sprintf("  - %s: %.2f %s", pt.Date.Format("2006-01-02"), pt.Price, req.Currency))
	}
	histStr := strings.Join(histLines, "\n")
	if histStr == "" {
		histStr = "  (aucun historique disponible)"
	}

	horizon := req.HorizonDays
	if horizon <= 0 {
		horizon = 30
	}

	prompt := fmt.Sprintf(
		`Tu es un expert en prévision de prix pour le marché français. Analyse l'historique de prix suivant pour "%s" et génère une prévision pour les %d prochains jours.

Historique (du plus ancien au plus récent) :
%s

Génère une prévision jour par jour avec des intervalles de confiance bas/médian/haut. Donne également :
- Un score de confiance global entre 0 et 1
- Un résumé en français expliquant la tendance (1-2 phrases)
- La meilleure fenêtre d'achat (ex: "dans 2-3 semaines", "maintenant", "dans 1 mois")

Utilise la tendance des prix passés pour extrapoler intelligemment.`,
		req.ItemName, horizon, histStr,
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{forecastTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts the generate_price_forecast tool call payload.
func parseToolResponse(resp llm.CompletionResponse) (toolPayload, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "generate_price_forecast" {
			continue
		}
		return unmarshalPayload(tc.Input)
	}
	return toolPayload{}, fmt.Errorf("LLM did not call generate_price_forecast")
}

func unmarshalPayload(input map[string]any) (toolPayload, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return toolPayload{}, fmt.Errorf("re-marshal tool input: %w", err)
	}
	var p toolPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return toolPayload{}, fmt.Errorf("unmarshal forecast payload: %w", err)
	}
	return p, nil
}

// buildResult assembles a PriceForecastResult from the LLM payload.
func buildResult(req ForecastRequest, p toolPayload) *PriceForecastResult {
	conf := p.Confidence
	if conf < 0 {
		conf = 0
	}
	if conf > 1 {
		conf = 1
	}

	forecastPts := make([]ForecastPoint, 0, len(p.ForecastPoints))
	for _, fp := range p.ForecastPoints {
		var d time.Time
		if t, err := time.Parse("2006-01-02", fp.Date); err == nil {
			d = t
		} else {
			d = time.Now()
		}
		forecastPts = append(forecastPts, ForecastPoint{
			Date: d,
			Low:  fp.Low,
			Mid:  fp.Mid,
			High: fp.High,
		})
	}

	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}

	return &PriceForecastResult{
		ItemID:        req.ItemID,
		ItemName:      req.ItemName,
		Currency:      currency,
		Historical:    req.Historical,
		Forecast:      forecastPts,
		Confidence:    conf,
		Summary:       p.Summary,
		BestBuyWindow: p.BestBuyWindow,
		GeneratedAt:   time.Now(),
	}
}

// forecastTool returns the LLM tool schema for the generate_price_forecast function.
func forecastTool() llm.Tool {
	return llm.Tool{
		Name:        "generate_price_forecast",
		Description: "Generate a price forecast trajectory with confidence intervals for a shopping item based on historical prices.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"forecastPoints", "confidence", "summary", "bestBuyWindow"},
			"properties": map[string]any{
				"forecastPoints": map[string]any{
					"type":        "array",
					"description": "Daily price forecast points for the next horizon days",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"date", "low", "mid", "high"},
						"properties": map[string]any{
							"date": map[string]any{
								"type":        "string",
								"description": "ISO date string YYYY-MM-DD",
							},
							"low": map[string]any{
								"type":        "number",
								"description": "Lower confidence bound price",
								"minimum":     0,
							},
							"mid": map[string]any{
								"type":        "number",
								"description": "Median predicted price",
								"minimum":     0,
							},
							"high": map[string]any{
								"type":        "number",
								"description": "Upper confidence bound price",
								"minimum":     0,
							},
						},
					},
				},
				"confidence": map[string]any{
					"type":        "number",
					"description": "Overall forecast confidence between 0 (low) and 1 (high)",
					"minimum":     0,
					"maximum":     1,
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "1-2 sentence French summary of the price trend and recommendation",
				},
				"bestBuyWindow": map[string]any{
					"type":        "string",
					"description": "French description of the best time to buy (e.g. 'dans 2-3 semaines', 'maintenant')",
				},
			},
		},
	}
}
