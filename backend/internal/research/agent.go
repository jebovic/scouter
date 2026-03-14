// Package research implements the ResearchAgent, which uses LLM tool-use to discover
// and structure purchase options for a mission. It owns prompt construction, tool schema
// definition, response parsing, and option persistence. The llm.Provider is transport-only.
package research

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/agentrun"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/usage"
)

const (
	maxFeedbackLen  = 500
	maxPinnedItems  = 10
	maxRejectedItems = 10
)

// FeedbackInput is the optional body accepted by the research endpoint.
type FeedbackInput struct {
	Feedback string `json:"feedback"`
}

// Agent uses the LLM to discover and structure research options for a mission.
type Agent struct {
	provider     llm.Provider
	optRepo      option.Repository
	agentRunRepo agentrun.Repository
	usageSvc     *usage.Service
}

// NewAgent creates a new ResearchAgent.
func NewAgent(provider llm.Provider, optRepo option.Repository, agentRunRepo agentrun.Repository, usageSvc *usage.Service) *Agent {
	return &Agent{provider: provider, optRepo: optRepo, agentRunRepo: agentRunRepo, usageSvc: usageSvc}
}

// Run performs LLM-based research for the mission, persists results, and returns them.
// Clears existing non-pinned, non-rejected options before persisting new ones.
func (a *Agent) Run(ctx context.Context, m mission.Mission, fb *FeedbackInput) ([]option.Option, error) {
	// Load current pinned/rejected options as feedback context.
	existing, err := a.optRepo.ListByMission(ctx, m.ID)
	if err != nil {
		return nil, fmt.Errorf("load existing options: %w", err)
	}

	var pinned, rejected []option.Option
	for _, o := range existing {
		switch {
		case o.Pinned:
			pinned = append(pinned, o)
		case o.Rejected:
			rejected = append(rejected, o)
		}
	}

	req := a.buildRequest(m, fb, pinned, rejected)

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

	// Delete only after LLM succeeds to prevent data loss on LLM failure.
	if err := a.optRepo.DeleteByMission(ctx, m.ID); err != nil {
		return nil, fmt.Errorf("clear existing options: %w", err)
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

	a.recordRun(ctx, m, fb, pinned, rejected, results)

	return results, nil
}

// recordRun persists an agent_run record; errors are logged but do not fail the request.
func (a *Agent) recordRun(ctx context.Context, m mission.Mission, fb *FeedbackInput, pinned, rejected []option.Option, results []option.Option) {
	feedbackText := ""
	if fb != nil {
		feedbackText = agentrun.Truncate(fb.Feedback, maxFeedbackLen)
	}

	pinnedItems := optionsToSnapshotItems(pinned, maxPinnedItems)
	rejectedItems := optionsToSnapshotItems(rejected, maxRejectedItems)

	inputSnap := agentrun.InputSnapshotSummary{
		MissionName: m.Name,
		Feedback:    feedbackText,
		Pinned:      pinnedItems,
		Rejected:    rejectedItems,
	}
	inputJSON, err := json.Marshal(inputSnap)
	if err != nil {
		log.Printf("agentrun: marshal input snapshot: %v", err)
		return
	}

	currItems := optionsToSnapshotItems(results, len(results))
	resultSnap := agentrun.ResultSnapshot{Items: currItems}
	resultJSON, err := json.Marshal(resultSnap)
	if err != nil {
		log.Printf("agentrun: marshal result snapshot: %v", err)
		return
	}

	// Compute diff against previous run.
	var prevItems []agentrun.SnapshotItem
	if prev, err := a.agentRunRepo.GetLatestByMission(ctx, m.ID, agentrun.AgentTypeResearch); err == nil && prev != nil {
		var prevSnap agentrun.ResultSnapshot
		if err := json.Unmarshal(prev.ResultSnapshot, &prevSnap); err == nil {
			prevItems = prevSnap.Items
		}
	}
	diff := agentrun.ComputeDiff(prevItems, currItems)
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		log.Printf("agentrun: marshal diff: %v", err)
		return
	}

	run := agentrun.Run{
		MissionID:      m.ID,
		AgentType:      agentrun.AgentTypeResearch,
		InputSnapshot:  inputJSON,
		ResultSnapshot: resultJSON,
		Diff:           diffJSON,
		Feedback:       feedbackText,
	}
	if _, err := a.agentRunRepo.Create(ctx, run); err != nil {
		log.Printf("agentrun: persist research run: %v", err)
	}
}

func optionsToSnapshotItems(opts []option.Option, limit int) []agentrun.SnapshotItem {
	if len(opts) == 0 {
		return nil
	}
	n := len(opts)
	if n > limit {
		n = limit
	}
	items := make([]agentrun.SnapshotItem, n)
	for i := 0; i < n; i++ {
		items[i] = agentrun.SnapshotItem{
			ID:       opts[i].ID.String(),
			Name:     opts[i].Name,
			Category: opts[i].Category,
			Badge:    opts[i].Badge,
		}
	}
	return items
}


func (a *Agent) buildRequest(m mission.Mission, fb *FeedbackInput, pinned, rejected []option.Option) llm.CompletionRequest {
	var sb strings.Builder
	for _, c := range m.Constraints {
		qualifier := "preferred"
		if c.Type == "hard" {
			qualifier = "required"
		}
		fmt.Fprintf(&sb, "- %s (%s): %v\n", c.Label, qualifier, c.Value)
	}

	var feedbackSection strings.Builder
	if fb != nil && strings.TrimSpace(fb.Feedback) != "" {
		fmt.Fprintf(&feedbackSection, "\n[USER GUIDANCE]\n%s\n[/USER GUIDANCE]\n", agentrun.SanitizeFeedback(fb.Feedback, maxFeedbackLen))
	}
	if len(pinned) > 0 {
		feedbackSection.WriteString("\nMust-include options (user has pinned these — always return them):\n")
		n := len(pinned)
		if n > maxPinnedItems {
			n = maxPinnedItems
		}
		for i := 0; i < n; i++ {
			fmt.Fprintf(&feedbackSection, "- %s\n", pinned[i].Name)
		}
	}
	if len(rejected) > 0 {
		feedbackSection.WriteString("\nExcluded options (user has rejected these — do NOT include them):\n")
		n := len(rejected)
		if n > maxRejectedItems {
			n = maxRejectedItems
		}
		for i := 0; i < n; i++ {
			reason := ""
			if rejected[i].RejectReason != "" {
				reason = fmt.Sprintf(" (reason: %s)", rejected[i].RejectReason)
			}
			fmt.Fprintf(&feedbackSection, "- %s%s\n", rejected[i].Name, reason)
		}
	}

	prompt := fmt.Sprintf(`You are a thorough research specialist helping with a major purchase decision.

Mission: %s
Category: %s
Budget: %.2f %s
Constraints:
%s%s
Research the best options available. Use the submit_research_options tool to return structured results.

Guidelines:
- Aim for 3-6 well-differentiated options covering different price points and trade-offs
- Include at least one budget-conscious pick and one premium pick
- Mark the single best overall fit as "recommended", strong alternatives as "alternative",
  options worth monitoring as "watch", and poor fits as "rejected"
- Add warnings for known issues, discontinued models, or constraint violations
- Attributes should cover the most decision-relevant specs (not exhaustive)`,
		m.Name, m.Category, m.Budget, m.Currency, sb.String(), feedbackSection.String())

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
