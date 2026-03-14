// Package research implements the ResearchAgent, which uses LLM tool-use to discover
// and structure purchase options for a mission. It owns prompt construction, tool schema
// definition, response parsing, and option persistence. The llm.Provider is transport-only.
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/usage"
)

// Agent uses the LLM to discover and structure research options for a mission.
type Agent struct {
	provider llm.Provider
	optRepo  option.Repository
	usageSvc *usage.Service
}

// NewAgent creates a new ResearchAgent.
func NewAgent(provider llm.Provider, optRepo option.Repository, usageSvc *usage.Service) *Agent {
	return &Agent{provider: provider, optRepo: optRepo, usageSvc: usageSvc}
}

// Run performs LLM-based research for the mission, persists results, and returns them.
// Clears existing options for the mission before persisting new ones (idempotent re-runs).
func (a *Agent) Run(ctx context.Context, m mission.Mission) ([]option.Option, error) {
	if err := a.optRepo.DeleteByMission(ctx, m.ID); err != nil {
		return nil, fmt.Errorf("clear existing options: %w", err)
	}

	req := a.buildRequest(m)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("research llm call: %w", err)
	}
	a.usageSvc.Log(ctx, resp.Usage, "research", m.ID, resp.WasFallback)

	rawOptions, err := parseToolResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("parse research response: %w", err)
	}

	results := make([]option.Option, 0, len(rawOptions))
	for _, o := range rawOptions {
		o.MissionID = m.ID
		created, err := a.optRepo.Create(ctx, o)
		if err != nil {
			return nil, fmt.Errorf("persist option %q: %w", o.Name, err)
		}
		results = append(results, *created)
	}

	return results, nil
}

func (a *Agent) buildRequest(m mission.Mission) llm.CompletionRequest {
	var sb strings.Builder
	for _, c := range m.Constraints {
		qualifier := "preferred"
		if c.Type == "hard" {
			qualifier = "required"
		}
		fmt.Fprintf(&sb, "- %s (%s): %v\n", c.Label, qualifier, c.Value)
	}

	prompt := fmt.Sprintf(`You are a thorough research specialist helping with a major purchase decision.

Mission: %s
Category: %s
Budget: %.2f %s
Constraints:
%s
Research the best options available. Use the submit_research_options tool to return structured results.

Guidelines:
- Aim for 3-6 well-differentiated options covering different price points and trade-offs
- Include at least one budget-conscious pick and one premium pick
- Mark the single best overall fit as "recommended", strong alternatives as "alternative",
  options worth monitoring as "watch", and poor fits as "rejected"
- Add warnings for known issues, discontinued models, or constraint violations
- Attributes should cover the most decision-relevant specs (not exhaustive)`,
		m.Name, m.Category, m.Budget, m.Currency, sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{researchTool()},
		MaxTokens: 4096,
	}
}

// parseToolResponse extracts options from the first submit_research_options tool call.
func parseToolResponse(resp llm.CompletionResponse) ([]option.Option, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_research_options" {
			continue
		}
		return unmarshalOptions(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call submit_research_options")
}

func unmarshalOptions(input map[string]any) ([]option.Option, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Options []struct {
			Name       string             `json:"name"`
			Category   string             `json:"category"`
			Badge      string             `json:"badge"`
			Attributes []option.Attribute `json:"attributes"`
			PriceRange *option.PriceRange `json:"price_range"`
			Notes      string             `json:"notes"`
			Warnings   []string           `json:"warnings"`
			URL        string             `json:"url"`
		} `json:"options"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal options payload: %w", err)
	}

	opts := make([]option.Option, 0, len(payload.Options))
	for _, p := range payload.Options {
		if p.Attributes == nil {
			p.Attributes = []option.Attribute{}
		}
		if p.Warnings == nil {
			p.Warnings = []string{}
		}
		opts = append(opts, option.Option{
			Name:       p.Name,
			Category:   p.Category,
			Badge:      p.Badge,
			Attributes: p.Attributes,
			PriceRange: p.PriceRange,
			Notes:      p.Notes,
			Warnings:   p.Warnings,
			URL:        p.URL,
		})
	}

	return opts, nil
}

// researchTool returns the tool schema for structured option submission.
func researchTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_research_options",
		Description: "Submit structured research results for the mission. Call this once with all discovered options.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"options"},
			"properties": map[string]any{
				"options": map[string]any{
					"type":        "array",
					"description": "List of researched purchase options",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"name", "category", "badge"},
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Product or option name",
							},
							"category": map[string]any{
								"type":        "string",
								"description": "Sub-category within the mission domain",
							},
							"badge": map[string]any{
								"type": "string",
								"enum": []string{"recommended", "alternative", "watch", "rejected"},
							},
							"attributes": map[string]any{
								"type":        "array",
								"description": "Key decision-relevant specifications",
								"items": map[string]any{
									"type":     "object",
									"required": []string{"key", "label", "value", "type"},
									"properties": map[string]any{
										"key":   map[string]any{"type": "string"},
										"label": map[string]any{"type": "string"},
										"value": map[string]any{},
										"type":  map[string]any{"type": "string", "enum": []string{"text", "price", "score", "boolean"}},
										"max":   map[string]any{"type": "integer", "description": "Max value for score type"},
									},
								},
							},
							"price_range": map[string]any{
								"type":        "object",
								"description": "Observed or estimated price range",
								"properties": map[string]any{
									"min":  map[string]any{"type": "number"},
									"max":  map[string]any{"type": "number"},
									"best": map[string]any{"type": "number", "description": "Best current price found"},
								},
							},
							"notes":    map[string]any{"type": "string", "description": "Summary notes about this option"},
							"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Known issues or constraint violations"},
							"url":      map[string]any{"type": "string", "description": "Primary product or review URL"},
						},
					},
				},
			},
		},
	}
}
