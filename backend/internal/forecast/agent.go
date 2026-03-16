package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/llm"
)

// ForecastAgent uses the LLM to produce a structured budget forecast
// for a mission. It owns prompt construction, tool schema, response parsing,
// and persistence. The llm.Provider is transport-only.
type ForecastAgent struct {
	provider llm.Provider
	repo     Repository
	pool     *pgxpool.Pool
}

// NewForecastAgent creates a new ForecastAgent.
func NewForecastAgent(provider llm.Provider, repo Repository, pool *pgxpool.Pool) *ForecastAgent {
	return &ForecastAgent{
		provider: provider,
		repo:     repo,
		pool:     pool,
	}
}

// missionRow is a lightweight projection of missions used by the agent.
type missionRow struct {
	ID       string
	Name     string
	Budget   float64
	Currency string
	Category string
	Phase    string
}

// shoppingItemRow is a lightweight projection of shopping_items.
type shoppingItemRow struct {
	Name  string
	Price float64
}

// Run fetches mission context from the DB, calls the LLM, persists, and returns
// the ForecastResult.
func (a *ForecastAgent) Run(ctx context.Context, missionID string) (ForecastResult, error) {
	m, err := a.fetchMission(ctx, missionID)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("fetch mission: %w", err)
	}

	items, err := a.fetchShoppingItems(ctx, missionID)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("fetch shopping items: %w", err)
	}

	req := a.buildRequest(m, items)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "forecast",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return a.computedFallback(ctx, missionID, m, items), nil
	}

	partial, err := parseToolResponse(resp)
	if err != nil {
		// Retry in JSON-only mode using a fresh timeout.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "forecast_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "budget_forecast")
		if err != nil {
			return a.computedFallback(ctx, missionID, m, items), nil
		}
		partial, err = parseToolResponse(resp)
		if err != nil {
			return a.computedFallback(ctx, missionID, m, items), nil
		}
	}

	partial.MissionID = missionID
	return a.repo.Save(ctx, partial)
}

func (a *ForecastAgent) fetchMission(ctx context.Context, missionID string) (missionRow, error) {
	var m missionRow
	err := a.pool.QueryRow(ctx,
		`SELECT id, name, budget, currency, category, phase FROM missions WHERE id = $1`,
		missionID,
	).Scan(&m.ID, &m.Name, &m.Budget, &m.Currency, &m.Category, &m.Phase)
	if err != nil {
		return missionRow{}, fmt.Errorf("query mission: %w", err)
	}
	return m, nil
}

func (a *ForecastAgent) fetchShoppingItems(ctx context.Context, missionID string) ([]shoppingItemRow, error) {
	rows, err := a.pool.Query(ctx,
		`SELECT name, price FROM shopping_items WHERE mission_id = $1 LIMIT 20`,
		missionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query shopping items: %w", err)
	}
	defer rows.Close()

	var items []shoppingItemRow
	for rows.Next() {
		var item shoppingItemRow
		if err := rows.Scan(&item.Name, &item.Price); err != nil {
			return nil, fmt.Errorf("scan shopping item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *ForecastAgent) buildRequest(m missionRow, items []shoppingItemRow) llm.CompletionRequest {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Mission: %s\n", m.Name))
	sb.WriteString(fmt.Sprintf("Category: %s\n", m.Category))
	sb.WriteString(fmt.Sprintf("Budget: %.2f %s\n", m.Budget, m.Currency))
	sb.WriteString(fmt.Sprintf("Phase: %s\n", m.Phase))

	var totalSpent float64
	if len(items) > 0 {
		sb.WriteString("\nShopping items:\n")
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("  - %s: %.2f %s\n", item.Name, item.Price, m.Currency))
			totalSpent += item.Price
		}
		sb.WriteString(fmt.Sprintf("\nTotal spent so far: %.2f %s\n", totalSpent, m.Currency))
		remaining := m.Budget - totalSpent
		sb.WriteString(fmt.Sprintf("Remaining budget: %.2f %s\n", remaining, m.Currency))
	} else {
		sb.WriteString("\nNo shopping items tracked yet.\n")
	}

	prompt := fmt.Sprintf(`You are a budget forecasting expert. Analyze this purchase mission and provide a structured forecast.

%s
Use the budget_forecast tool to return your analysis with estimated total cost, confidence level, 3-5 actionable recommendations, key risk factors, and optionally months needed to save and the optimal buy time.`,
		sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{forecastTool()},
		MaxTokens: 2048,
	}
}

// forecastTool returns the tool schema for structured forecast submission.
func forecastTool() llm.Tool {
	return llm.Tool{
		Name:        "budget_forecast",
		Description: "Provide a structured budget forecast for the mission",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"estimated_total", "confidence", "recommendations", "risk_factors"},
			"properties": map[string]any{
				"estimated_total": map[string]any{
					"type":        "number",
					"description": "Estimated total cost in mission currency",
				},
				"confidence": map[string]any{
					"type": "string",
					"enum": []string{"low", "medium", "high"},
				},
				"recommendations": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "3-5 actionable recommendations",
				},
				"risk_factors": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Key risk factors",
				},
				"months_to_save": map[string]any{
					"type":        "integer",
					"description": "Months needed to save if budget is insufficient, optional",
				},
				"optimal_buy_time": map[string]any{
					"type":        "string",
					"description": "Best time to buy, e.g. 'Black Friday 2026', optional",
				},
			},
		},
	}
}

// computedFallback builds a deterministic ForecastResult from DB data when the LLM
// is unavailable. It is not persisted so the next request will retry the LLM.
func (a *ForecastAgent) computedFallback(_ context.Context, missionID string, m missionRow, items []shoppingItemRow) ForecastResult {
	var totalSpent float64
	for _, item := range items {
		totalSpent += item.Price
	}

	estimated := m.Budget
	if totalSpent > 0 {
		estimated = totalSpent
	}

	recs := []string{
		"Compare prices across multiple merchants before purchasing.",
		"Set a target price alert to track price drops.",
		"Consider buying during seasonal sales for better deals.",
	}
	risks := []string{
		"Price fluctuations may affect final cost.",
		"Budget forecast is estimated — LLM analysis unavailable.",
	}

	confidence := "low"
	return ForecastResult{
		MissionID:       missionID,
		EstimatedTotal:  &estimated,
		Confidence:      confidence,
		Recommendations: recs,
		RiskFactors:     risks,
	}
}

// parseToolResponse extracts a ForecastResult from the first budget_forecast tool call.
func parseToolResponse(resp llm.CompletionResponse) (ForecastResult, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "budget_forecast" {
			continue
		}
		return unmarshalForecast(tc.Input)
	}
	return ForecastResult{}, fmt.Errorf("LLM did not call budget_forecast")
}

func unmarshalForecast(input map[string]any) (ForecastResult, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		EstimatedTotal  *float64 `json:"estimated_total"`
		Confidence      string   `json:"confidence"`
		Recommendations []string `json:"recommendations"`
		RiskFactors     []string `json:"risk_factors"`
		MonthsToSave    *float64 `json:"months_to_save"` // JSON numbers are float64
		OptimalBuyTime  *string  `json:"optimal_buy_time"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return ForecastResult{}, fmt.Errorf("unmarshal forecast payload: %w", err)
	}

	result := ForecastResult{
		EstimatedTotal:  payload.EstimatedTotal,
		Confidence:      payload.Confidence,
		Recommendations: payload.Recommendations,
		RiskFactors:     payload.RiskFactors,
		OptimalBuyTime:  payload.OptimalBuyTime,
	}

	if result.Recommendations == nil {
		result.Recommendations = []string{}
	}
	if result.RiskFactors == nil {
		result.RiskFactors = []string{}
	}

	if payload.MonthsToSave != nil {
		m := int(*payload.MonthsToSave)
		result.MonthsToSave = &m
	}

	return result, nil
}
