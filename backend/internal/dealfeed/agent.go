package dealfeed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// PromoAgent is the interface satisfied by PromoFinderAgent (and fakes in tests).
type PromoAgent interface {
	FindPromos(ctx context.Context, query, category string) (*PromoFeedResult, error)
}

// PromoFinderAgent uses the LLM to generate realistic French promo offers.
type PromoFinderAgent struct {
	provider llm.Provider
}

// NewPromoFinderAgent creates a new PromoFinderAgent backed by the given LLM provider.
func NewPromoFinderAgent(provider llm.Provider) *PromoFinderAgent {
	return &PromoFinderAgent{provider: provider}
}

// FindPromos calls the LLM to generate 4-6 realistic French promo offers for
// the given product query and category. Uses CapTextOnly for text generation.
func (a *PromoFinderAgent) FindPromos(ctx context.Context, query, category string) (*PromoFeedResult, error) {
	req := a.buildRequest(query, category)

	llmCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "promo_finder",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("promo finder llm call: %w", err)
	}

	offers, err := parsePromoResponse(resp)
	if err != nil {
		// Retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapText,
			Label:        "promo_finder_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "find_french_promos")
		if err != nil {
			return nil, fmt.Errorf("promo finder json fallback: %w", err)
		}
		offers, err = parsePromoResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse promo response: %w", err)
		}
	}

	return &PromoFeedResult{
		Offers:    offers,
		Query:     query,
		Category:  category,
		FetchedAt: time.Now().UTC(),
	}, nil
}

// buildRequest assembles the LLM CompletionRequest with tool schema and prompt.
func (a *PromoFinderAgent) buildRequest(query, category string) llm.CompletionRequest {
	cat := category
	if cat == "" {
		cat = "général"
	}
	// System context is prepended into the user message since CompletionRequest
	// has no dedicated System field.
	content := fmt.Sprintf(
		`Tu es un expert des bons plans français. Tu connais parfaitement les prix du marché français et les promotions actuelles chez Fnac, Darty, Boulanger, Amazon FR, Cdiscount, Leclerc, Carrefour, LDLC, Materiel.net, Grosbill.

Génère 4 à 6 promotions réalistes et actuelles du marché français pour : "%s" (catégorie : %s).

Utilise uniquement des revendeurs français : Fnac, Darty, Boulanger, Amazon.fr, Cdiscount, Leclerc, Carrefour, LDLC, Materiel.net, Grosbill.
Les remises doivent être réalistes (5%% à 60%%).
Utilise l'outil find_french_promos pour retourner les offres structurées.`,
		query, cat,
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: content}},
		Tools:     []llm.Tool{promoTool()},
		MaxTokens: 2048,
	}
}

// promoTool returns the JSON schema for the LLM tool call.
func promoTool() llm.Tool {
	return llm.Tool{
		Name:        "find_french_promos",
		Description: "Retourne 4-6 promotions actuelles du marché français pour un produit donné.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"offers"},
			"properties": map[string]any{
				"offers": map[string]any{
					"type":     "array",
					"minItems": 4,
					"maxItems": 6,
					"items": map[string]any{
						"type":     "object",
						"required": []string{"title", "retailer", "original_price", "promo_price", "discount_pct", "url_hint", "is_flash"},
						"properties": map[string]any{
							"title": map[string]any{
								"type":        "string",
								"description": "Titre de l'offre promotionnelle",
							},
							"retailer": map[string]any{
								"type":        "string",
								"description": "Nom du revendeur français",
							},
							"original_price": map[string]any{
								"type":        "number",
								"description": "Prix original en EUR",
								"minimum":     0,
							},
							"promo_price": map[string]any{
								"type":        "number",
								"description": "Prix promotionnel en EUR",
								"minimum":     0,
							},
							"discount_pct": map[string]any{
								"type":        "integer",
								"description": "Pourcentage de remise (0-100)",
								"minimum":     0,
								"maximum":     100,
							},
							"url_hint": map[string]any{
								"type":        "string",
								"description": "Slug URL du produit, ex: fnac.com/produit-exemple",
							},
							"is_flash": map[string]any{
								"type":        "boolean",
								"description": "Vrai si c'est une vente flash",
							},
							"expires_days": map[string]any{
								"type":        "integer",
								"description": "Nombre de jours avant expiration (0 = pas d'expiration)",
								"minimum":     0,
							},
						},
					},
				},
			},
		},
	}
}

// parsePromoResponse extracts offers from the first find_french_promos tool call.
func parsePromoResponse(resp llm.CompletionResponse) ([]PromoOffer, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "find_french_promos" {
			continue
		}
		return unmarshalOffers(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call find_french_promos")
}

// unmarshalOffers converts raw tool input into PromoOffer slice.
func unmarshalOffers(input map[string]any) ([]PromoOffer, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Offers []struct {
			Title         string  `json:"title"`
			Retailer      string  `json:"retailer"`
			OriginalPrice float64 `json:"original_price"`
			PromoPrice    float64 `json:"promo_price"`
			DiscountPct   int     `json:"discount_pct"`
			URLHint       string  `json:"url_hint"`
			IsFlash       bool    `json:"is_flash"`
			ExpiresDays   int     `json:"expires_days"`
		} `json:"offers"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal offers payload: %w", err)
	}

	now := time.Now().UTC()
	offers := make([]PromoOffer, 0, len(payload.Offers))
	for _, o := range payload.Offers {
		offer := PromoOffer{
			Title:         o.Title,
			Retailer:      o.Retailer,
			OriginalPrice: o.OriginalPrice,
			PromoPrice:    o.PromoPrice,
			Discount:      o.DiscountPct,
			URL:           o.URLHint,
			IsFlash:       o.IsFlash,
			Source:        "ai_generated",
		}
		if o.ExpiresDays > 0 {
			t := now.Add(time.Duration(o.ExpiresDays) * 24 * time.Hour)
			offer.ExpiresAt = &t
		}
		offers = append(offers, offer)
	}

	return offers, nil
}
