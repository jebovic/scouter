package summary_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/summary"
)

// ── fake LLM provider ────────────────────────────────────────────────────────

type fakeLLMProvider struct {
	content string
	err     error
	calls   int
}

func (f *fakeLLMProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	f.calls++
	return llm.CompletionResponse{Content: f.content}, f.err
}

// ── fixtures ─────────────────────────────────────────────────────────────────

func makeMission() mission.Mission {
	return mission.Mission{
		ID:       uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:     "Nouveau Laptop",
		Category: "computing",
		Budget:   1500.0,
		Currency: "EUR",
	}
}

func makeOptions() []option.Option {
	return []option.Option{
		{
			ID:    uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Name:  "MacBook Air M3",
			Badge: "recommended",
			PriceRange: &option.PriceRange{
				Min:  1199.0,
				Best: 1199.0,
				Max:  1299.0,
			},
		},
		{
			ID:    uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Name:  "Dell XPS 15",
			Badge: "alternative",
			PriceRange: &option.PriceRange{
				Min:  1099.0,
				Best: 1099.0,
				Max:  1299.0,
			},
		},
	}
}

func makeShoppingItems() []shopping.Item {
	return []shopping.Item{
		{
			ID:       uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
			Name:     "MacBook Air M3",
			Merchant: "Apple Store",
			Price:    1199.0,
			Status:   "buy",
		},
	}
}

const sampleMarkdown = `# Rapport d'achat — Nouveau Laptop

## Résumé exécutif
Mission d'achat d'un laptop pour 1 500 EUR.

## Top 3 options recommandées
1. **MacBook Air M3** — Meilleur rapport qualité/prix
2. **Dell XPS 15** — Bonne alternative Windows

## Meilleur prix trouvé
Apple Store : 1 199 EUR

## Conseil d'achat
Acheter maintenant pendant les soldes.

## Évaluation budgétaire
Budget suffisant pour les deux options.`

// ── Agent.Generate happy-path ─────────────────────────────────────────────────

func TestAgent_Generate_HappyPath(t *testing.T) {
	provider := &fakeLLMProvider{content: sampleMarkdown}
	agent := summary.NewAgent(provider)

	content, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty markdown content")
	}
	if !strings.Contains(content, "Nouveau Laptop") {
		t.Errorf("expected content to mention mission name, got: %s", content)
	}
}

func TestAgent_Generate_ReturnsLLMContent(t *testing.T) {
	provider := &fakeLLMProvider{content: sampleMarkdown}
	agent := summary.NewAgent(provider)

	content, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != sampleMarkdown {
		t.Errorf("expected returned content to match LLM response")
	}
}

func TestAgent_Generate_CallsLLMOnce(t *testing.T) {
	provider := &fakeLLMProvider{content: sampleMarkdown}
	agent := summary.NewAgent(provider)

	_, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.calls != 1 {
		t.Errorf("expected LLM called exactly once, got %d", provider.calls)
	}
}

// ── Agent.Generate error paths ────────────────────────────────────────────────

func TestAgent_Generate_LLMError(t *testing.T) {
	provider := &fakeLLMProvider{err: errors.New("network timeout")}
	agent := summary.NewAgent(provider)

	_, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

// ── Agent.Generate edge cases ─────────────────────────────────────────────────

func TestAgent_Generate_NoOptions(t *testing.T) {
	provider := &fakeLLMProvider{content: sampleMarkdown}
	agent := summary.NewAgent(provider)

	// Should still call LLM with no options gracefully.
	content, err := agent.Generate(context.Background(), makeMission(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error with empty options: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty content even with no options")
	}
}

func TestAgent_Generate_UsesCapTextOnly(t *testing.T) {
	// Verify that no tool calls are in the request (text-only mode).
	var captured llm.CompletionRequest
	capturingProvider := &capturingFakeProvider{
		content: sampleMarkdown,
		capture: &captured,
	}
	agent := summary.NewAgent(capturingProvider)

	_, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.Tools) > 0 {
		t.Errorf("expected no tools in request (text-only), got %d tools", len(captured.Tools))
	}
}

func TestAgent_Generate_MaxTokensCapped(t *testing.T) {
	var captured llm.CompletionRequest
	capturingProvider := &capturingFakeProvider{
		content: sampleMarkdown,
		capture: &captured,
	}
	agent := summary.NewAgent(capturingProvider)

	_, err := agent.Generate(context.Background(), makeMission(), makeOptions(), makeShoppingItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.MaxTokens > 1500 {
		t.Errorf("expected MaxTokens <= 1500, got %d", captured.MaxTokens)
	}
	if captured.MaxTokens <= 0 {
		t.Errorf("expected MaxTokens > 0, got %d", captured.MaxTokens)
	}
}

// capturingFakeProvider captures the last CompletionRequest for inspection.
type capturingFakeProvider struct {
	content string
	capture *llm.CompletionRequest
}

func (f *capturingFakeProvider) Complete(_ context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	*f.capture = req
	return llm.CompletionResponse{Content: f.content}, nil
}

// ── ShoppingSummary model ─────────────────────────────────────────────────────

func TestShoppingSummary_Fields(t *testing.T) {
	id := uuid.New()
	missionID := uuid.New()
	now := time.Now().UTC()

	s := summary.ShoppingSummary{
		ID:          id,
		MissionID:   missionID,
		Content:     "# Test\nMarkdown content",
		GeneratedAt: now,
	}

	if s.ID != id {
		t.Errorf("expected ID %s, got %s", id, s.ID)
	}
	if s.MissionID != missionID {
		t.Errorf("expected MissionID %s, got %s", missionID, s.MissionID)
	}
	if s.Content != "# Test\nMarkdown content" {
		t.Errorf("unexpected content: %s", s.Content)
	}
	if s.GeneratedAt != now {
		t.Errorf("unexpected GeneratedAt: %v", s.GeneratedAt)
	}
}
