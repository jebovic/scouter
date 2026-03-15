package substitute_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/substitute"
)

// ── fake LLM provider ────────────────────────────────────────────────────────

type fakeProvider struct {
	resp llm.CompletionResponse
	err  error
}

func (f *fakeProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return f.resp, f.err
}

func substituteToolCall(subs []substitute.Substitute) llm.CompletionResponse {
	raw, _ := json.Marshal(map[string]any{
		"substitutes": subs,
	})
	var input map[string]any
	_ = json.Unmarshal(raw, &input)
	return llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "suggest_substitutes", Input: input},
		},
	}
}

func sampleSubstitutes() []substitute.Substitute {
	return []substitute.Substitute{
		{
			Name:      "Samsung Galaxy S23",
			Brand:     "Samsung",
			Price:     799.0,
			Currency:  "EUR",
			Retailer:  "Fnac",
			Reason:    "Comparable specs at a lower price",
			Advantage: "cheaper",
			URL:       "https://fnac.com/samsung-s23",
		},
		{
			Name:      "Google Pixel 8",
			Brand:     "Google",
			Price:     650.0,
			Currency:  "EUR",
			Retailer:  "Darty",
			Reason:    "Better camera, pure Android experience",
			Advantage: "better_rated",
		},
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestAgent_FindSubstitutes_HappyPath(t *testing.T) {
	subs := sampleSubstitutes()
	provider := &fakeProvider{resp: substituteToolCall(subs)}
	agent := substitute.NewAgent(provider)

	result, err := agent.FindSubstitutes(context.Background(), "iPhone 15 Pro", "smartphone", 1000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 substitutes, got %d", len(result))
	}
	if result[0].Name != "Samsung Galaxy S23" {
		t.Errorf("expected first substitute to be Samsung Galaxy S23, got %s", result[0].Name)
	}
	if result[1].Advantage != "better_rated" {
		t.Errorf("expected second advantage to be better_rated, got %s", result[1].Advantage)
	}
}

func TestAgent_FindSubstitutes_ProviderError(t *testing.T) {
	provider := &fakeProvider{err: errors.New("llm unavailable")}
	agent := substitute.NewAgent(provider)

	_, err := agent.FindSubstitutes(context.Background(), "iPhone 15 Pro", "smartphone", 1000.0)
	if err == nil {
		t.Fatal("expected error when provider fails, got nil")
	}
}

func TestAgent_FindSubstitutes_NoToolCall(t *testing.T) {
	// Provider returns content but no tool call — fallback should be tried.
	// Since the fakeProvider doesn't implement RetryAsJSON, we test the fallback path
	// by providing a response with no tool calls and expecting an error.
	provider := &fakeProvider{resp: llm.CompletionResponse{Content: "no tool call"}}
	agent := substitute.NewAgent(provider)

	_, err := agent.FindSubstitutes(context.Background(), "iPhone 15 Pro", "smartphone", 1000.0)
	// Should error because no tool call was made and fallback fails too.
	if err == nil {
		t.Fatal("expected error when no tool call returned, got nil")
	}
}

func TestAgent_FindSubstitutes_EmptySubstitutes(t *testing.T) {
	// Tool call returns empty array — valid response.
	provider := &fakeProvider{resp: substituteToolCall([]substitute.Substitute{})}
	agent := substitute.NewAgent(provider)

	result, err := agent.FindSubstitutes(context.Background(), "SomeProduct", "category", 500.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 substitutes, got %d", len(result))
	}
}

func TestAgent_FindSubstitutes_OptionalURLOmitted(t *testing.T) {
	subs := []substitute.Substitute{
		{
			Name:      "OnePlus 12",
			Brand:     "OnePlus",
			Price:     699.0,
			Currency:  "EUR",
			Retailer:  "Amazon France",
			Reason:    "Flagship killer",
			Advantage: "cheaper",
			// URL omitted intentionally
		},
	}
	provider := &fakeProvider{resp: substituteToolCall(subs)}
	agent := substitute.NewAgent(provider)

	result, err := agent.FindSubstitutes(context.Background(), "iPhone 15 Pro", "smartphone", 1000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].URL != "" {
		t.Errorf("expected empty URL, got %s", result[0].URL)
	}
}

func TestAgent_FindSubstitutes_AllAdvantageTypes(t *testing.T) {
	advantages := []string{"cheaper", "better_rated", "eco_friendly", "local"}
	for _, adv := range advantages {
		subs := []substitute.Substitute{
			{
				Name:      "Test Product",
				Brand:     "TestBrand",
				Price:     100.0,
				Currency:  "EUR",
				Retailer:  "Leclerc",
				Reason:    "test reason",
				Advantage: adv,
			},
		}
		provider := &fakeProvider{resp: substituteToolCall(subs)}
		agent := substitute.NewAgent(provider)

		result, err := agent.FindSubstitutes(context.Background(), "Original", "cat", 200.0)
		if err != nil {
			t.Fatalf("advantage=%s: unexpected error: %v", adv, err)
		}
		if result[0].Advantage != adv {
			t.Errorf("advantage=%s: got %s", adv, result[0].Advantage)
		}
	}
}

// ── cache tests ───────────────────────────────────────────────────────────────

func TestCache_SetAndGet(t *testing.T) {
	c := substitute.NewCache(time.Hour)
	resp := &substitute.SubstituteResponse{
		OptionID:    "opt-1",
		ProductName: "iPhone 15 Pro",
		Substitutes: sampleSubstitutes(),
		GeneratedAt: time.Now(),
	}
	c.Set("opt-1", resp)

	got, ok := c.Get("opt-1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.OptionID != "opt-1" {
		t.Errorf("expected optionId opt-1, got %s", got.OptionID)
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	c := substitute.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := substitute.NewCache(1 * time.Millisecond)
	resp := &substitute.SubstituteResponse{OptionID: "x", Substitutes: []substitute.Substitute{}}
	c.Set("x", resp)
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("x")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}
