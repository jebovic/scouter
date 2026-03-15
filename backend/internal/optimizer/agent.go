package optimizer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/shopping"
)

// OptimizerAgent is the interface satisfied by Agent (and fakes in tests).
type OptimizerAgent interface {
	Optimize(ctx context.Context, items []shopping.Item) (*OptimizedPlan, error)
}

// Agent uses an LLM to build an optimized shopping plan.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new optimizer Agent.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Optimize calls the LLM to build an ordered action plan for the given items.
// If the LLM is unavailable it falls back to a deterministic heuristic plan.
func (a *Agent) Optimize(ctx context.Context, items []shopping.Item) (*OptimizedPlan, error) {
	req := buildRequest(items)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "optimizer",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		// LLM unavailable — use deterministic fallback.
		return buildFallbackPlan("", "EUR", items), nil
	}

	plan, err := parsePlanResponse(resp, items)
	if err != nil {
		// Retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "optimizer_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "optimize_shopping_list")
		if err != nil {
			return buildFallbackPlan("", "EUR", items), nil
		}
		plan, err = parsePlanResponse(resp, items)
		if err != nil {
			return buildFallbackPlan("", "EUR", items), nil
		}
	}

	return plan, nil
}

// buildRequest assembles the LLM CompletionRequest with tool schema and prompt.
func buildRequest(items []shopping.Item) llm.CompletionRequest {
	var sb strings.Builder
	sb.WriteString("Tu es un expert en shopping intelligent. Voici la liste de courses :\n\n")
	for i, it := range items {
		sb.WriteString(fmt.Sprintf("%d. %s — %s — %.2f € (statut: %s)\n",
			i+1, it.Name, it.Merchant, it.Price, it.Status))
	}
	sb.WriteString(`
Analyse cette liste et propose un plan d'achat optimal en français.
Groupe les articles du même marchand pour bénéficier de la livraison gratuite.
Signale les articles en promo ou à surveiller.
Utilise l'outil optimize_shopping_list pour retourner le plan structuré.`)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: sb.String()}},
		Tools:     []llm.Tool{optimizeTool()},
		MaxTokens: 2048,
	}
}

// optimizeTool defines the JSON schema for the LLM tool call.
func optimizeTool() llm.Tool {
	return llm.Tool{
		Name:        "optimize_shopping_list",
		Description: "Retourne un plan d'achat optimisé et ordonné en français.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"actions", "summary"},
			"properties": map[string]any{
				"actions": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"priority", "item_ids", "merchant", "action", "reason"},
						"properties": map[string]any{
							"priority": map[string]any{"type": "integer"},
							"item_ids": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
							"merchant":    map[string]any{"type": "string"},
							"action":      map[string]any{"type": "string"},
							"reason":      map[string]any{"type": "string"},
							"est_savings": map[string]any{"type": "number"},
						},
					},
				},
				"summary": map[string]any{"type": "string"},
			},
		},
	}
}

// parsePlanResponse extracts the OptimizedPlan from the LLM tool call response.
func parsePlanResponse(resp llm.CompletionResponse, items []shopping.Item) (*OptimizedPlan, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "optimize_shopping_list" {
			continue
		}
		return unmarshalPlan(tc.Input, items)
	}
	return nil, fmt.Errorf("LLM did not call optimize_shopping_list")
}

// unmarshalPlan converts raw tool input map into an OptimizedPlan.
func unmarshalPlan(input map[string]any, items []shopping.Item) (*OptimizedPlan, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Actions []struct {
			Priority   int      `json:"priority"`
			ItemIDs    []string `json:"item_ids"`
			Merchant   string   `json:"merchant"`
			Action     string   `json:"action"`
			Reason     string   `json:"reason"`
			EstSavings float64  `json:"est_savings"`
		} `json:"actions"`
		Summary string `json:"summary"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal plan payload: %w", err)
	}

	// Build item name lookup.
	nameByID := make(map[string]string, len(items))
	currencyByID := make(map[string]string, len(items))
	for _, it := range items {
		nameByID[it.ID.String()] = it.Name
		currencyByID[it.ID.String()] = "EUR"
	}

	actions := make([]ShoppingAction, 0, len(payload.Actions))
	for _, a := range payload.Actions {
		names := make([]string, 0, len(a.ItemIDs))
		for _, id := range a.ItemIDs {
			if n, ok := nameByID[id]; ok {
				names = append(names, n)
			}
		}
		currency := "EUR"
		if len(a.ItemIDs) > 0 {
			if c, ok := currencyByID[a.ItemIDs[0]]; ok {
				currency = c
			}
		}
		actions = append(actions, ShoppingAction{
			Priority:   a.Priority,
			ItemIDs:    a.ItemIDs,
			ItemNames:  names,
			Merchant:   a.Merchant,
			Action:     a.Action,
			Reason:     a.Reason,
			EstSavings: a.EstSavings,
			Currency:   currency,
		})
	}

	return &OptimizedPlan{
		GeneratedAt: time.Now().UTC(),
		TotalItems:  len(items),
		Actions:     actions,
		Summary:     payload.Summary,
	}, nil
}

// buildFallbackPlan creates a deterministic plan when the LLM is unavailable.
// Items are sorted by status priority (flash-sale > buy > watch > defer > other).
func buildFallbackPlan(missionID, currency string, items []shopping.Item) *OptimizedPlan {
	actions := make([]ShoppingAction, 0, len(items))
	for i, it := range items {
		action, reason := fallbackActionForStatus(it.Status)
		c := currency
		if c == "" {
			c = "EUR"
		}
		actions = append(actions, ShoppingAction{
			Priority:   i + 1,
			ItemIDs:    []string{it.ID.String()},
			ItemNames:  []string{it.Name},
			Merchant:   it.Merchant,
			Action:     action,
			Reason:     reason,
			EstSavings: 0,
			Currency:   c,
		})
	}

	summary := "Plan d'achat généré automatiquement. Vérifiez les prix avant d'acheter."
	if len(items) == 0 {
		summary = "Aucun article à optimiser."
	}

	return &OptimizedPlan{
		MissionID:   missionID,
		GeneratedAt: time.Now().UTC(),
		TotalItems:  len(items),
		Actions:     actions,
		Summary:     summary,
	}
}

// fallbackActionForStatus maps a shopping status to a French action label and reason.
func fallbackActionForStatus(status string) (action, reason string) {
	switch status {
	case "buy", "flash-sale", "recommended":
		return "Acheter maintenant", "Le statut indique que c'est le bon moment pour acheter."
	case "watch":
		return "Comparer en magasin", "Continuez à surveiller le prix avant de décider."
	case "defer":
		return "Attendre la promo", "Différez cet achat jusqu'à une meilleure offre."
	default:
		return "Comparer en magasin", "Évaluez les options disponibles."
	}
}
