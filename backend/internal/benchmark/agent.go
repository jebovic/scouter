package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// BenchmarkAgentInterface is the narrow interface consumed by Handler so tests
// can inject a fake without depending on the real LLM.
type BenchmarkAgentInterface interface {
	Benchmark(ctx context.Context, itemID, itemName string, currentPrice float64) (*BenchmarkResult, error)
}

// Agent uses the LLM to estimate market price ranges and produce a verdict
// for the French market.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new benchmark Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Benchmark calls the LLM to produce a market price range and verdict for the
// given shopping item. The verdict is computed from diffPercent derived from
// the LLM-supplied marketAverage.
func (a *Agent) Benchmark(ctx context.Context, itemID, itemName string, currentPrice float64) (*BenchmarkResult, error) {
	req := a.buildRequest(itemName, currentPrice)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "benchmark-price",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("benchmark llm call: %w", err)
	}

	raw, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "benchmark-price-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "benchmark_price")
		if err != nil {
			return nil, fmt.Errorf("benchmark json fallback: %w", err)
		}
		raw, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse benchmark response: %w", err)
		}
	}

	return buildResult(itemID, itemName, currentPrice, raw), nil
}

// toolPayload is the shape of data returned by the benchmark_price tool.
type toolPayload struct {
	MarketLow     float64 `json:"marketLow"`
	MarketAverage float64 `json:"marketAverage"`
	MarketHigh    float64 `json:"marketHigh"`
	Explanation   string  `json:"explanation"`
}

func (a *Agent) buildRequest(itemName string, currentPrice float64) llm.CompletionRequest {
	prompt := fmt.Sprintf(
		`Vous êtes un expert en prix du marché français. Pour le produit '%s' vendu à %.2f EUR, estimez la fourchette de prix typique sur le marché français (low/average/high). Donnez un verdict: good_deal (>10%% sous la moyenne), average (±10%%), overpriced (>10%% au-dessus). Expliquez en une phrase en français.`,
		itemName, currentPrice,
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{benchmarkTool()},
		MaxTokens: 512,
	}
}

// parseToolResponse extracts the benchmark_price tool call payload.
func parseToolResponse(resp llm.CompletionResponse) (toolPayload, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "benchmark_price" {
			continue
		}
		return unmarshalPayload(tc.Input)
	}
	return toolPayload{}, fmt.Errorf("LLM did not call benchmark_price")
}

func unmarshalPayload(input map[string]any) (toolPayload, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return toolPayload{}, fmt.Errorf("re-marshal tool input: %w", err)
	}
	var p toolPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return toolPayload{}, fmt.Errorf("unmarshal benchmark payload: %w", err)
	}
	return p, nil
}

// buildResult assembles a BenchmarkResult from the LLM payload.
// Verdict is computed from diffPercent to ensure consistency with prompt rules:
//
//	diffPercent < -10 → good_deal
//	diffPercent >  10 → overpriced
//	else              → average
func buildResult(itemID, itemName string, currentPrice float64, p toolPayload) *BenchmarkResult {
	var diffPercent float64
	if p.MarketAverage > 0 {
		diffPercent = (currentPrice - p.MarketAverage) / p.MarketAverage * 100
	}

	verdict := computeVerdict(diffPercent)

	return &BenchmarkResult{
		ItemID:       itemID,
		ItemName:     itemName,
		CurrentPrice: currentPrice,
		Currency:     "EUR",
		MarketRange: PriceRange{
			Low:     p.MarketLow,
			Average: p.MarketAverage,
			High:    p.MarketHigh,
		},
		Verdict:     verdict,
		DiffPercent: diffPercent,
		Explanation: p.Explanation,
		CachedAt:    time.Now().Unix(),
	}
}

// computeVerdict maps a diffPercent to one of the three verdict strings.
func computeVerdict(diffPercent float64) string {
	switch {
	case diffPercent < -10:
		return "good_deal"
	case diffPercent > 10:
		return "overpriced"
	default:
		return "average"
	}
}

// benchmarkTool returns the LLM tool schema for the benchmark_price function.
func benchmarkTool() llm.Tool {
	return llm.Tool{
		Name:        "benchmark_price",
		Description: "Estimate the typical French market price range for a product and provide a verdict.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"marketLow", "marketAverage", "marketHigh", "explanation"},
			"properties": map[string]any{
				"marketLow": map[string]any{
					"type":        "number",
					"description": "Lowest typical market price in EUR",
					"minimum":     0,
				},
				"marketAverage": map[string]any{
					"type":        "number",
					"description": "Average market price in EUR",
					"minimum":     0,
				},
				"marketHigh": map[string]any{
					"type":        "number",
					"description": "Highest typical market price in EUR",
					"minimum":     0,
				},
				"explanation": map[string]any{
					"type":        "string",
					"description": "One-sentence French explanation of the verdict",
				},
			},
		},
	}
}
