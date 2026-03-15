package coach

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
)

// MockProvider returns a controlled response for testing.
type MockProvider struct {
	response *llm.CompletionResponse
	err      error
}

func (m *MockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	if m.err != nil {
		return llm.CompletionResponse{}, m.err
	}
	if m.response == nil {
		return llm.CompletionResponse{}, fmt.Errorf("no response configured")
	}
	return *m.response, nil
}

func TestAgent_BuildRequest(t *testing.T) {
	agent := &Agent{
		provider: &MockProvider{},
	}

	m := mission.Mission{
		ID:       uuid.New(),
		Name:     "Gaming Laptop",
		Category: "electronics",
		Budget:   1500,
		Currency: "USD",
		Phase:    "researching",
	}

	req := agent.buildRequest(m, 3, 5)

	if len(req.Messages) == 0 {
		t.Error("buildRequest should include at least one message")
	}
	if len(req.Messages[0].Content) == 0 {
		t.Error("message content should not be empty")
	}
	if req.Tools == nil || len(req.Tools) == 0 {
		t.Error("buildRequest should include tools")
	}
	if req.Tools[0].Name != "submit_tips" {
		t.Errorf("tool name = %q; want submit_tips", req.Tools[0].Name)
	}
}

func TestParseToolResponse_ValidCall(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{
				Name: "submit_tips",
				Input: map[string]any{
					"tips": []any{
						map[string]any{
							"category": "research",
							"priority": "high",
							"title":    "Consider brand reputation",
							"body":     "Brands like Lenovo and Dell have strong track records...",
						},
						map[string]any{
							"category": "budget",
							"priority": "medium",
							"title":    "Watch for sales events",
							"body":     "Prime Day and Black Friday can save you...",
						},
					},
				},
			},
		},
	}

	tips, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse() error: %v", err)
	}
	if len(tips) != 2 {
		t.Errorf("len(tips) = %d; want 2", len(tips))
	}
	if tips[0].Category != "research" {
		t.Errorf("Category = %q; want research", tips[0].Category)
	}
	if tips[0].Priority != "high" {
		t.Errorf("Priority = %q; want high", tips[0].Priority)
	}
}

func TestParseToolResponse_NoToolCall(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "wrong_tool", Input: map[string]any{}},
		},
	}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("parseToolResponse with wrong tool name should return error")
	}
}

func TestParseToolResponse_EmptyToolCalls(t *testing.T) {
	resp := llm.CompletionResponse{ToolCalls: nil}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("parseToolResponse with no tool calls should return error")
	}
}

func TestUnmarshalTips_Valid(t *testing.T) {
	input := map[string]any{
		"tips": []any{
			map[string]any{
				"category": "research",
				"priority": "high",
				"title":    "Expand your options",
				"body":     "You only have 2 options so far...",
				"action":   "Launch research",
			},
		},
	}

	tips, err := unmarshalTips(input)
	if err != nil {
		t.Fatalf("unmarshalTips() error: %v", err)
	}
	if len(tips) != 1 {
		t.Errorf("len(tips) = %d; want 1", len(tips))
	}
	if tips[0].Title != "Expand your options" {
		t.Errorf("Title = %q; want 'Expand your options'", tips[0].Title)
	}
	if tips[0].Action != "Launch research" {
		t.Errorf("Action = %q; want 'Launch research'", tips[0].Action)
	}
}

func TestUnmarshalTips_NoAction(t *testing.T) {
	input := map[string]any{
		"tips": []any{
			map[string]any{
				"category": "timing",
				"priority": "low",
				"title":    "Good timing",
				"body":     "Wait 2 weeks for new models...",
				// action is optional
			},
		},
	}

	tips, err := unmarshalTips(input)
	if err != nil {
		t.Fatalf("unmarshalTips() error: %v", err)
	}
	if len(tips) != 1 {
		t.Fatalf("expected 1 tip, got %d", len(tips))
	}
	if tips[0].Action != "" {
		t.Errorf("Action = %q; want empty", tips[0].Action)
	}
}

func TestUnmarshalTips_InvalidStructure(t *testing.T) {
	input := map[string]any{
		"tips": "not-an-array",
	}

	_, err := unmarshalTips(input)
	if err == nil {
		t.Error("unmarshalTips with invalid structure should return error")
	}
}

func TestUnmarshalTips_EmptyTips(t *testing.T) {
	input := map[string]any{
		"tips": []any{},
	}

	tips, err := unmarshalTips(input)
	if err != nil {
		t.Fatalf("unmarshalTips() error: %v", err)
	}
	if len(tips) != 0 {
		t.Errorf("len(tips) = %d; want 0", len(tips))
	}
}

func TestGenerateTipID(t *testing.T) {
	missionID := uuid.New()

	// Same input should generate same ID
	id1 := generateTipID(missionID, "tip content")
	id2 := generateTipID(missionID, "tip content")
	if id1 != id2 {
		t.Errorf("generateTipID should be deterministic; got %q and %q", id1, id2)
	}

	// Different content should generate different ID
	id3 := generateTipID(missionID, "different content")
	if id1 == id3 {
		t.Error("generateTipID with different content should generate different ID")
	}
}

func TestAgent_GetAnalysisContext(t *testing.T) {
	missionID := uuid.New()
	m := mission.Mission{
		ID:       missionID,
		Name:     "Laptop Purchase",
		Category: "electronics",
		Budget:   2000,
		Currency: "EUR",
		Phase:    "options",
	}

	tests := []struct {
		name          string
		optionsCount  int
		itemsCount    int
		wantOptional  bool
		wantRecommend bool
	}{
		{
			name:          "few options, no items",
			optionsCount:  1,
			itemsCount:    0,
			wantOptional:  true,  // should recommend research
			wantRecommend: false, // cannot recommend without pricing
		},
		{
			name:          "many options, many items",
			optionsCount:  5,
			itemsCount:    10,
			wantOptional:  false, // has enough options
			wantRecommend: true,  // has pricing data
		},
		{
			name:          "moderate coverage",
			optionsCount:  3,
			itemsCount:    2,
			wantOptional:  false,
			wantRecommend: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// Test that buildRequest can accept these values
			agent := &Agent{provider: &MockProvider{}}
			req := agent.buildRequest(m, tt.optionsCount, tt.itemsCount)

			// Verify request is valid
			if len(req.Messages) == 0 {
				t.Error("request should have messages")
			}
			if len(req.Tools) == 0 {
				t.Error("request should have tools")
			}
		})
	}
}
