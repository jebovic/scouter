package timing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Advisor is the narrow interface consumed by Handler so tests can inject
// a fake without depending on the real LLM.
type Advisor interface {
	Advise(ctx context.Context, missionName, category string, budgetEur float64, createdAt time.Time) (*TimingAdvice, error)
}

// Agent uses the LLM to generate purchase timing advice for a mission.
// It follows the same tool-use pattern as reviews.Agent.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new TimingAgent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Advise calls the LLM to produce a TimingAdvice for the given mission context.
// It uses tool_use (submit_timing_advice) and falls back to JSON-only mode
// when the model does not emit a tool call.
func (a *Agent) Advise(ctx context.Context, missionName, category string, budgetEur float64, createdAt time.Time) (*TimingAdvice, error) {
	req := a.buildRequest(missionName, category, budgetEur, createdAt)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapToolUse,
		Label:        "timing",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return nil, fmt.Errorf("timing llm call: %w", err)
	}

	advice, err := parseToolResponse(resp)
	if err != nil {
		// Tool call missing — retry in JSON-only mode with a fresh context.
		retryCtx := llm.WithRequestOpts(ctx, llm.RequestOpts{
			Capabilities: llm.CapToolUse,
			Label:        "timing_fallback",
		})
		resp, err = llm.RetryAsJSON(retryCtx, a.provider, req, "submit_timing_advice")
		if err != nil {
			return nil, fmt.Errorf("timing json fallback: %w", err)
		}
		advice, err = parseToolResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("parse timing response: %w", err)
		}
	}

	return advice, nil
}

func (a *Agent) buildRequest(missionName, category string, budgetEur float64, createdAt time.Time) llm.CompletionRequest {
	agedays := int(time.Since(createdAt).Hours() / 24)
	prompt := fmt.Sprintf(`You are a French consumer market timing advisor.

Mission: %s
Category: %s
Budget: %.2f EUR
Mission age: %d days

Provide purchase timing advice using the submit_timing_advice tool.

French seasonal sale patterns to consider:
- Soldes d'hiver: mid-January to mid-February (significant discounts 20-50%%)
- Soldes d'été: late June to late July (significant discounts 20-50%%)
- Black Friday: last Friday of November (electronics, computing, fashion)
- Tax refund season / March promotions: March (retailers target tax refund recipients)
- Back to school: August-September (electronics, computing, stationery)

Category-specific guidance:
- electronics: depreciate quickly, newer models announced Feb/Sep; wait for seasonal sales unless urgent
- computing: price drops around back-to-school (Aug-Sep) and Black Friday; new models often announced in Spring
- travel: prices fluctuate with demand; book early for summer, last-minute for low season
- renovation: prices stable; autumn promotions Sep-Oct common at large retailers
- custom: general seasonal patterns apply

Rules:
- recommendation must be one of: BUY_NOW, WAIT, UNSURE
- confidence is a float 0..1 (how confident you are in the recommendation)
- waitUntil is an ISO date (YYYY-MM-DD) only when recommendation is WAIT
- nextSaleEvent is a human-readable label for the next relevant French sale event (e.g. "Soldes d'été 2026")
- rationale must be 1-3 concise sentences explaining the recommendation`, missionName, category, budgetEur, agedays)

	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{timingTool()},
		MaxTokens: 512,
	}
}

// parseToolResponse extracts a TimingAdvice from the first submit_timing_advice tool call.
func parseToolResponse(resp llm.CompletionResponse) (*TimingAdvice, error) {
	for _, tc := range resp.ToolCalls {
		if tc.Name != "submit_timing_advice" {
			continue
		}
		return unmarshalAdvice(tc.Input)
	}
	return nil, fmt.Errorf("LLM did not call submit_timing_advice")
}

func unmarshalAdvice(input map[string]any) (*TimingAdvice, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("re-marshal tool input: %w", err)
	}

	var payload struct {
		Recommendation string  `json:"recommendation"`
		Rationale      string  `json:"rationale"`
		WaitUntil      string  `json:"wait_until"`
		NextSaleEvent  string  `json:"next_sale_event"`
		Confidence     float64 `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal timing advice payload: %w", err)
	}

	return &TimingAdvice{
		Recommendation: payload.Recommendation,
		Rationale:      payload.Rationale,
		WaitUntil:      payload.WaitUntil,
		NextSaleEvent:  payload.NextSaleEvent,
		Confidence:     payload.Confidence,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// timingTool returns the tool schema for structured timing advice submission.
func timingTool() llm.Tool {
	return llm.Tool{
		Name:        "submit_timing_advice",
		Description: "Submit structured purchase timing advice for the mission. Call this once with all data.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"recommendation", "rationale", "confidence"},
			"properties": map[string]any{
				"recommendation": map[string]any{
					"type":        "string",
					"description": "Whether to buy now, wait, or unsure",
					"enum":        []string{"BUY_NOW", "WAIT", "UNSURE"},
				},
				"rationale": map[string]any{
					"type":        "string",
					"description": "1-3 sentence rationale for the recommendation",
				},
				"wait_until": map[string]any{
					"type":        "string",
					"description": "ISO date (YYYY-MM-DD) indicating when to buy if recommendation is WAIT. Omit if BUY_NOW or UNSURE.",
				},
				"next_sale_event": map[string]any{
					"type":        "string",
					"description": "Human-readable label of the next relevant French sale event (e.g. 'Soldes d'été 2026')",
				},
				"confidence": map[string]any{
					"type":        "number",
					"description": "Confidence score 0.0 (very uncertain) to 1.0 (very confident)",
					"minimum":     0.0,
					"maximum":     1.0,
				},
			},
		},
	}
}
