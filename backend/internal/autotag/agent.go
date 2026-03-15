package autotag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// TaggerInterface is the narrow interface consumed by Handler so tests can
// inject a fake without depending on the real LLM.
type TaggerInterface interface {
	Suggest(ctx context.Context, input TagInput) (*CategorySuggestion, error)
}

// Agent uses the LLM to suggest a French category for a mission.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new autotag Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Suggest calls the LLM to suggest a category.
// Returns an error when the LLM is unavailable; callers should fall back to
// KeywordFallback.
func (a *Agent) Suggest(ctx context.Context, input TagInput) (*CategorySuggestion, error) {
	req := buildRequest(input)

	llmCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "autotag-suggest",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("autotag llm call: %w", err)
	}

	sug, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry with JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapText,
			Label:        "autotag-suggest-fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "suggest_category")
		if err != nil {
			return nil, fmt.Errorf("autotag json fallback: %w", err)
		}
		sug, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse autotag response: %w", err)
		}
	}

	if sug.Alternatives == nil {
		sug.Alternatives = []string{}
	}
	return sug, nil
}

func buildRequest(input TagInput) llm.CompletionRequest {
	constraintsPart := ""
	if len(input.Constraints) > 0 {
		constraintsPart = "\nContraintes: " + strings.Join(input.Constraints, ", ")
	}

	categoryList := strings.Join(Categories, ", ")

	prompt := fmt.Sprintf(`Tu es un assistant de catégorisation d'achats pour le marché français.
Analyse cette mission d'achat et propose la meilleure catégorie parmi la liste fournie.

Nom de la mission: %s%s

Catégories disponibles: %s

Utilise l'outil suggest_category pour retourner la catégorie la plus appropriée, un score de confiance entre 0 et 1, et jusqu'à 2 alternatives.`,
		input.Name, constraintsPart, categoryList)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{categorisationTool()},
		MaxTokens: 256,
	}
}

func parseToolResponse(resp llm.CompletionResponse) (*CategorySuggestion, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "suggest_category" {
			continue
		}
		return unmarshalSuggestion(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call suggest_category")
}

func unmarshalSuggestion(input map[string]any) (*CategorySuggestion, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Category     string   `json:"category"`
		Confidence   float64  `json:"confidence"`
		Alternatives []string `json:"alternatives"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal suggestion payload: %w", err)
	}

	if payload.Category == "" {
		return nil, fmt.Errorf("empty category in LLM response")
	}
	if payload.Alternatives == nil {
		payload.Alternatives = []string{}
	}

	return &CategorySuggestion{
		Category:     payload.Category,
		Confidence:   payload.Confidence,
		Alternatives: payload.Alternatives,
	}, nil
}

func categorisationTool() llm.Tool {
	categoryEnum := make([]any, len(Categories))
	for i, c := range Categories {
		categoryEnum[i] = c
	}

	return llm.Tool{
		Name:        "suggest_category",
		Description: "Suggest the best French category for a purchase mission",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"category", "confidence"},
			"properties": map[string]any{
				"category": map[string]any{
					"type":        "string",
					"enum":        categoryEnum,
					"description": "The most appropriate French category",
				},
				"confidence": map[string]any{
					"type":        "number",
					"description": "Confidence score between 0 and 1",
				},
				"alternatives": map[string]any{
					"type":        "array",
					"description": "Up to 2 alternative category suggestions",
					"items":       map[string]any{"type": "string"},
				},
			},
		},
	}
}

// KeywordFallback returns a deterministic CategorySuggestion based on keyword
// matching in the mission name. Used when the LLM is unavailable.
func KeywordFallback(name string) *CategorySuggestion {
	lower := strings.ToLower(name)

	type rule struct {
		keywords []string
		category string
	}
	rules := []rule{
		{[]string{"macbook", "laptop", "ordinateur", "pc", "imac", "ipad", "tablette", "informatique", "logiciel"}, "Informatique"},
		{[]string{"iphone", "samsung", "téléphone", "smartphone", "phone", "airpods", "écouteurs", "casque audio"}, "Électronique"},
		{[]string{"télévision", "tv", "4k", "oled", "projecteur", "home cinéma", "enceinte", "hifi"}, "Audiovisuel"},
		{[]string{"vol", "avion", "hôtel", "séjour", "voyage", "vacances", "airbnb", "billet"}, "Voyage"},
		{[]string{"train", "bus", "voiture location", "transport", "navigo", "abonnement transport"}, "Transport"},
		{[]string{"nourriture", "alimentation", "épicerie", "repas", "restaurant", "traiteur"}, "Alimentation"},
		{[]string{"cuisine", "casserole", "robot culinaire", "four", "réfrigérateur", "lave-vaisselle"}, "Cuisine"},
		{[]string{"vêtement", "veste", "pantalon", "robe", "chemise", "pull", "manteau", "mode"}, "Mode"},
		{[]string{"chaussure", "basket", "sneaker", "botte", "sandale"}, "Chaussures"},
		{[]string{"canapé", "meuble", "lit", "armoire", "bureau", "table", "chaise", "bibliothèque"}, "Mobilier"},
		{[]string{"décoration", "tapis", "rideau", "lampe", "luminaire", "tableau", "cadre"}, "Décoration"},
		{[]string{"maison", "appartement", "rénovation", "travaux", "bricolage", "jardinage"}, "Maison"},
		{[]string{"sport", "vélo", "trottinette", "gym", "fitness", "musculation", "running", "ski"}, "Sport"},
		{[]string{"jeu", "console", "playstation", "xbox", "nintendo", "gaming", "ps5"}, "Jeux"},
		{[]string{"santé", "médicament", "pharmacie", "médecin", "mutuelle", "chirurgie"}, "Santé"},
		{[]string{"beauté", "cosmétique", "parfum", "soin", "maquillage", "shampoing"}, "Beauté"},
		{[]string{"voiture", "auto", "automobile", "bmw", "renault", "peugeot", "assurance auto"}, "Auto"},
		{[]string{"moto", "scooter", "deux roues"}, "Moto"},
		{[]string{"livre", "roman", "bd", "manga", "librairie", "kindle"}, "Livres"},
		{[]string{"musique", "guitare", "piano", "violon", "instrument"}, "Musique"},
		{[]string{"cinema", "spectacle", "concert", "culture", "musée", "exposition"}, "Culture"},
		{[]string{"abonnement", "netflix", "spotify", "amazon prime", "apple one", "licence"}, "Abonnements"},
		{[]string{"assurance", "service", "prestation", "consultation", "artisan"}, "Services"},
	}

	for _, r := range rules {
		for _, kw := range r.keywords {
			if strings.Contains(lower, kw) {
				return &CategorySuggestion{
					Category:     r.category,
					Confidence:   0.6,
					Alternatives: []string{},
				}
			}
		}
	}

	return &CategorySuggestion{
		Category:     "Autre",
		Confidence:   0.3,
		Alternatives: []string{},
	}
}
