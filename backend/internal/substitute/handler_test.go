package substitute_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/substitute"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeSubstituteAgent struct {
	result []substitute.Substitute
	err    error
	called int
}

func (f *fakeSubstituteAgent) FindSubstitutes(_ context.Context, name, category string, budget float64) ([]substitute.Substitute, error) {
	f.called++
	return f.result, f.err
}

type fakeOptionGetter struct {
	name     string
	category string
	budget   float64
	err      error
}

func (f *fakeOptionGetter) GetOptionInfo(_ context.Context, optionID string) (name, category string, budget float64, err error) {
	return f.name, f.category, f.budget, f.err
}

// ── router helper ────────────────────────────────────────────────────────────

func mountSubstituteRouter(h *substitute.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/options/{id}/substitutes", h.GetSubstitutes)
	return r
}

func goodSubstituteResponse(optionID string) *substitute.SubstituteResponse {
	return &substitute.SubstituteResponse{
		OptionID:    optionID,
		ProductName: "iPhone 15 Pro",
		Substitutes: sampleSubstitutes(),
		GeneratedAt: time.Now(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetSubstitutes_HappyPath(t *testing.T) {
	const optID = "11111111-1111-1111-1111-111111111111"
	agent := &fakeSubstituteAgent{result: sampleSubstitutes()}
	getter := &fakeOptionGetter{name: "iPhone 15 Pro", category: "smartphone", budget: 1000.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got substitute.SubstituteResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.OptionID != optID {
		t.Errorf("expected optionId %s, got %s", optID, got.OptionID)
	}
	if len(got.Substitutes) != 2 {
		t.Errorf("expected 2 substitutes, got %d", len(got.Substitutes))
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetSubstitutes_CacheHit(t *testing.T) {
	const optID = "22222222-2222-2222-2222-222222222222"
	agent := &fakeSubstituteAgent{result: sampleSubstitutes()}
	getter := &fakeOptionGetter{name: "iPhone 15 Pro", category: "smartphone", budget: 1000.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)
	router := mountSubstituteRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called exactly once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetSubstitutes_OptionNotFound(t *testing.T) {
	const optID = "33333333-3333-3333-3333-333333333333"
	agent := &fakeSubstituteAgent{}
	getter := &fakeOptionGetter{name: ""} // empty name → not found
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when option not found, called %d times", agent.called)
	}
}

func TestGetSubstitutes_GetterError(t *testing.T) {
	const optID = "44444444-4444-4444-4444-444444444444"
	agent := &fakeSubstituteAgent{}
	getter := &fakeOptionGetter{err: errors.New("db error")}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when getter fails, called %d times", agent.called)
	}
}

func TestGetSubstitutes_AgentError(t *testing.T) {
	const optID = "55555555-5555-5555-5555-555555555555"
	agent := &fakeSubstituteAgent{err: errors.New("llm unavailable")}
	getter := &fakeOptionGetter{name: "Sony TV", category: "electronics", budget: 800.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}
