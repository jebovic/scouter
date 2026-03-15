package coach

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
)

// Agent uses the LLM to generate actionable tips for a mission.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new CoachAgent.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Run analyzes a mission and returns coaching tips.
func (a *Agent) Run(ctx context.Context, m mission.Mission, optionsCount, itemsCount int) ([]CoachTip, error) {
	req := a.buildRequest(m, optionsCount, itemsCount)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "coach",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("coach llm call: %w", err)
	}

	rawTips, err := parseToolResponse(resp)
	if err != nil {
		// Retry as JSON-only mode
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapText,
			Label:        "coach_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "submit_tips")
		if err != nil {
			return nil, fmt.Errorf("coach json fallback: %w", err)
		}
		rawTips, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse coach response: %w", err)
		}
	}

	// Generate deterministic IDs and return
	tips := make([]CoachTip, 0, len(rawTips))
	for _, t := range rawTips {
		id := generateTipID(m.ID, t.Title+t.Body)
		t.ID = id
		tips = append(tips, t)
	}

	return tips, nil
}

// buildRequest constructs the LLM prompt and tool schema for coaching analysis.
func (a *Agent) buildRequest(m mission.Mission, optionsCount, itemsCount int) llm.CompletionRequest {
	prompt := fmt.Sprintf(`You are an expert shopping advisor helping someone with a major purchase decision.

Mission: %s
Category: %s
Budget: %.2f %s
Current Phase: %s
Options Discovered: %d
Shopping Items: %d

Analyze the current state of this mission and provide 2-4 actionable, specific tips to guide the user forward.
Use the submit_tips tool to return structured coaching suggestions.

Guidelines:
- Categories: "research" (explore options), "budget" (cost-related), "timing" (when to buy), "decision" (choosing between options)
- Priority: "high" (urgent/critical), "medium" (important), "low" (nice-to-have)
- Titles should be short (5-8 words)
- Body should be concrete and actionable (2-3 sentences)
- Action is optional but helpful when you suggest a next step (e.g., "Launch research", "Set price alerts")
- Avoid generic advice — focus on this specific mission`,
		m.Name, m.Category, m.Budget, m.Currency, m.Phase, optionsCount, itemsCount)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{coachTool()},
		MaxTokens: 1024,
	}
}

// parseToolResponse extracts tips from the first submit_tips tool call.
func parseToolResponse(resp llm.CompletionResponse) ([]CoachTip, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_tips" {
			continue
		}
		return unmarshalTips(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call submit_tips")
}

// unmarshalTips parses the tool input into CoachTip slice.
func unmarshalTips(input map[string]any) ([]CoachTip, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Tips []struct {
			Category string `json:"category"`
			Priority string `json:"priority"`
			Title    string `json:"title"`
			Body     string `json:"body"`
			Action   string `json:"action"`
		} `json:"tips"`
	}

	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal tips payload: %w", err)
	}

	tips := make([]CoachTip, 0, len(payload.Tips))
	for _, p := range payload.Tips {
		tips = append(tips, CoachTip{
			Category: p.Category,
			Priority: p.Priority,
			Title:    p.Title,
			Body:     p.Body,
			Action:   p.Action,
		})
	}

	return tips, nil
}

// generateTipID creates a deterministic short ID from mission+content.
func generateTipID(missionID uuid.UUID, content string) string {
	h := sha256.Sum256([]byte(missionID.String() + content))
	return hex.EncodeToString(h[:])[:8]
}

// coachTool returns the tool schema for structured tip submission.
func coachTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_tips",
		Description: "Submit coaching tips for the mission. Call this once with all tips.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"tips"},
			"properties": map[string]any{
				"tips": map[string]any{
					"type":        "array",
					"description": "List of coaching tips",
					"items": map[string]any{
						"type":     "object",
						"required": []string{"category", "priority", "title", "body"},
						"properties": map[string]any{
							"category": map[string]any{
								"type":        "string",
								"enum":        []string{"research", "budget", "timing", "decision"},
								"description": "Type of advice",
							},
							"priority": map[string]any{
								"type":        "string",
								"enum":        []string{"high", "medium", "low"},
								"description": "Urgency level",
							},
							"title": map[string]any{
								"type":        "string",
								"description": "Short title (5-8 words)",
							},
							"body": map[string]any{
								"type":        "string",
								"description": "Detailed advice (2-3 sentences)",
							},
							"action": map[string]any{
								"type":        "string",
								"description": "Optional call-to-action (e.g., 'Launch research')",
							},
						},
					},
				},
			},
		},
	}
}
