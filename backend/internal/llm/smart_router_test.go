package llm

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"
)

// mockProvider is a test double for Provider.
type mockProvider struct {
	resp CompletionResponse
	err  error
}

func (m *mockProvider) Complete(_ context.Context, _ CompletionRequest) (CompletionResponse, error) {
	return m.resp, m.err
}

// captureProvider records the last request it received.
type captureProvider struct {
	lastReq CompletionRequest
	resp    CompletionResponse
	err     error
}

func (c *captureProvider) Complete(_ context.Context, req CompletionRequest) (CompletionResponse, error) {
	c.lastReq = req
	return c.resp, c.err
}

// infraErr returns a net.OpError (treated as infra error by isInfraError).
func infraErr() error {
	return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestSmartRouter_SingleModelSuccess(t *testing.T) {
	p := &mockProvider{resp: CompletionResponse{Content: "hello"}}
	pool := NewModelPool(ModelEntry{
		Name:         "qwen3:14b",
		Provider:     p,
		Capabilities: CapText | CapToolUse,
		Priority:     1,
		Timeout:      5 * time.Second,
	})
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapToolUse})
	resp, err := sr.Complete(ctx, CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hello" {
		t.Errorf("content = %q, want %q", resp.Content, "hello")
	}
	if resp.ModelName != "qwen3:14b" {
		t.Errorf("model_name = %q, want %q", resp.ModelName, "qwen3:14b")
	}
	if resp.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", resp.Attempts)
	}
	if resp.Degraded {
		t.Error("degraded should be false for first-choice success")
	}
}

func TestSmartRouter_CascadeOnInfraError(t *testing.T) {
	primary := &mockProvider{err: infraErr()}
	secondary := &mockProvider{resp: CompletionResponse{Content: "fallback"}}
	pool := NewModelPool(
		ModelEntry{Name: "primary", Provider: primary, Capabilities: CapText, Priority: 1},
		ModelEntry{Name: "secondary", Provider: secondary, Capabilities: CapText, Priority: 2},
	)
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	resp, err := sr.Complete(ctx, CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("content = %q, want fallback", resp.Content)
	}
	if resp.ModelName != "secondary" {
		t.Errorf("model_name = %q, want secondary", resp.ModelName)
	}
	if resp.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", resp.Attempts)
	}
	if !resp.Degraded {
		t.Error("degraded should be true when second model used")
	}
	if !resp.WasFallback {
		t.Error("was_fallback should be true when second model used")
	}
}

func TestSmartRouter_NoInfraError_NosCascade(t *testing.T) {
	// 4xx-style error: should not cascade.
	modelErr := &StatusError{Code: 400}
	primary := &mockProvider{err: modelErr}
	secondary := &mockProvider{resp: CompletionResponse{Content: "should not reach"}}
	pool := NewModelPool(
		ModelEntry{Name: "primary", Provider: primary, Capabilities: CapText, Priority: 1},
		ModelEntry{Name: "secondary", Provider: secondary, Capabilities: CapText, Priority: 2},
	)
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, modelErr) {
		t.Errorf("expected 400 StatusError, got %v", err)
	}
}

func TestSmartRouter_AllCircuitsOpen_ReturnsErrAllModelsUnavailable(t *testing.T) {
	p := &mockProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:         "model-a",
		Provider:     p,
		Capabilities: CapText,
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	// Force circuit open.
	cb := sr.breakers["model-a"]
	for i := 0; i < 5; i++ {
		cb.recordFailure()
	}

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, ErrAllModelsUnavailable) {
		t.Errorf("expected ErrAllModelsUnavailable, got %v", err)
	}
}

func TestSmartRouter_NoCapabilityMatch_ReturnsErrAllModelsUnavailable(t *testing.T) {
	p := &mockProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:         "text-only",
		Provider:     p,
		Capabilities: CapText, // no CapToolUse
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapToolUse})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, ErrAllModelsUnavailable) {
		t.Errorf("expected ErrAllModelsUnavailable, got %v", err)
	}
}

func TestSmartRouter_NoRequestOpts_DefaultsToCapText(t *testing.T) {
	p := &mockProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:         "model",
		Provider:     p,
		Capabilities: CapText,
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	// No RequestOpts attached — should still succeed using CapText default.
	resp, err := sr.Complete(context.Background(), CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("content = %q, want ok", resp.Content)
	}
}

func TestSmartRouter_SystemPromptPrefix_Injected(t *testing.T) {
	cap := &captureProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:               "qwen3:14b",
		Provider:           cap,
		Capabilities:       CapText,
		Priority:           1,
		SystemPromptPrefix: "/no_think",
	})
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	req := CompletionRequest{Messages: []Message{{Role: "user", Content: "hello"}}}
	_, err := sr.Complete(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cap.lastReq.Messages) < 2 {
		t.Fatalf("expected injected system message, got %d messages", len(cap.lastReq.Messages))
	}
	if cap.lastReq.Messages[0].Role != "system" {
		t.Errorf("first message role = %q, want system", cap.lastReq.Messages[0].Role)
	}
	if cap.lastReq.Messages[0].Content != "/no_think" {
		t.Errorf("system content = %q, want /no_think", cap.lastReq.Messages[0].Content)
	}
	if cap.lastReq.Messages[1].Content != "hello" {
		t.Errorf("user content = %q, want hello", cap.lastReq.Messages[1].Content)
	}
}

func TestSmartRouter_SystemPromptPrefix_ExistingSystemMessage(t *testing.T) {
	cap := &captureProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:               "qwen3:14b",
		Provider:           cap,
		Capabilities:       CapText,
		Priority:           1,
		SystemPromptPrefix: "/no_think",
	})
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	req := CompletionRequest{Messages: []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "hello"},
	}}
	_, err := sr.Complete(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.lastReq.Messages[0].Content != "/no_think\nYou are a helpful assistant." {
		t.Errorf("system content = %q", cap.lastReq.Messages[0].Content)
	}
}

func TestSmartRouter_CircuitBreakerRecordsSuccess(t *testing.T) {
	p := &mockProvider{resp: CompletionResponse{Content: "ok"}}
	pool := NewModelPool(ModelEntry{
		Name:         "model",
		Provider:     p,
		Capabilities: CapText,
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	// Record some failures (not enough to open).
	cb := sr.breakers["model"]
	cb.recordFailure()
	cb.recordFailure()

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After success, circuit should be closed (failures reset).
	if cb.state() != "closed" {
		t.Errorf("circuit state = %q, want closed after success", cb.state())
	}
}

func TestSmartRouter_InfraError_RecordsFailure(t *testing.T) {
	p := &mockProvider{err: infraErr()}
	pool := NewModelPool(ModelEntry{
		Name:         "model",
		Provider:     p,
		Capabilities: CapText,
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, ErrAllModelsUnavailable) {
		t.Errorf("expected ErrAllModelsUnavailable, got %v", err)
	}

	// After 1 failure, circuit should still be closed (threshold is 3).
	cb := sr.breakers["model"]
	if cb.state() != "closed" {
		t.Errorf("circuit state = %q after 1 failure, want closed", cb.state())
	}
}

func TestSmartRouter_AllInfraErrors_Exhausted(t *testing.T) {
	p1 := &mockProvider{err: infraErr()}
	p2 := &mockProvider{err: infraErr()}
	pool := NewModelPool(
		ModelEntry{Name: "m1", Provider: p1, Capabilities: CapText, Priority: 1},
		ModelEntry{Name: "m2", Provider: p2, Capabilities: CapText, Priority: 2},
	)
	sr := NewSmartRouter(pool, testLogger())

	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, ErrAllModelsUnavailable) {
		t.Errorf("expected ErrAllModelsUnavailable, got %v", err)
	}
}

func TestSmartRouter_Breakers_Exposed(t *testing.T) {
	p := &mockProvider{}
	pool := NewModelPool(ModelEntry{Name: "m", Provider: p, Capabilities: CapText, Priority: 1})
	sr := NewSmartRouter(pool, testLogger())
	if sr.Breakers() == nil {
		t.Error("Breakers() should not be nil")
	}
	if _, ok := sr.Breakers()["m"]; !ok {
		t.Error("breaker for model 'm' should exist")
	}
}

func TestSmartRouter_HalfOpen_ProbeFailure_ReopensCircuit(t *testing.T) {
	p := &mockProvider{err: infraErr()}
	pool := NewModelPool(ModelEntry{
		Name:         "model",
		Provider:     p,
		Capabilities: CapText,
		Priority:     1,
	})
	sr := NewSmartRouter(pool, testLogger())

	// Force circuit open by hitting the failure threshold.
	cb := sr.breakers["model"]
	for i := 0; i < 3; i++ {
		cb.recordFailure()
	}
	if cb.state() != "open" {
		t.Fatalf("circuit should be open after 3 failures, got %q", cb.state())
	}

	// Manually expire the open window to put circuit in half_open state.
	cb.mu.Lock()
	cb.openUntil = time.Now().Add(-time.Second)
	cb.mu.Unlock()
	if cb.state() != "half_open" {
		t.Fatalf("circuit should be half_open after window expires, got %q", cb.state())
	}

	// Probe: allow() returns true for half_open, so Complete will try the model.
	// The model returns an infra error → recordFailure → circuit re-opens.
	ctx := WithRequestOpts(context.Background(), RequestOpts{Capabilities: CapText})
	_, err := sr.Complete(ctx, CompletionRequest{})
	if !errors.Is(err, ErrAllModelsUnavailable) {
		t.Errorf("expected ErrAllModelsUnavailable, got %v", err)
	}
	if cb.state() != "open" {
		t.Errorf("circuit should re-open after failed probe, got %q", cb.state())
	}
}

func TestInjectSystemPrefix_EmptyMessages(t *testing.T) {
	req := CompletionRequest{Messages: []Message{}}
	result := injectSystemPrefix(req, "/no_think")
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Role != "system" || result.Messages[0].Content != "/no_think" {
		t.Errorf("unexpected system message: %+v", result.Messages[0])
	}
}
