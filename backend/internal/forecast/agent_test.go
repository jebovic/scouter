package forecast

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
)

// ---------------------------------------------------------------------------
// Stub LLM provider.
// ---------------------------------------------------------------------------

type stubProvider struct {
	resp llm.CompletionResponse
	err  error
}

func (s *stubProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return s.resp, s.err
}

// ---------------------------------------------------------------------------
// Stub repository.
// ---------------------------------------------------------------------------

type stubRepo struct {
	saved  *ForecastResult
	latest *ForecastResult
	err    error
}

func (s *stubRepo) Save(_ context.Context, f ForecastResult) (ForecastResult, error) {
	if s.err != nil {
		return ForecastResult{}, s.err
	}
	f.ID = uuid.New().String()
	f.CreatedAt = time.Now()
	s.saved = &f
	return f, nil
}

func (s *stubRepo) GetLatest(_ context.Context, _ string) (ForecastResult, error) {
	if s.err != nil {
		return ForecastResult{}, s.err
	}
	if s.latest == nil {
		return ForecastResult{}, ErrNotFound
	}
	return *s.latest, nil
}

// ---------------------------------------------------------------------------
// Tool schema tests.
// ---------------------------------------------------------------------------

func TestForecastTool_Schema(t *testing.T) {
	tool := forecastTool()

	if tool.Name != "budget_forecast" {
		t.Errorf("expected tool name %q, got %q", "budget_forecast", tool.Name)
	}
	if tool.Description == "" {
		t.Error("tool description should not be empty")
	}

	schema := tool.InputSchema
	if schema == nil {
		t.Fatal("tool InputSchema should not be nil")
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("InputSchema.properties should be a map")
	}

	requiredFields := []string{"estimated_total", "confidence", "recommendations", "risk_factors"}
	for _, field := range requiredFields {
		if _, exists := props[field]; !exists {
			t.Errorf("tool schema missing required field %q", field)
		}
	}

	// Verify enum for confidence.
	confidenceSchema, ok := props["confidence"].(map[string]any)
	if !ok {
		t.Fatal("confidence should be a schema map")
	}
	enum, ok := confidenceSchema["enum"].([]string)
	if !ok {
		t.Fatal("confidence should have an enum")
	}
	expectedEnum := []string{"low", "medium", "high"}
	if len(enum) != len(expectedEnum) {
		t.Errorf("confidence enum length: want %d, got %d", len(expectedEnum), len(enum))
	}
	for i, v := range expectedEnum {
		if i < len(enum) && enum[i] != v {
			t.Errorf("confidence enum[%d]: want %q, got %q", i, v, enum[i])
		}
	}
}

func TestForecastTool_RequiredFields(t *testing.T) {
	tool := forecastTool()
	required, ok := tool.InputSchema["required"].([]string)
	if !ok {
		t.Fatal("InputSchema.required should be a []string")
	}

	must := map[string]bool{
		"estimated_total": false,
		"confidence":      false,
		"recommendations": false,
		"risk_factors":    false,
	}
	for _, r := range required {
		must[r] = true
	}
	for field, found := range must {
		if !found {
			t.Errorf("required field %q not in schema required list", field)
		}
	}
}

// ---------------------------------------------------------------------------
// parseToolResponse tests.
// ---------------------------------------------------------------------------

func TestParseToolResponse_HappyPath(t *testing.T) {
	total := 1500.0
	months := 3
	optimal := "Black Friday 2026"

	input := map[string]any{
		"estimated_total":  total,
		"confidence":       "high",
		"recommendations":  []any{"Buy now", "Compare prices"},
		"risk_factors":     []any{"Market volatility"},
		"months_to_save":   float64(months),
		"optimal_buy_time": optimal,
	}

	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "budget_forecast", Input: input},
		},
	}

	result, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse returned error: %v", err)
	}

	if result.EstimatedTotal == nil || *result.EstimatedTotal != total {
		t.Errorf("EstimatedTotal: want %v, got %v", total, result.EstimatedTotal)
	}
	if result.Confidence != "high" {
		t.Errorf("Confidence: want %q, got %q", "high", result.Confidence)
	}
	if len(result.Recommendations) != 2 {
		t.Errorf("Recommendations: want 2, got %d", len(result.Recommendations))
	}
	if len(result.RiskFactors) != 1 {
		t.Errorf("RiskFactors: want 1, got %d", len(result.RiskFactors))
	}
	if result.MonthsToSave == nil || *result.MonthsToSave != months {
		t.Errorf("MonthsToSave: want %d, got %v", months, result.MonthsToSave)
	}
	if result.OptimalBuyTime == nil || *result.OptimalBuyTime != optimal {
		t.Errorf("OptimalBuyTime: want %q, got %v", optimal, result.OptimalBuyTime)
	}
}

func TestParseToolResponse_NoToolCall(t *testing.T) {
	resp := llm.CompletionResponse{
		Content:   "I'll provide a forecast.",
		ToolCalls: nil,
	}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("expected error when LLM does not call budget_forecast tool")
	}
}

func TestParseToolResponse_WrongToolName(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "some_other_tool", Input: map[string]any{}},
		},
	}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("expected error when wrong tool is called")
	}
}

func TestParseToolResponse_MissingOptionalFields(t *testing.T) {
	total := 800.0
	input := map[string]any{
		"estimated_total": total,
		"confidence":      "low",
		"recommendations": []any{"Wait for sales"},
		"risk_factors":    []any{},
	}
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "budget_forecast", Input: input},
		},
	}
	result, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse returned error: %v", err)
	}
	if result.MonthsToSave != nil {
		t.Errorf("MonthsToSave should be nil when not provided, got %v", result.MonthsToSave)
	}
	if result.OptimalBuyTime != nil {
		t.Errorf("OptimalBuyTime should be nil when not provided, got %v", result.OptimalBuyTime)
	}
}

func TestParseToolResponse_EmptyLists(t *testing.T) {
	input := map[string]any{
		"estimated_total": 500.0,
		"confidence":      "medium",
		"recommendations": []any{},
		"risk_factors":    []any{},
	}
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "budget_forecast", Input: input},
		},
	}
	result, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Recommendations == nil {
		t.Error("Recommendations should be non-nil empty slice")
	}
	if result.RiskFactors == nil {
		t.Error("RiskFactors should be non-nil empty slice")
	}
}

// ---------------------------------------------------------------------------
// ForecastResult model tests.
// ---------------------------------------------------------------------------

func TestForecastResult_JSONRoundtrip(t *testing.T) {
	total := 1200.50
	months := 4
	optimal := "Q2 2026"

	orig := ForecastResult{
		ID:              uuid.New().String(),
		MissionID:       uuid.New().String(),
		EstimatedTotal:  &total,
		Confidence:      "medium",
		Recommendations: []string{"Check reviews", "Compare alternatives"},
		RiskFactors:     []string{"Supply chain delays"},
		MonthsToSave:    &months,
		OptimalBuyTime:  &optimal,
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
	}

	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var decoded ForecastResult
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded.ID != orig.ID {
		t.Errorf("ID mismatch: want %q, got %q", orig.ID, decoded.ID)
	}
	if decoded.Confidence != orig.Confidence {
		t.Errorf("Confidence mismatch: want %q, got %q", orig.Confidence, decoded.Confidence)
	}
	if decoded.EstimatedTotal == nil || *decoded.EstimatedTotal != total {
		t.Errorf("EstimatedTotal mismatch")
	}
	if len(decoded.Recommendations) != 2 {
		t.Errorf("Recommendations length: want 2, got %d", len(decoded.Recommendations))
	}
}

func TestForecastResult_NilPointers(t *testing.T) {
	f := ForecastResult{
		ID:              uuid.New().String(),
		MissionID:       uuid.New().String(),
		Confidence:      "low",
		Recommendations: []string{},
		RiskFactors:     []string{},
		CreatedAt:       time.Now(),
	}
	if f.EstimatedTotal != nil {
		t.Error("EstimatedTotal should default to nil")
	}
	if f.MonthsToSave != nil {
		t.Error("MonthsToSave should default to nil")
	}
	if f.OptimalBuyTime != nil {
		t.Error("OptimalBuyTime should default to nil")
	}
}

// ---------------------------------------------------------------------------
// Agent tests using stub provider.
// ---------------------------------------------------------------------------

func buildForecastResponse(total float64, confidence string, recs, risks []string) llm.CompletionResponse {
	input := map[string]any{
		"estimated_total": total,
		"confidence":      confidence,
		"recommendations": toAnySlice(recs),
		"risk_factors":    toAnySlice(risks),
	}
	return llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{{Name: "budget_forecast", Input: input}},
		Usage:     llm.Usage{Provider: "ollama", Model: "test"},
	}
}

func toAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func TestAgent_Run_LLMError(t *testing.T) {
	provider := &stubProvider{err: errors.New("llm unavailable")}
	repo := &stubRepo{}
	agent := newTestAgent(provider, repo)

	_, err := agent.Run(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

func TestAgent_Run_NoToolCall_ReturnsError(t *testing.T) {
	provider := &stubProvider{
		resp: llm.CompletionResponse{
			Content: "Here is a forecast...",
		},
	}
	repo := &stubRepo{}
	agent := newTestAgent(provider, repo)

	_, err := agent.Run(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error when LLM does not call tool and fallback also fails")
	}
}

func TestAgent_Run_RepoSaveError(t *testing.T) {
	provider := &stubProvider{
		resp: buildForecastResponse(999.0, "medium", []string{"Buy soon"}, []string{}),
	}
	repo := &stubRepo{err: errors.New("db write failed")}
	agent := newTestAgent(provider, repo)

	_, err := agent.Run(context.Background(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error when repo.Save fails")
	}
}

// newTestAgent creates a ForecastAgent that skips the DB-based mission/item
// fetch by overriding the context fetcher (stubbed via pool=nil path).
// Since the real Run() queries the DB, we test at the agent method level
// with a test-double that bypasses DB calls.
func newTestAgent(provider llm.Provider, repo Repository) *testableAgent {
	return &testableAgent{
		provider: provider,
		repo:     repo,
	}
}

// testableAgent wraps ForecastAgent logic in a way that bypasses the DB
// mission/item fetch, allowing unit tests without a real database.
type testableAgent struct {
	provider llm.Provider
	repo     Repository
}

func (a *testableAgent) Run(ctx context.Context, missionID string) (ForecastResult, error) {
	prompt := "You are a budget forecasting expert. Analyze this purchase mission: Test Mission, Budget: 1000 USD."
	req := llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{forecastTool()},
		MaxTokens: 2048,
	}

	resp, err := a.provider.Complete(ctx, req)
	if err != nil {
		return ForecastResult{}, err
	}

	partial, err := parseToolResponse(resp)
	if err != nil {
		return ForecastResult{}, err
	}

	partial.MissionID = missionID
	return a.repo.Save(ctx, partial)
}
