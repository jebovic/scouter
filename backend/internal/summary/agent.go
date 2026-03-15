package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// Briefer is the narrow interface consumed by Handler so tests can inject
// a fake without depending on the real LLM.
type Briefer interface {
	Brief(ctx context.Context, m mission.Mission, opts []option.Option, items []shopping.Item) (*MissionSummary, error)
}

// Agent uses the LLM to generate a structured French executive brief via tool-use.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new summary Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// summarizeTool is the tool schema for structured LLM output.
var summarizeTool = llm.Tool{
	Name:        "summarize_mission",
	Description: "Génère un bilan exécutif structuré de la mission d'achat en français.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bullets": map[string]any{
				"type":        "array",
				"description": "3 à 5 points clés en français avec emojis décrivant l'état de la mission",
				"items":       map[string]any{"type": "string"},
				"minItems":    3,
				"maxItems":    5,
			},
			"verdict": map[string]any{
				"type":        "string",
				"description": "Recommandation finale : go (achetez maintenant), wait (attendez), research_more (plus d'infos nécessaires)",
				"enum":        []string{"go", "wait", "research_more"},
			},
		},
		"required": []string{"bullets", "verdict"},
	},
}

// Brief calls the LLM and returns a structured MissionSummary.
func (a *Agent) Brief(ctx context.Context, m mission.Mission, opts []option.Option, items []shopping.Item) (*MissionSummary, error) {
	prompt := buildBriefPrompt(m, opts, items)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText | llm.CapToolUse,
		Label:        "mission-brief",
	})

	req := llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{summarizeTool},
		MaxTokens: 800,
	}

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("summary agent llm call: %w", err)
	}

	return parseToolResponse(m.Slug, resp)
}

// buildBriefPrompt constructs the user prompt sent to the LLM.
func buildBriefPrompt(m mission.Mission, opts []option.Option, items []shopping.Item) string {
	var sb strings.Builder

	fmt.Fprintf(&sb,
		"Vous êtes un assistant d'achat personnel. Résumez en 3-5 points clés (en français, avec emojis) l'état de la mission '%s' (objectif: %.2f %s).\n\n",
		m.Name, m.Budget, m.Currency,
	)

	// Top options (up to 5)
	if len(opts) > 0 {
		sb.WriteString("Options disponibles:\n")
		limit := len(opts)
		if limit > 5 {
			limit = 5
		}
		for _, o := range opts[:limit] {
			priceStr := ""
			if o.PriceRange != nil {
				priceStr = fmt.Sprintf(" — %.2f %s", o.PriceRange.Best, m.Currency)
			}
			fmt.Fprintf(&sb, "- %s (%s)%s\n", o.Name, o.Badge, priceStr)
		}
		sb.WriteString("\n")
	}

	// Shopping items (up to 5)
	if len(items) > 0 {
		sb.WriteString("Articles en panier:\n")
		limit := len(items)
		if limit > 5 {
			limit = 5
		}
		for _, i := range items[:limit] {
			fmt.Fprintf(&sb, "- %s chez %s : %.2f %s (%s)\n", i.Name, i.Merchant, i.Price, m.Currency, i.Status)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Donnez un verdict: 'go' (achetez maintenant), 'wait' (attendez une opportunité), ou 'research_more' (besoin de plus d'infos).\n")
	sb.WriteString("Appelez l'outil summarize_mission avec vos conclusions.\n")

	return sb.String()
}

// parseToolResponse extracts the structured result from the LLM tool call.
func parseToolResponse(slug string, resp llm.CompletionResponse) (*MissionSummary, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "summarize_mission" {
			continue
		}
		return parseToolInput(slug, tc.Input)
	}

	// Fallback: try to parse JSON from text content if no tool call was made.
	if resp.Content != "" {
		return parseTextFallback(slug, resp.Content)
	}

	return nil, fmt.Errorf("summary agent: LLM did not call summarize_mission tool")
}

// parseToolInput converts the raw tool input map into a MissionSummary.
func parseToolInput(slug string, input map[string]any) (*MissionSummary, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("summary agent: marshal tool input: %w", err)
	}

	var parsed struct {
		Bullets []string `json:"bullets"`
		Verdict string   `json:"verdict"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("summary agent: unmarshal tool input: %w", err)
	}

	if len(parsed.Bullets) == 0 {
		return nil, fmt.Errorf("summary agent: no bullets in tool response")
	}
	if !isValidVerdict(parsed.Verdict) {
		return nil, fmt.Errorf("summary agent: invalid verdict %q", parsed.Verdict)
	}

	return &MissionSummary{
		MissionSlug: slug,
		Bullets:     parsed.Bullets,
		Verdict:     parsed.Verdict,
		CachedAt:    time.Now().Unix(),
	}, nil
}

// parseTextFallback attempts to extract JSON from a text response when the LLM
// does not invoke the tool (e.g. some Ollama models return JSON in the content).
func parseTextFallback(slug, content string) (*MissionSummary, error) {
	content = strings.TrimSpace(content)
	// Find outermost JSON object.
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("summary agent: no JSON found in text response")
	}
	jsonStr := content[start : end+1]

	var parsed struct {
		Bullets []string `json:"bullets"`
		Verdict string   `json:"verdict"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("summary agent: parse text fallback: %w", err)
	}
	if len(parsed.Bullets) == 0 {
		return nil, fmt.Errorf("summary agent: no bullets in text fallback")
	}
	if !isValidVerdict(parsed.Verdict) {
		return nil, fmt.Errorf("summary agent: invalid verdict in text fallback: %q", parsed.Verdict)
	}

	return &MissionSummary{
		MissionSlug: slug,
		Bullets:     parsed.Bullets,
		Verdict:     parsed.Verdict,
		CachedAt:    time.Now().Unix(),
	}, nil
}

func isValidVerdict(v string) bool {
	return v == "go" || v == "wait" || v == "research_more"
}
