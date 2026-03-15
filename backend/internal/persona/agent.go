package persona

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/llm"
)

// defaultPersona is returned when no purchase history exists.
var defaultPersona = Persona{
	Archetype: "New Scouter",
	Summary:   "Start tracking purchases to unlock your spending persona.",
	Traits:    []string{"Early adopter"},
	Tips:      []string{"Complete your first mission to get insights"},
}

// categoryRow holds aggregated stats per spending category.
type categoryRow struct {
	Category     string
	Count        int
	TotalSpent   float64
	AvgSatisfaction float64
}

// PersonaAgent uses the LLM to analyze purchase history and generate a spending persona.
type PersonaAgent struct {
	provider llm.Provider
	repo     Repository
	pool     *pgxpool.Pool
}

// NewPersonaAgent creates a new PersonaAgent.
func NewPersonaAgent(provider llm.Provider, repo Repository, pool *pgxpool.Pool) *PersonaAgent {
	return &PersonaAgent{
		provider: provider,
		repo:     repo,
		pool:     pool,
	}
}

// Run fetches purchase stats, calls the LLM, persists the result, and returns it.
// If no purchase history exists, returns and persists a default persona.
func (a *PersonaAgent) Run(ctx context.Context) (Persona, error) {
	totalCount, totalSpent, avgSatisfaction, err := a.fetchOverallStats(ctx)
	if err != nil {
		return Persona{}, fmt.Errorf("fetch overall stats: %w", err)
	}

	// No purchase history — persist and return the default persona so callers
	// always receive a saved record they can retrieve via GET.
	if totalCount == 0 {
		return a.repo.Save(ctx, defaultPersona)
	}

	categories, err := a.fetchCategoryStats(ctx)
	if err != nil {
		return Persona{}, fmt.Errorf("fetch category stats: %w", err)
	}

	req := buildRequest(totalCount, totalSpent, avgSatisfaction, categories)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "persona",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return Persona{}, fmt.Errorf("persona llm call: %w", err)
	}

	partial, err := parseToolResponse(resp)
	if err != nil {
		// Retry in JSON-only mode with a fresh timeout.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "persona_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "analyze_spending_persona")
		if err != nil {
			return Persona{}, fmt.Errorf("persona json fallback: %w", err)
		}
		partial, err = parseToolResponse(resp)
		if err != nil {
			return Persona{}, fmt.Errorf("parse persona response: %w", err)
		}
	}

	return a.repo.Save(ctx, partial)
}

// fetchOverallStats returns aggregate totals across all purchase records.
func (a *PersonaAgent) fetchOverallStats(ctx context.Context) (count int, totalSpent, avgSatisfaction float64, err error) {
	row := a.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(final_price),0), COALESCE(AVG(satisfaction),0)
		FROM purchase_records`)
	err = row.Scan(&count, &totalSpent, &avgSatisfaction)
	if err != nil {
		err = fmt.Errorf("query overall stats: %w", err)
	}
	return
}

// fetchCategoryStats returns per-category spending aggregates ordered by purchase count.
func (a *PersonaAgent) fetchCategoryStats(ctx context.Context) ([]categoryRow, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT m.category, COUNT(*), SUM(pr.final_price), AVG(pr.satisfaction)
		FROM purchase_records pr
		JOIN missions m ON m.id = pr.mission_id
		GROUP BY m.category
		ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("query category stats: %w", err)
	}
	defer rows.Close()

	var cats []categoryRow
	for rows.Next() {
		var c categoryRow
		if err := rows.Scan(&c.Category, &c.Count, &c.TotalSpent, &c.AvgSatisfaction); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func buildRequest(totalCount int, totalSpent, avgSatisfaction float64, categories []categoryRow) llm.CompletionRequest {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total purchases: %d\n", totalCount))
	sb.WriteString(fmt.Sprintf("Total spent: %.2f\n", totalSpent))
	sb.WriteString(fmt.Sprintf("Average satisfaction (1-5): %.2f\n", avgSatisfaction))

	if len(categories) > 0 {
		sb.WriteString("\nSpending by category (ordered by frequency):\n")
		for _, c := range categories {
			sb.WriteString(fmt.Sprintf("  - %s: %d purchases, total %.2f, avg satisfaction %.2f\n",
				c.Category, c.Count, c.TotalSpent, c.AvgSatisfaction))
		}
	}

	prompt := fmt.Sprintf(`You are a consumer behavior expert. Analyze the following spending patterns and generate an insightful buyer persona.

%s
Use the analyze_spending_persona tool to return your analysis with an archetype name, 2-3 sentence summary, 3-5 characteristic traits, and 3-5 personalized spending tips.`,
		sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{personaTool()},
		MaxTokens: 1024,
	}
}

// personaTool returns the tool schema for the spending persona analysis.
func personaTool() llm.Tool {
	return llm.Tool{
		Name:        "analyze_spending_persona",
		Description: "Analyze spending patterns and generate a buyer persona",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"archetype", "summary", "traits", "tips"},
			"properties": map[string]any{
				"archetype": map[string]any{
					"type":        "string",
					"description": "Short name like 'Deal Hunter', 'Quality Seeker', 'Methodical Researcher'",
				},
				"summary": map[string]any{
					"type":        "string",
					"description": "2-3 sentence persona description",
				},
				"traits": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "3-5 characteristic traits",
				},
				"tips": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "3-5 personalized spending tips",
				},
			},
		},
	}
}

// parseToolResponse extracts a Persona from the first analyze_spending_persona tool call.
func parseToolResponse(resp llm.CompletionResponse) (Persona, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "analyze_spending_persona" {
			continue
		}
		return unmarshalPersona(tc.Input)
	}
	return Persona{}, fmt.Errorf("LLM did not call analyze_spending_persona")
}

func unmarshalPersona(input map[string]any) (Persona, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return Persona{}, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Archetype string   `json:"archetype"`
		Summary   string   `json:"summary"`
		Traits    []string `json:"traits"`
		Tips      []string `json:"tips"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return Persona{}, fmt.Errorf("unmarshal persona payload: %w", err)
	}

	p := Persona{
		Archetype: payload.Archetype,
		Summary:   payload.Summary,
		Traits:    payload.Traits,
		Tips:      payload.Tips,
	}
	if p.Traits == nil {
		p.Traits = []string{}
	}
	if p.Tips == nil {
		p.Tips = []string{}
	}

	return p, nil
}
