package carbon_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/carbon"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeAgent struct {
	result *carbon.CarbonEstimate
	err    error
	called int
}

func (f *fakeAgent) Estimate(_ context.Context, itemName, category string) (*carbon.CarbonEstimate, error) {
	f.called++
	return f.result, f.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func sampleEstimate() *carbon.CarbonEstimate {
	return &carbon.CarbonEstimate{
		ItemName:   "MacBook Pro",
		Category:   "electronics",
		KgCO2:      185.0,
		Label:      "Très élevé",
		Tips:       []string{"Prolongez la durée de vie de votre appareil", "Choisissez un reconditionné"},
		Confidence: "high",
	}
}

func doGet(h *carbon.Handler, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.GetEstimate(w, req)
	return w
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetEstimate_AgentCalledOnMiss(t *testing.T) {
	agent := &fakeAgent{result: sampleEstimate()}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	w := doGet(h, "/api/carbon?itemName=MacBook+Pro&category=electronics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got carbon.CarbonEstimate
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemName != "MacBook Pro" {
		t.Errorf("expected itemName MacBook Pro, got %s", got.ItemName)
	}
	if got.KgCO2 != 185.0 {
		t.Errorf("expected kgCO2 185.0, got %f", got.KgCO2)
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetEstimate_CacheHitSkipsAgent(t *testing.T) {
	agent := &fakeAgent{result: sampleEstimate()}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	// First request — populates cache.
	w1 := doGet(h, "/api/carbon?itemName=MacBook+Pro&category=electronics")
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache.
	w2 := doGet(h, "/api/carbon?itemName=MacBook+Pro&category=electronics")
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetEstimate_EmptyItemName_Returns400(t *testing.T) {
	agent := &fakeAgent{result: sampleEstimate()}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	w := doGet(h, "/api/carbon?itemName=&category=electronics")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when itemName is empty, called %d times", agent.called)
	}
}

func TestGetEstimate_MissingItemName_Returns400(t *testing.T) {
	agent := &fakeAgent{result: sampleEstimate()}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	w := doGet(h, "/api/carbon")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetEstimate_NoCategoryParam(t *testing.T) {
	estimate := &carbon.CarbonEstimate{
		ItemName:   "Cafetière",
		Category:   "",
		KgCO2:      25.0,
		Label:      "Élevé",
		Tips:       []string{"Utilisez-la jusqu'à la fin de vie", "Réparez plutôt que remplacer"},
		Confidence: "medium",
	}
	agent := &fakeAgent{result: estimate}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	w := doGet(h, "/api/carbon?itemName=Cafeti%C3%A8re")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got carbon.CarbonEstimate
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemName != "Cafetière" {
		t.Errorf("expected itemName Cafetière, got %s", got.ItemName)
	}
}

func TestGetEstimate_AgentError_ReturnsDegradedDTO(t *testing.T) {
	agent := &fakeAgent{err: errors.New("llm unavailable")}
	cache := carbon.NewCache(24 * time.Hour)
	h := carbon.NewHandler(agent, cache)

	w := doGet(h, "/api/carbon?itemName=MacBook+Pro&category=electronics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded, got %d: %s", w.Code, w.Body.String())
	}
	var got carbon.CarbonEstimate
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode degraded response: %v", err)
	}
	if got.Confidence != "unavailable" {
		t.Errorf("expected confidence 'unavailable', got %q", got.Confidence)
	}
	if got.ItemName != "MacBook Pro" {
		t.Errorf("expected itemName 'MacBook Pro', got %q", got.ItemName)
	}
}
