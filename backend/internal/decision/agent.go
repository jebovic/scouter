package decision

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/usage"
)

// Agent wraps the LLM to generate a natural-language recommendation from scored options.
type Agent struct {
	provider llm.Provider
	usageSvc *usage.Service
}

// NewAgent creates a new DecisionAgent.
func NewAgent(provider llm.Provider, usageSvc *usage.Service) *Agent {
	return &Agent{provider: provider, usageSvc: usageSvc}
}

// Summarize asks the LLM to produce a decision summary given the scored options.
func (a *Agent) Summarize(ctx context.Context, m mission.Mission, scores []ScoreResult) (string, error) {
	req := a.buildRequest(m, scores)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "decision",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return "", fmt.Errorf("decision llm call: %w", err)
	}
	a.usageSvc.Log(ctx, resp.Usage, "decision", m.ID, resp.WasFallback)

	summary, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh timeout.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "decision_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "submit_decision_summary")
		if err != nil {
			return "", fmt.Errorf("decision json fallback: %w", err)
		}
		a.usageSvc.Log(ctx, resp.Usage, "decision_fallback", m.ID, resp.WasFallback)
		summary, err = parseToolResponse(resp)
		if err != nil {
			return "", fmt.Errorf("parse decision response: %w", err)
		}
	}
	return summary, nil
}

func (a *Agent) buildRequest(m mission.Mission, scores []ScoreResult) llm.CompletionRequest {
	// Sort by score descending for the prompt
	sorted := make([]ScoreResult, len(scores))
	copy(sorted, scores)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	var sb strings.Builder
	for i, s := range sorted {
		if s.Eliminated {
			fmt.Fprintf(&sb, "%d. %s — ELIMINATED (%s)\n", i+1, s.OptionName, s.EliminReason)
		} else {
			fmt.Fprintf(&sb, "%d. %s — Score: %.0f/100 (price %.0f, quality %.0f, features %.0f)\n",
				i+1, s.OptionName, s.Score, s.PriceScore, s.QualScore, s.FeatScore)
		}
	}

	prompt := fmt.Sprintf(`You are a purchasing advisor helping make a final buying decision.

Mission: %s
Budget: %.2f %s
Category: %s

Option scores (composite 0-100, higher is better):
%s
Using the scoring data above and your knowledge of the products, provide a clear, actionable purchase recommendation. Explain:
1. Which option is the top pick and why
2. Any important trade-offs to be aware of
3. Whether the budget is well-matched to the options
4. A final one-sentence action recommendation

Use the submit_decision_summary tool to return your recommendation.`,
		m.Name, m.Budget, m.Currency, m.Category, sb.String())

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{decisionTool()},
		MaxTokens: 1024,
	}
}

func parseToolResponse(resp llm.CompletionResponse) (string, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_decision_summary" {
			continue
		}
		return unmarshalSummary(tc.Input)
	}
	// Fallback: use text content if tool was not called
	if resp.Content != "" {
		return resp.Content, nil
	}
	return "", fmt.Errorf("LLM did not call submit_decision_summary")
}

func unmarshalSummary(input map[string]any) (string, error) {
	s, ok := input["summary"].(string)
	if !ok {
		return "", fmt.Errorf("summary field missing or not a string in tool response")
	}
	return s, nil
}

func decisionTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_decision_summary",
		Description: "Submit the final decision recommendation. Call once with the complete summary.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"summary"},
			"properties": map[string]any{
				"summary": map[string]any{
					"type":        "string",
					"description": "Clear, actionable purchase recommendation covering top pick, trade-offs, budget fit, and final action.",
				},
			},
		},
	}
}
