package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// pingableMock is a Provider that also implements Pinger.
type pingableMock struct {
	pingErr error
	latency time.Duration
}

func (p *pingableMock) Complete(_ context.Context, _ CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{}, nil
}

func (p *pingableMock) Ping(_ context.Context) error {
	if p.latency > 0 {
		time.Sleep(p.latency)
	}
	return p.pingErr
}

// nonPingableMock is a Provider without Ping support.
type nonPingableMock struct{}

func (n *nonPingableMock) Complete(_ context.Context, _ CompletionRequest) (CompletionResponse, error) {
	return CompletionResponse{}, nil
}

func makePoolEntry(name string, prov Provider, caps Capability) ModelEntry {
	return ModelEntry{
		Name:         name,
		Provider:     prov,
		Capabilities: caps,
		Priority:     1,
		Timeout:      5 * time.Second,
	}
}

func TestHealthChecker_Check_PingSuccess(t *testing.T) {
	prov := &pingableMock{}
	pool := ModelPool{makePoolEntry("model-a", prov, CapText|CapToolUse)}
	checker := NewHealthChecker(pool, nil)

	health := checker.Check(context.Background())

	if len(health.Models) != 1 {
		t.Fatalf("expected 1 model status, got %d", len(health.Models))
	}
	status := health.Models[0]
	if status.Name != "model-a" {
		t.Errorf("name: want %q, got %q", "model-a", status.Name)
	}
	if !status.Healthy {
		t.Error("expected Healthy=true when Ping succeeds")
	}
	if status.CircuitState != "closed" {
		t.Errorf("circuit state: want %q, got %q", "closed", status.CircuitState)
	}
}

func TestHealthChecker_Check_PingFailure(t *testing.T) {
	prov := &pingableMock{pingErr: errors.New("connection refused")}
	pool := ModelPool{makePoolEntry("model-b", prov, CapText)}
	checker := NewHealthChecker(pool, nil)

	health := checker.Check(context.Background())

	if len(health.Models) != 1 {
		t.Fatalf("expected 1 model status, got %d", len(health.Models))
	}
	status := health.Models[0]
	if status.Healthy {
		t.Error("expected Healthy=false when Ping fails")
	}
}

func TestHealthChecker_Check_NoPingSupport_AssumedHealthy(t *testing.T) {
	prov := &nonPingableMock{}
	pool := ModelPool{makePoolEntry("model-noping", prov, CapText)}
	checker := NewHealthChecker(pool, nil)

	health := checker.Check(context.Background())

	status := health.Models[0]
	// circuit is "closed" (no breaker) → assumed healthy
	if !status.Healthy {
		t.Error("expected Healthy=true for non-Pinger provider with closed circuit")
	}
	if status.CircuitState != "closed" {
		t.Errorf("circuit state: want %q, got %q", "closed", status.CircuitState)
	}
}

func TestHealthChecker_Check_CircuitBreaker_Open(t *testing.T) {
	prov := &nonPingableMock{}
	pool := ModelPool{makePoolEntry("model-open", prov, CapText)}

	cb := newCircuitBreaker()
	// Force open circuit by recording enough failures
	for i := 0; i < cb.maxFailures; i++ {
		cb.recordFailure()
	}

	breakers := map[string]*circuitBreaker{"model-open": cb}
	checker := NewHealthChecker(pool, breakers)

	health := checker.Check(context.Background())

	status := health.Models[0]
	if status.CircuitState != "open" {
		t.Errorf("circuit state: want %q, got %q", "open", status.CircuitState)
	}
	// Non-Pinger with open circuit → not healthy
	if status.Healthy {
		t.Error("expected Healthy=false when circuit is open and provider has no Ping")
	}
}

func TestHealthChecker_Check_MultipleModels(t *testing.T) {
	healthy := &pingableMock{}
	unhealthy := &pingableMock{pingErr: errors.New("down")}

	pool := ModelPool{
		makePoolEntry("ok", healthy, CapText),
		makePoolEntry("bad", unhealthy, CapText),
	}
	checker := NewHealthChecker(pool, nil)

	health := checker.Check(context.Background())

	if len(health.Models) != 2 {
		t.Fatalf("expected 2 model statuses, got %d", len(health.Models))
	}

	for _, s := range health.Models {
		switch s.Name {
		case "ok":
			if !s.Healthy {
				t.Error("model ok: expected Healthy=true")
			}
		case "bad":
			if s.Healthy {
				t.Error("model bad: expected Healthy=false")
			}
		default:
			t.Errorf("unexpected model name: %q", s.Name)
		}
	}
}

func TestCapabilityNames(t *testing.T) {
	tests := []struct {
		cap  Capability
		want []string
	}{
		{CapText, []string{"text"}},
		{CapToolUse, []string{"tool_use"}},
		{CapLongCtx, []string{"long_ctx"}},
		{CapText | CapToolUse, []string{"text", "tool_use"}},
		{CapText | CapToolUse | CapLongCtx, []string{"text", "tool_use", "long_ctx"}},
		{0, nil},
	}

	for _, tt := range tests {
		got := capabilityNames(tt.cap)
		if len(got) != len(tt.want) {
			t.Errorf("capabilityNames(%d): want %v, got %v", tt.cap, tt.want, got)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("capabilityNames(%d)[%d]: want %q, got %q", tt.cap, i, tt.want[i], got[i])
			}
		}
	}
}

func TestHealthHandler_WritesJSON(t *testing.T) {
	prov := &pingableMock{}
	pool := ModelPool{makePoolEntry("test-model", prov, CapText|CapToolUse)}
	checker := NewHealthChecker(pool, nil)

	handler := HealthHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/api/health/llm", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code: want %d, got %d", http.StatusOK, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: want %q, got %q", "application/json", ct)
	}

	var health PoolHealth
	if err := json.NewDecoder(rec.Body).Decode(&health); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(health.Models) != 1 {
		t.Fatalf("expected 1 model in response, got %d", len(health.Models))
	}
	if health.Models[0].Name != "test-model" {
		t.Errorf("model name: want %q, got %q", "test-model", health.Models[0].Name)
	}
}

func TestHealthHandler_CapabilitiesInResponse(t *testing.T) {
	prov := &pingableMock{}
	pool := ModelPool{makePoolEntry("caps-model", prov, CapText|CapToolUse|CapLongCtx)}
	checker := NewHealthChecker(pool, nil)

	handler := HealthHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/api/health/llm", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	var health PoolHealth
	if err := json.NewDecoder(rec.Body).Decode(&health); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	caps := health.Models[0].Capabilities
	expected := []string{"text", "tool_use", "long_ctx"}
	if len(caps) != len(expected) {
		t.Fatalf("capabilities: want %v, got %v", expected, caps)
	}
	for i, c := range caps {
		if c != expected[i] {
			t.Errorf("capabilities[%d]: want %q, got %q", i, expected[i], c)
		}
	}
}

func TestHealthChecker_Check_EmptyPool(t *testing.T) {
	checker := NewHealthChecker(ModelPool{}, nil)
	health := checker.Check(context.Background())
	if len(health.Models) != 0 {
		t.Errorf("expected 0 models, got %d", len(health.Models))
	}
}

func TestCircuitBreaker_State(t *testing.T) {
	t.Run("new circuit is closed", func(t *testing.T) {
		cb := newCircuitBreaker()
		if cb.state() != "closed" {
			t.Errorf("new circuit: want %q, got %q", "closed", cb.state())
		}
	})

	t.Run("circuit with recorded success is closed", func(t *testing.T) {
		cb := newCircuitBreaker()
		cb.recordSuccess()
		if cb.state() != "closed" {
			t.Errorf("after success: want %q, got %q", "closed", cb.state())
		}
	})

	t.Run("circuit after max failures is open", func(t *testing.T) {
		cb := newCircuitBreaker()
		for i := 0; i < cb.maxFailures; i++ {
			cb.recordFailure()
		}
		if cb.state() != "open" {
			t.Errorf("after max failures: want %q, got %q", "open", cb.state())
		}
	})
}
