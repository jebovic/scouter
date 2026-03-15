package rebalancer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
)

// Agent calls the LLM to produce a RebalancePlan.
// It owns prompt construction, tool schema, and response parsing.
// The llm.Provider is transport-only.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new rebalancer Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Rebalance calls the LLM with envelope + mission data and returns a plan.
func (a *Agent) Rebalance(ctx context.Context, envelopes []envelope.Envelope, missions []mission.Mission) (*RebalancePlan, error) {
	req := a.buildRequest(envelopes, missions)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "rebalancer",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("rebalancer llm call: %w", err)
	}

	plan, err := parseToolResponse(resp, envelopes)
	if err != nil {
		// Retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "rebalancer_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "rebalance_budget")
		if err != nil {
			return nil, fmt.Errorf("rebalancer json fallback: %w", err)
		}
		plan, err = parseToolResponse(resp, envelopes)
		if err != nil {
			return nil, fmt.Errorf("parse rebalancer response: %w", err)
		}
	}

	return plan, nil
}

func (a *Agent) buildRequest(envelopes []envelope.Envelope, missions []mission.Mission) llm.CompletionRequest {
	var sb strings.Builder

	sb.WriteString("Enveloppes budgétaires mensuelles:\n")
	for _, env := range envelopes {
		sb.WriteString(fmt.Sprintf("  - %s: %.2f %s/mois (id: %s)\n",
			env.Name, env.MonthlyAmount, env.Currency, env.ID.String()))
	}

	sb.WriteString("\nMissions (projets de dépenses):\n")
	for _, m := range missions {
		envRef := "(sans enveloppe)"
		if m.EnvelopeID != nil {
			envRef = fmt.Sprintf("(enveloppe: %s)", m.EnvelopeID.String())
		}
		sb.WriteString(fmt.Sprintf("  - %s: budget %.2f %s, phase: %s %s\n",
			m.Name, m.Budget, m.Currency, m.Phase, envRef))
	}

	prompt := fmt.Sprintf(`Vous êtes un conseiller financier personnel. Analysez les enveloppes budgétaires suivantes et l'historique de dépenses des missions associées. Suggérez des ajustements pour mieux refléter les dépenses réelles. Soyez précis et donnez des raisons concrètes en français.

%s

Utilisez l'outil rebalance_budget pour retourner vos suggestions.`, sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{rebalanceTool()},
		MaxTokens: 2048,
	}
}

// rebalanceTool returns the tool schema for the rebalance_budget tool.
func rebalanceTool() llm.Tool {
	return llm.Tool{
		Name:        "rebalance_budget",
		Description: "Suggère des ajustements budgétaires pour chaque enveloppe en fonction des dépenses réelles",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"adjustments", "summary"},
			"properties": map[string]any{
				"adjustments": map[string]any{
					"type":        "array",
					"description": "Liste des ajustements suggérés par enveloppe",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"envelopeId", "suggestedMonthly", "reason"},
						"properties": map[string]any{
							"envelopeId": map[string]any{
								"type":        "string",
								"description": "UUID de l'enveloppe",
							},
							"suggestedMonthly": map[string]any{
								"type":        "number",
								"description": "Montant mensuel suggéré",
							},
							"reason": map[string]any{
								"type":        "string",
								"description": "Explication en français du changement suggéré",
							},
						},
					},
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "Résumé exécutif en français de la recommandation globale",
				},
			},
		},
	}
}

// rawAdjustment is the parsed shape of each item from the LLM tool call.
type rawAdjustment struct {
	EnvelopeID       string  `json:"envelopeId"`
	SuggestedMonthly float64 `json:"suggestedMonthly"`
	Reason           string  `json:"reason"`
}

// parseToolResponse extracts a RebalancePlan from the LLM response.
// It merges the LLM suggestions with the known envelope data to fill in
// CurrentMonthly, DeltaPercent, and totals.
func parseToolResponse(resp llm.CompletionResponse, envelopes []envelope.Envelope) (*RebalancePlan, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "rebalance_budget" {
			continue
		}
		return unmarshalPlan(tc.Input, envelopes)
	}
	return nil, fmt.Errorf("LLM did not call rebalance_budget")
}

func unmarshalPlan(input map[string]any, envelopes []envelope.Envelope) (*RebalancePlan, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Adjustments []rawAdjustment `json:"adjustments"`
		Summary     string          `json:"summary"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal rebalance payload: %w", err)
	}

	// Build a lookup for quick current-amount resolution.
	envelopeByID := make(map[string]envelope.Envelope, len(envelopes))
	for _, env := range envelopes {
		envelopeByID[env.ID.String()] = env
	}

	adjustments := make([]EnvelopeAdjustment, 0, len(payload.Adjustments))
	var totalCurrent, totalSuggested float64

	for _, raw := range payload.Adjustments {
		env, ok := envelopeByID[raw.EnvelopeID]
		if !ok {
			continue // skip unknown envelope IDs returned by the LLM
		}
		current := env.MonthlyAmount
		suggested := raw.SuggestedMonthly
		var delta float64
		if current > 0 {
			delta = ((suggested - current) / current) * 100
		}
		adjustments = append(adjustments, EnvelopeAdjustment{
			EnvelopeID:       raw.EnvelopeID,
			EnvelopeName:     env.Name,
			CurrentMonthly:   current,
			SuggestedMonthly: suggested,
			DeltaPercent:     delta,
			Reason:           raw.Reason,
		})
		totalCurrent += current
		totalSuggested += suggested
	}

	// For envelopes not mentioned by the LLM, add a zero-delta entry so the
	// frontend always has the full picture.
	mentioned := make(map[string]struct{}, len(payload.Adjustments))
	for _, adj := range payload.Adjustments {
		mentioned[adj.EnvelopeID] = struct{}{}
	}
	for _, env := range envelopes {
		if _, ok := mentioned[env.ID.String()]; ok {
			continue
		}
		adjustments = append(adjustments, EnvelopeAdjustment{
			EnvelopeID:       env.ID.String(),
			EnvelopeName:     env.Name,
			CurrentMonthly:   env.MonthlyAmount,
			SuggestedMonthly: env.MonthlyAmount,
			DeltaPercent:     0,
			Reason:           "Allocation optimale — aucun ajustement nécessaire.",
		})
		totalCurrent += env.MonthlyAmount
		totalSuggested += env.MonthlyAmount
	}

	return &RebalancePlan{
		Adjustments:    adjustments,
		Summary:        payload.Summary,
		TotalCurrent:   totalCurrent,
		TotalSuggested: totalSuggested,
		CachedAt:       time.Now().Unix(),
	}, nil
}
