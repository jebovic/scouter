package seasonal_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/seasonal"
)

type fakeAgent struct {
	result *seasonal.Prediction
	err    error
	called int
}

func (f *fakeAgent) Predict(_ context.Context, itemName, category string) (*seasonal.Prediction, error) {
	f.called++
	return f.result, f.err
}

func samplePrediction() *seasonal.Prediction {
	return &seasonal.Prediction{
		ItemName:       "MacBook Pro",
		Category:       "electronics",
		BestMonths:     []string{"Novembre", "Décembre"},
		AvgDiscountPct: 20.0,
		Events:         []string{"Black Friday"},
		NextEventDays:  45,
		Rationale:      "Les remises sont plus importantes en fin d'année.",
		Confidence:     "high",
	}
}

func doGet(h *seasonal.Handler, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	h.GetPrediction(w, req)
	return w
}

func TestGetPrediction_AgentCalledOnMiss(t *testing.T) {
	agent := &fakeAgent{result: samplePrediction()}
	cache := seasonal.NewCache(24 * time.Hour)
	h := seasonal.NewHandler(agent, cache)

	w := doGet(h, "/api/seasonal?itemName=MacBook+Pro&category=electronics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got seasonal.Prediction
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemName != "MacBook Pro" {
		t.Errorf("expected itemName MacBook Pro, got %s", got.ItemName)
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetPrediction_EmptyItemName_Returns400(t *testing.T) {
	agent := &fakeAgent{result: samplePrediction()}
	cache := seasonal.NewCache(24 * time.Hour)
	h := seasonal.NewHandler(agent, cache)

	w := doGet(h, "/api/seasonal?itemName=")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetPrediction_CacheHitSkipsAgent(t *testing.T) {
	agent := &fakeAgent{result: samplePrediction()}
	cache := seasonal.NewCache(24 * time.Hour)
	h := seasonal.NewHandler(agent, cache)

	doGet(h, "/api/seasonal?itemName=MacBook+Pro&category=electronics")
	doGet(h, "/api/seasonal?itemName=MacBook+Pro&category=electronics")

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetPrediction_AgentError_ReturnsDegradedDTO(t *testing.T) {
	agent := &fakeAgent{err: errors.New("llm unavailable")}
	cache := seasonal.NewCache(24 * time.Hour)
	h := seasonal.NewHandler(agent, cache)

	w := doGet(h, "/api/seasonal?itemName=MacBook+Pro&category=electronics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded, got %d: %s", w.Code, w.Body.String())
	}
	var got seasonal.Prediction
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
