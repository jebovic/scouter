package purchaseadvisor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

const cacheTTL = 2 * time.Hour

// cacheEntry stores a cached advice with an expiry timestamp.
type cacheEntry struct {
	advice    PurchaseAdvice
	expiresAt time.Time
}

// PurchaseAdvisor uses the LLM to generate personalized purchase recommendations.
type PurchaseAdvisor struct {
	provider llm.Provider
	cache    sync.Map // key: itemID (string) → cacheEntry
}

// NewPurchaseAdvisor creates a new PurchaseAdvisor backed by the given LLM provider.
// Panics if provider is nil.
func NewPurchaseAdvisor(provider llm.Provider) *PurchaseAdvisor {
	if provider == nil {
		panic("purchaseadvisor.NewPurchaseAdvisor: provider must not be nil")
	}
	return &PurchaseAdvisor{provider: provider}
}

// Advise calls the LLM with a French-market focused prompt and returns a
// structured PurchaseAdvice for the item. Results are cached for 2 hours.
// Falls back to deterministic logic if the LLM call fails.
func (a *PurchaseAdvisor) Advise(ctx context.Context, req AdvisorRequest) (PurchaseAdvice, error) {
	// Serve from in-memory cache.
	if entry, ok := a.cache.Load(req.ItemID); ok {
		e := entry.(cacheEntry)
		if time.Now().Before(e.expiresAt) {
			return e.advice, nil
		}
		a.cache.Delete(req.ItemID)
	}

	advice, err := a.callLLM(ctx, req)
	if err != nil {
		// Deterministic fallback on LLM failure.
		advice = deterministicFallback(req)
	}

	a.cache.Store(req.ItemID, cacheEntry{advice: advice, expiresAt: time.Now().Add(cacheTTL)})
	return advice, nil
}

// callLLM invokes the LLM with tool-use and parses the structured response.
func (a *PurchaseAdvisor) callLLM(ctx context.Context, req AdvisorRequest) (PurchaseAdvice, error) {
	completionReq := buildRequest(req)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "purchase-advisor",
	})

	resp, err := a.provider.Complete(llmCtx, completionReq)
	if err != nil {
		return PurchaseAdvice{}, fmt.Errorf("purchase advisor llm call: %w", err)
	}

	advice, err := parseToolResponse(resp, req)
	if err != nil {
		// Tool call missing — retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "purchase-advisor-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, completionReq, "record_advice")
		if err != nil {
			return PurchaseAdvice{}, fmt.Errorf("purchase advisor json fallback: %w", err)
		}
		advice, err = parseToolResponse(resp, req)
		if err != nil {
			return PurchaseAdvice{}, fmt.Errorf("parse purchase advisor response: %w", err)
		}
	}

	return advice, nil
}

// deterministicFallback returns a rule-based decision when the LLM is unavailable.
func deterministicFallback(req AdvisorRequest) PurchaseAdvice {
	decision := DecisionWait
	reasoning := "Attendez une meilleure opportunité ou une baisse de prix."

	if req.TargetPrice > 0 && req.CurrentPrice <= req.TargetPrice {
		decision = DecisionBuyNow
		reasoning = "Le prix actuel est inférieur ou égal à votre objectif. C'est le bon moment d'acheter."
	} else if req.Status == "flash-sale" {
		decision = DecisionBuyNow
		reasoning = "Vente flash en cours. Les promotions flash sont éphémères — agissez maintenant."
	}

	return PurchaseAdvice{
		ItemID:    req.ItemID,
		ItemName:  req.ItemName,
		Decision:  decision,
		Confidence: 50,
		Reasoning: reasoning,
	}
}

// buildRequest constructs the LLM completion request with the record_advice tool.
func buildRequest(req AdvisorRequest) llm.CompletionRequest {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Vous êtes un conseiller d'achat expert pour le marché français.\n\n")
	fmt.Fprintf(&sb, "Produit : %s\n", req.ItemName)
	fmt.Fprintf(&sb, "Catégorie : %s\n", req.Category)
	fmt.Fprintf(&sb, "Prix actuel : %.2f €\n", req.CurrentPrice)
	if req.TargetPrice > 0 {
		fmt.Fprintf(&sb, "Prix cible : %.2f €\n", req.TargetPrice)
	}
	if req.Budget > 0 {
		fmt.Fprintf(&sb, "Budget total : %.2f €\n", req.Budget)
	}
	fmt.Fprintf(&sb, "Statut actuel : %s\n", req.Status)

	if len(req.PriceHistory) > 0 {
		sb.WriteString("Historique des prix (€) : ")
		for i, p := range req.PriceHistory {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%.2f", p)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nAnalysez ces données et fournissez une recommandation d'achat personnalisée. ")
	sb.WriteString("Utilisez l'outil record_advice pour retourner votre analyse structurée en français.")

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: sb.String()}},
		Tools:     []llm.Tool{advisorTool()},
		MaxTokens: 1024,
	}
}

// advisorTool returns the LLM tool schema for the record_advice function.
func advisorTool() llm.Tool {
	return llm.Tool{
		Name:        "record_advice",
		Description: "Enregistre une recommandation d'achat personnalisée structurée pour un article du panier.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"decision", "confidence", "reasoning"},
			"properties": map[string]any{
				"decision": map[string]any{
					"type":        "string",
					"enum":        []string{DecisionBuyNow, DecisionWait, DecisionSkip, DecisionNegotiate},
					"description": "Recommandation principale : acheter maintenant, attendre, ignorer ou négocier.",
				},
				"confidence": map[string]any{
					"type":        "integer",
					"minimum":     0,
					"maximum":     100,
					"description": "Niveau de confiance de la recommandation (0 = très incertain, 100 = très certain).",
				},
				"reasoning": map[string]any{
					"type":        "string",
					"description": "Explication détaillée en français justifiant la recommandation (2-4 phrases).",
				},
				"alternativeSuggestion": map[string]any{
					"type":        "string",
					"description": "Suggestion d'alternative ou d'action complémentaire en français (facultatif).",
				},
				"bestTimeToAct": map[string]any{
					"type":        "string",
					"description": "Meilleur moment pour agir, ex. 'Vendredi Noir', 'Soldes de janvier', 'Maintenant' (facultatif).",
				},
				"estimatedSavingsByWaiting": map[string]any{
					"type":        "number",
					"description": "Économies estimées en euros si l'acheteur attend une meilleure occasion (0 si recommandation immédiate).",
				},
			},
		},
	}
}

// toolPayload mirrors the record_advice tool output shape for unmarshalling.
type toolPayload struct {
	Decision                  string  `json:"decision"`
	Confidence                int     `json:"confidence"`
	Reasoning                 string  `json:"reasoning"`
	AlternativeSuggestion     string  `json:"alternativeSuggestion"`
	BestTimeToAct             string  `json:"bestTimeToAct"`
	EstimatedSavingsByWaiting float64 `json:"estimatedSavingsByWaiting"`
}

// parseToolResponse extracts PurchaseAdvice from the first record_advice tool call.
func parseToolResponse(resp llm.CompletionResponse, req AdvisorRequest) (PurchaseAdvice, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "record_advice" {
			continue
		}
		raw, err := json.Marshal(tc.Input)
		if err != nil {
			return PurchaseAdvice{}, fmt.Errorf("re-marshal tool input: %w", err)
		}
		var p toolPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return PurchaseAdvice{}, fmt.Errorf("unmarshal advisor payload: %w", err)
		}
		// Clamp confidence to valid range.
		if p.Confidence < 0 {
			p.Confidence = 0
		}
		if p.Confidence > 100 {
			p.Confidence = 100
		}
		// Validate decision value; default to "wait" if unrecognised.
		switch p.Decision {
		case DecisionBuyNow, DecisionWait, DecisionSkip, DecisionNegotiate:
		default:
			p.Decision = DecisionWait
		}
		return PurchaseAdvice{
			ItemID:                    req.ItemID,
			ItemName:                  req.ItemName,
			Decision:                  p.Decision,
			Confidence:                p.Confidence,
			Reasoning:                 p.Reasoning,
			AlternativeSuggestion:     p.AlternativeSuggestion,
			BestTimeToAct:             p.BestTimeToAct,
			EstimatedSavingsByWaiting: p.EstimatedSavingsByWaiting,
		}, nil
	}
	return PurchaseAdvice{}, fmt.Errorf("LLM did not call record_advice")
}
