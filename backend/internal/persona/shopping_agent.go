package persona

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// knownArchetypes is the set of French shopping archetypes the LLM may assign.
var knownArchetypes = []string{
	"Le Chasseur de Deals",
	"Le Chercheur",
	"L'Impulsif",
	"Le Budgetaire",
	"Le Premium",
	"Le Prudent",
}

// defaultShoppingPersona is returned when data is insufficient (< 2 missions).
var defaultShoppingPersona = ShoppingPersona{
	Archetype:   "Nouveau Scouter",
	Icon:        "🌱",
	Description: "Vous commencez tout juste votre parcours d'acheteur averti. Ajoutez des missions pour débloquer votre archétype personnalisé.",
	Strengths:   []string{"Curieux et ouvert", "Prêt à apprendre", "Sans habitudes à corriger"},
	Blindspots:  []string{"Manque de données pour une analyse complète", "Pas encore de tendances identifiables"},
	Tips: []string{
		"Créez au moins 2 missions pour débloquer votre profil",
		"Suivez vos achats dans l'onglet Shopping",
		"Comparez les options avant d'acheter",
	},
	ScoreProfile: map[string]int{
		"frugalité":   50,
		"impulsivité": 50,
		"recherche":   50,
		"budget":      50,
		"qualité":     50,
	},
}

// shoppingPersonaTool is the LLM tool schema for analyze_shopping_persona.
func shoppingPersonaTool() llm.Tool {
	return llm.Tool{
		Name:        "analyze_shopping_persona",
		Description: "Analyse le profil d'achat et génère un archétype d'acheteur français",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"archetype", "icon", "description", "strengths", "blindspots", "tips", "scoreProfile"},
			"properties": map[string]any{
				"archetype": map[string]any{
					"type":        "string",
					"description": "Archétype français parmi: Le Chasseur de Deals, Le Chercheur, L'Impulsif, Le Budgetaire, Le Premium, Le Prudent",
				},
				"icon": map[string]any{
					"type":        "string",
					"description": "Emoji représentant l'archétype (ex: 🎯 pour Le Chasseur de Deals)",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Description de 2-3 phrases en français de cet archétype d'acheteur",
				},
				"strengths": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "3 points forts en français",
				},
				"blindspots": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "2 angles morts ou faiblesses en français",
				},
				"tips": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "3 conseils personnalisés en français",
				},
				"scoreProfile": map[string]any{
					"type":        "object",
					"description": "Scores de 0-100 pour: frugalité, impulsivité, recherche, budget, qualité",
					"properties": map[string]any{
						"frugalité":   map[string]any{"type": "integer"},
						"impulsivité": map[string]any{"type": "integer"},
						"recherche":   map[string]any{"type": "integer"},
						"budget":      map[string]any{"type": "integer"},
						"qualité":     map[string]any{"type": "integer"},
					},
				},
			},
		},
	}
}

// ShoppingPersonaAgent uses LLM tool-use to generate a global ShoppingPersona.
type ShoppingPersonaAgent struct {
	provider llm.Provider
}

// NewShoppingPersonaAgent creates a ShoppingPersonaAgent backed by the given LLM provider.
func NewShoppingPersonaAgent(provider llm.Provider) *ShoppingPersonaAgent {
	return &ShoppingPersonaAgent{provider: provider}
}

// shoppingSummary holds aggregated data gathered before calling the LLM.
type shoppingSummary struct {
	missionCount int
	totalSpent   float64
	categories   map[string]int
	missionPhases map[string]int
}

// Analyze calls the LLM with the provided summary and returns a ShoppingPersona.
// If provider is nil (e.g. tests that don't exercise LLM), it returns the default persona.
func (a *ShoppingPersonaAgent) Analyze(ctx context.Context, s shoppingSummary) (ShoppingPersona, error) {
	if a.provider == nil {
		return defaultShoppingPersona, nil
	}

	req := buildShoppingRequest(s)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "shopping_persona",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return ShoppingPersona{}, fmt.Errorf("shopping persona llm call: %w", err)
	}

	p, err := parseShoppingToolResponse(resp)
	if err != nil {
		// Retry in JSON-only mode.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "shopping_persona_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "analyze_shopping_persona")
		if err != nil {
			return ShoppingPersona{}, fmt.Errorf("shopping persona json fallback: %w", err)
		}
		p, err = parseShoppingToolResponse(resp)
		if err != nil {
			return ShoppingPersona{}, fmt.Errorf("parse shopping persona response: %w", err)
		}
	}

	return p, nil
}

// buildShoppingRequest constructs the LLM CompletionRequest from aggregated shopping data.
func buildShoppingRequest(s shoppingSummary) llm.CompletionRequest {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Nombre de missions: %d\n", s.missionCount))
	sb.WriteString(fmt.Sprintf("Dépense totale: %.2f€\n", s.totalSpent))

	if len(s.categories) > 0 {
		sb.WriteString("Catégories principales: ")
		cats := make([]string, 0, len(s.categories))
		for cat, count := range s.categories {
			cats = append(cats, fmt.Sprintf("%s (%d missions)", cat, count))
		}
		sb.WriteString(strings.Join(cats, ", "))
		sb.WriteString("\n")
	}

	if len(s.missionPhases) > 0 {
		sb.WriteString("Phases des missions: ")
		phases := make([]string, 0, len(s.missionPhases))
		for phase, count := range s.missionPhases {
			phases = append(phases, fmt.Sprintf("%s: %d", phase, count))
		}
		sb.WriteString(strings.Join(phases, ", "))
		sb.WriteString("\n")
	}

	archetypeList := strings.Join(knownArchetypes, ", ")

	prompt := fmt.Sprintf(`Vous êtes un psychologue de la consommation française. Analysez ce profil d'achats et identifiez l'archétype d'acheteur parmi les suivants: %s.

Données du profil:
%s
Utilisez l'outil analyze_shopping_persona pour retourner votre analyse avec un archétype, une description en français, des points forts, des angles morts et des conseils personnalisés.`,
		archetypeList,
		sb.String(),
	)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{shoppingPersonaTool()},
		MaxTokens: 1024,
	}
}

// parseShoppingToolResponse extracts a ShoppingPersona from the LLM tool call.
func parseShoppingToolResponse(resp llm.CompletionResponse) (ShoppingPersona, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "analyze_shopping_persona" {
			continue
		}
		return unmarshalShoppingPersona(tc.Input)
	}
	return ShoppingPersona{}, fmt.Errorf("LLM did not call analyze_shopping_persona")
}

// unmarshalShoppingPersona deserializes tool call input into a ShoppingPersona.
func unmarshalShoppingPersona(input map[string]any) (ShoppingPersona, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return ShoppingPersona{}, fmt.Errorf("re-marshal shopping persona input: %w", err)
	}

	var payload struct {
		Archetype    string         `json:"archetype"`
		Icon         string         `json:"icon"`
		Description  string         `json:"description"`
		Strengths    []string       `json:"strengths"`
		Blindspots   []string       `json:"blindspots"`
		Tips         []string       `json:"tips"`
		ScoreProfile map[string]any `json:"scoreProfile"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ShoppingPersona{}, fmt.Errorf("unmarshal shopping persona payload: %w", err)
	}

	scoreProfile := make(map[string]int, len(payload.ScoreProfile))
	for k, v := range payload.ScoreProfile {
		switch n := v.(type) {
		case float64:
			scoreProfile[k] = int(n)
		case int:
			scoreProfile[k] = n
		}
	}
	// Ensure all standard keys are present with a default of 50.
	for _, key := range []string{"frugalité", "impulsivité", "recherche", "budget", "qualité"} {
		if _, ok := scoreProfile[key]; !ok {
			scoreProfile[key] = 50
		}
	}

	p := ShoppingPersona{
		Archetype:    payload.Archetype,
		Icon:         payload.Icon,
		Description:  payload.Description,
		Strengths:    payload.Strengths,
		Blindspots:   payload.Blindspots,
		Tips:         payload.Tips,
		ScoreProfile: scoreProfile,
	}
	if p.Strengths == nil {
		p.Strengths = []string{}
	}
	if p.Blindspots == nil {
		p.Blindspots = []string{}
	}
	if p.Tips == nil {
		p.Tips = []string{}
	}
	return p, nil
}
