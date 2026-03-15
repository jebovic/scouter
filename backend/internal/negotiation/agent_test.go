package negotiation_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/negotiation"
)

// --- stub provider ---

type stubProvider struct {
	resp llm.CompletionResponse
	err  error
}

func (s *stubProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return s.resp, s.err
}

// buildToolResponse creates a CompletionResponse that simulates a tool call
// returning the given NegotiationScript as JSON.
func buildToolResponse(script negotiation.NegotiationScript) llm.CompletionResponse {
	raw, _ := json.Marshal(script)
	var input map[string]any
	_ = json.Unmarshal(raw, &input)

	return llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "generate_negotiation_script", Input: input},
		},
	}
}

// --- NegotiationInput validation ---

func TestNewAgent_NilProvider_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic with nil provider")
		}
	}()
	negotiation.NewAgent(nil)
}

func TestValidateInput_EmptyOptionName(t *testing.T) {
	agent := negotiation.NewAgent(&stubProvider{})
	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}
	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty option name")
	}
}

func TestValidateInput_EmptyMissionName(t *testing.T) {
	agent := negotiation.NewAgent(&stubProvider{})
	input := negotiation.NegotiationInput{
		MissionName: "",
		OptionName:  "MacBook Pro",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}
	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty mission name")
	}
}

func TestValidateInput_NegativeOptionPrice(t *testing.T) {
	agent := negotiation.NewAgent(&stubProvider{})
	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro",
		OptionPrice: -100.0,
		Budget:      1000.0,
	}
	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for negative option price")
	}
}

func TestValidateInput_ZeroOptionPrice(t *testing.T) {
	agent := negotiation.NewAgent(&stubProvider{})
	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro",
		OptionPrice: 0.0,
		Budget:      1000.0,
	}
	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for zero option price")
	}
}

func TestValidateInput_NegativeBudget(t *testing.T) {
	agent := negotiation.NewAgent(&stubProvider{})
	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro",
		OptionPrice: 1200.0,
		Budget:      -500.0,
	}
	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for negative budget")
	}
}

// --- Happy-path ---

func TestRun_HappyPath_ReturnsScript(t *testing.T) {
	expected := negotiation.NegotiationScript{
		OpeningOffer:      1050.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{"Competitor offers same spec for $1,050"},
		CounterOfferScript: []string{
			"Start at $1,050",
			"If counter at $1,100, accept",
		},
		ConfidenceScore: 0.8,
	}

	provider := &stubProvider{resp: buildToolResponse(expected)}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		Category:    "computing",
		Budget:      1000.0,
		Currency:    "USD",
		OptionName:  "MacBook Pro 14",
		OptionPrice: 1200.0,
		Attributes:  map[string]string{"ram": "16GB", "storage": "512GB"},
		Constraints: []negotiation.ConstraintInfo{
			{Key: "delivery", Label: "Delivery", Value: "2 weeks", Type: "soft"},
		},
		MarketPrices: []negotiation.MarketPrice{
			{Merchant: "Amazon", Price: 1180.0},
			{Merchant: "BestBuy", Price: 1199.0},
		},
	}

	script, err := agent.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script.OpeningOffer != expected.OpeningOffer {
		t.Errorf("OpeningOffer: got %.2f, want %.2f", script.OpeningOffer, expected.OpeningOffer)
	}
	if script.WalkAwayPrice != expected.WalkAwayPrice {
		t.Errorf("WalkAwayPrice: got %.2f, want %.2f", script.WalkAwayPrice, expected.WalkAwayPrice)
	}
	if script.SuggestedDiscount != expected.SuggestedDiscount {
		t.Errorf("SuggestedDiscount: got %.2f, want %.2f", script.SuggestedDiscount, expected.SuggestedDiscount)
	}
	if len(script.TalkingPoints) != 1 {
		t.Errorf("TalkingPoints: got %d, want 1", len(script.TalkingPoints))
	}
	if len(script.CounterOfferScript) != 2 {
		t.Errorf("CounterOfferScript: got %d steps, want 2", len(script.CounterOfferScript))
	}
	if script.ConfidenceScore != expected.ConfidenceScore {
		t.Errorf("ConfidenceScore: got %.2f, want %.2f", script.ConfidenceScore, expected.ConfidenceScore)
	}
}

func TestRun_WithHistoricalLow(t *testing.T) {
	historicalLow := 1050.0
	expected := negotiation.NegotiationScript{
		OpeningOffer:      1050.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 8.0,
		TalkingPoints:     []string{"Historical low was $1,050"},
		CounterOfferScript: []string{"Open at historical low price"},
		ConfidenceScore:   0.75,
	}

	provider := &stubProvider{resp: buildToolResponse(expected)}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName:   "Laptop Purchase",
		OptionName:    "MacBook Pro 14",
		OptionPrice:   1200.0,
		Budget:        1000.0,
		HistoricalLow: &historicalLow,
	}

	script, err := agent.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if script.OpeningOffer != 1050.0 {
		t.Errorf("expected opening offer at historical low, got %.2f", script.OpeningOffer)
	}
}

func TestRun_EmptyMarketPrices_StillSucceeds(t *testing.T) {
	expected := negotiation.NegotiationScript{
		OpeningOffer:      900.0,
		WalkAwayPrice:     1000.0,
		SuggestedDiscount: 15.0,
		TalkingPoints:     []string{"Over budget"},
		CounterOfferScript: []string{"Start low"},
		ConfidenceScore:   0.6,
	}

	provider := &stubProvider{resp: buildToolResponse(expected)}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName:  "Laptop Purchase",
		OptionName:   "MacBook Pro 14",
		OptionPrice:  1200.0,
		Budget:       1000.0,
		MarketPrices: []negotiation.MarketPrice{}, // empty
	}

	_, err := agent.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error with empty market prices: %v", err)
	}
}

// --- LLM error handling ---

func TestRun_LLMError_ReturnsError(t *testing.T) {
	provider := &stubProvider{err: errors.New("llm unavailable")}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro 14",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}

	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

func TestRun_NoToolCall_ReturnsError(t *testing.T) {
	// Provider returns a response without the expected tool call
	provider := &stubProvider{
		resp: llm.CompletionResponse{
			Content:   "I would suggest negotiating firmly.",
			ToolCalls: nil,
		},
	}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro 14",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}

	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when LLM does not call the tool")
	}
}

func TestRun_WrongToolName_ReturnsError(t *testing.T) {
	provider := &stubProvider{
		resp: llm.CompletionResponse{
			ToolCalls: []llm.ToolCall{
				{Name: "some_other_tool", Input: map[string]any{}},
			},
		},
	}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro 14",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}

	_, err := agent.Run(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when wrong tool is called")
	}
}

// --- NegotiationScript validation ---

func TestScript_WalkAwayNotLessThanOpeningOffer_IsInvalid(t *testing.T) {
	// The agent returns a script where walkAway < openingOffer — this is a logical error
	badScript := negotiation.NegotiationScript{
		OpeningOffer:      1100.0, // higher than walk-away
		WalkAwayPrice:     1000.0, // walk-away should be >= opening
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{"point"},
		CounterOfferScript: []string{"step"},
		ConfidenceScore:   0.8,
	}

	if err := badScript.Validate(); err == nil {
		t.Fatal("expected validation error: walkAwayPrice must be >= openingOffer")
	}
}

func TestScript_DiscountOutOfBounds_IsInvalid(t *testing.T) {
	cases := []struct {
		discount float64
		name     string
	}{
		{-1.0, "negative discount"},
		{81.0, "discount over 80%"},
	}
	for _, tc := range cases {
		script := negotiation.NegotiationScript{
			OpeningOffer:      1000.0,
			WalkAwayPrice:     1100.0,
			SuggestedDiscount: tc.discount,
			TalkingPoints:     []string{"point"},
			CounterOfferScript: []string{"step"},
			ConfidenceScore:   0.5,
		}
		if err := script.Validate(); err == nil {
			t.Errorf("expected validation error for %s (discount=%.1f)", tc.name, tc.discount)
		}
	}
}

func TestScript_ConfidenceScoreOutOfBounds_IsInvalid(t *testing.T) {
	cases := []struct {
		score float64
		name  string
	}{
		{-0.1, "negative confidence"},
		{1.1, "confidence over 1.0"},
	}
	for _, tc := range cases {
		script := negotiation.NegotiationScript{
			OpeningOffer:      1000.0,
			WalkAwayPrice:     1100.0,
			SuggestedDiscount: 10.0,
			TalkingPoints:     []string{"point"},
			CounterOfferScript: []string{"step"},
			ConfidenceScore:   tc.score,
		}
		if err := script.Validate(); err == nil {
			t.Errorf("expected validation error for %s (score=%.1f)", tc.name, tc.score)
		}
	}
}

func TestScript_EmptyTalkingPoints_IsInvalid(t *testing.T) {
	script := negotiation.NegotiationScript{
		OpeningOffer:      1000.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{},
		CounterOfferScript: []string{"step"},
		ConfidenceScore:   0.8,
	}
	if err := script.Validate(); err == nil {
		t.Fatal("expected validation error for empty talking points")
	}
}

func TestScript_EmptyCounterOfferScript_IsInvalid(t *testing.T) {
	script := negotiation.NegotiationScript{
		OpeningOffer:      1000.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{"point"},
		CounterOfferScript: []string{},
		ConfidenceScore:   0.8,
	}
	if err := script.Validate(); err == nil {
		t.Fatal("expected validation error for empty counter offer script")
	}
}

func TestScript_ValidScript_PassesValidation(t *testing.T) {
	script := negotiation.NegotiationScript{
		OpeningOffer:      1000.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{"Competitor at $1,000"},
		CounterOfferScript: []string{"Open at $1,000", "Accept at $1,100"},
		ConfidenceScore:   0.8,
	}
	if err := script.Validate(); err != nil {
		t.Errorf("expected valid script to pass validation: %v", err)
	}
}

// --- Context cancellation ---

func TestRun_CanceledContext_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	provider := &stubProvider{err: context.Canceled}
	agent := negotiation.NewAgent(provider)

	input := negotiation.NegotiationInput{
		MissionName: "Laptop Purchase",
		OptionName:  "MacBook Pro 14",
		OptionPrice: 1200.0,
		Budget:      1000.0,
	}

	_, err := agent.Run(ctx, input)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
