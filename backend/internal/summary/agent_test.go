package summary_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/summary"
)

// ── fake LLM provider ─────────────────────────────────────────────────────────

type fakeLLMProvider struct {
	resp llm.CompletionResponse
	err  error
	calls int
}

func (f *fakeLLMProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	f.calls++
	return f.resp, f.err
}

// ── fixtures ──────────────────────────────────────────────────────────────────

func testMission() mission.Mission {
	return mission.Mission{
		ID:       uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Slug:     "casque-audio",
		Name:     "Casque Audio Premium",
		Category: "electronics",
		Budget:   300.0,
		Currency: "EUR",
	}
}

func testOptions() []option.Option {
	return []option.Option{
		{
			ID:    uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Name:  "Sony WH-1000XM5",
			Badge: "recommended",
			PriceRange: &option.PriceRange{
				Min:  279.0,
				Best: 279.0,
				Max:  299.0,
			},
		},
		{
			ID:    uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Name:  "Bose QC45",
			Badge: "alternative",
			PriceRange: &option.PriceRange{
				Min:  249.0,
				Best: 249.0,
				Max:  279.0,
			},
		},
	}
}

func testItems() []shopping.Item {
	return []shopping.Item{
		{
			ID:       uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
			Name:     "Sony WH-1000XM5",
			Merchant: "Fnac",
			Price:    279.0,
			Status:   "buy",
		},
	}
}

// toolCallResp returns a response simulating the LLM invoking the summarize_mission tool.
func toolCallResp(bullets []string, verdict string) llm.CompletionResponse {
	return llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{
				Name: "summarize_mission",
				Input: map[string]any{
					"bullets": func() []any {
						out := make([]any, len(bullets))
						for i, b := range bullets {
							out[i] = b
						}
						return out
					}(),
					"verdict": verdict,
				},
			},
		},
	}
}

// ── Agent.Brief — happy path ───────────────────────────────────────────────────

func TestAgent_Brief_GoVerdict(t *testing.T) {
	bullets := []string{
		"✅ Meilleure option: Sony WH-1000XM5 à 279€",
		"⚠️ Budget: 93% utilisé",
		"🎯 Prochaine action: acheter maintenant",
	}
	provider := &fakeLLMProvider{resp: toolCallResp(bullets, "go")}
	agent := summary.NewAgent(provider)

	got, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Verdict != "go" {
		t.Errorf("expected verdict 'go', got %q", got.Verdict)
	}
	if got.MissionSlug != "casque-audio" {
		t.Errorf("expected slug 'casque-audio', got %q", got.MissionSlug)
	}
	if len(got.Bullets) != 3 {
		t.Errorf("expected 3 bullets, got %d", len(got.Bullets))
	}
	if got.CachedAt == 0 {
		t.Error("expected CachedAt to be set")
	}
}

func TestAgent_Brief_WaitVerdict(t *testing.T) {
	bullets := []string{
		"⏳ Prix en baisse attendue en décembre",
		"⚠️ Budget: 85% utilisé",
		"📉 Tendance baissière sur 30 jours",
	}
	provider := &fakeLLMProvider{resp: toolCallResp(bullets, "wait")}
	agent := summary.NewAgent(provider)

	got, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Verdict != "wait" {
		t.Errorf("expected verdict 'wait', got %q", got.Verdict)
	}
}

func TestAgent_Brief_ResearchMoreVerdict(t *testing.T) {
	bullets := []string{
		"🔍 Peu d'options comparées",
		"❓ Aucun article dans le panier",
		"📋 Lancer l'agent de recherche",
	}
	provider := &fakeLLMProvider{resp: toolCallResp(bullets, "research_more")}
	agent := summary.NewAgent(provider)

	got, err := agent.Brief(context.Background(), testMission(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Verdict != "research_more" {
		t.Errorf("expected verdict 'research_more', got %q", got.Verdict)
	}
}

func TestAgent_Brief_CallsLLMOnce(t *testing.T) {
	bullets := []string{"✅ OK", "⚠️ Budget", "🎯 Action"}
	provider := &fakeLLMProvider{resp: toolCallResp(bullets, "go")}
	agent := summary.NewAgent(provider)

	_, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.calls != 1 {
		t.Errorf("expected LLM called exactly once, got %d", provider.calls)
	}
}

// ── Agent.Brief — error paths ─────────────────────────────────────────────────

func TestAgent_Brief_LLMError(t *testing.T) {
	provider := &fakeLLMProvider{err: errors.New("network timeout")}
	agent := summary.NewAgent(provider)

	_, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

func TestAgent_Brief_NoToolCall_NoContent_ReturnsError(t *testing.T) {
	// LLM returns neither tool call nor content.
	provider := &fakeLLMProvider{resp: llm.CompletionResponse{Content: ""}}
	agent := summary.NewAgent(provider)

	_, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err == nil {
		t.Fatal("expected error when LLM returns no tool call and no content")
	}
}

func TestAgent_Brief_TextFallback_ValidJSON(t *testing.T) {
	// LLM returns JSON in text content instead of tool call.
	jsonContent := `{"bullets": ["✅ OK", "⚠️ Budget OK", "🎯 Achetez maintenant"], "verdict": "go"}`
	provider := &fakeLLMProvider{resp: llm.CompletionResponse{Content: jsonContent}}
	agent := summary.NewAgent(provider)

	got, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error on text fallback: %v", err)
	}
	if got.Verdict != "go" {
		t.Errorf("expected verdict 'go', got %q", got.Verdict)
	}
	if len(got.Bullets) != 3 {
		t.Errorf("expected 3 bullets, got %d", len(got.Bullets))
	}
}

// ── Agent.Brief — tool request shape ─────────────────────────────────────────

func TestAgent_Brief_RequestIncludesTools(t *testing.T) {
	var captured llm.CompletionRequest
	capProv := &capturingProvider{
		resp:    toolCallResp([]string{"✅ OK", "⚠️ Budget", "🎯 Action"}, "go"),
		capture: &captured,
	}
	agent := summary.NewAgent(capProv)

	_, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.Tools) == 0 {
		t.Error("expected at least one tool in request")
	}
	if captured.Tools[0].Name != "summarize_mission" {
		t.Errorf("expected tool name 'summarize_mission', got %q", captured.Tools[0].Name)
	}
}

func TestAgent_Brief_MaxTokensReasonable(t *testing.T) {
	var captured llm.CompletionRequest
	capProv := &capturingProvider{
		resp:    toolCallResp([]string{"✅ OK", "⚠️ Budget", "🎯 Action"}, "go"),
		capture: &captured,
	}
	agent := summary.NewAgent(capProv)

	_, err := agent.Brief(context.Background(), testMission(), testOptions(), testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.MaxTokens <= 0 {
		t.Errorf("expected MaxTokens > 0, got %d", captured.MaxTokens)
	}
	if captured.MaxTokens > 1200 {
		t.Errorf("expected MaxTokens <= 1200, got %d", captured.MaxTokens)
	}
}

// capturingProvider captures the last CompletionRequest for inspection.
type capturingProvider struct {
	resp    llm.CompletionResponse
	capture *llm.CompletionRequest
}

func (f *capturingProvider) Complete(_ context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	*f.capture = req
	return f.resp, nil
}
