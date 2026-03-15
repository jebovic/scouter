package autotag_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/autotag"
	"github.com/jibei/scouter/internal/mission"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeMissionGetter struct {
	m   *mission.Mission
	err error
}

func (f *fakeMissionGetter) GetBySlug(_ context.Context, slug string) (*mission.Mission, error) {
	return f.m, f.err
}

type fakeTagger struct {
	suggestion *autotag.CategorySuggestion
	err        error
	called     int
}

func (f *fakeTagger) Suggest(_ context.Context, input autotag.TagInput) (*autotag.CategorySuggestion, error) {
	f.called++
	return f.suggestion, f.err
}

// ── router helper ─────────────────────────────────────────────────────────────

func mountAutotagRouter(h *autotag.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/missions/{slug}/suggest-category", h.Suggest)
	return r
}

func sampleMission() *mission.Mission {
	return &mission.Mission{
		Slug: "macbook-pro",
		Name: "Achat MacBook Pro",
	}
}

func sampleSuggestion() *autotag.CategorySuggestion {
	return &autotag.CategorySuggestion{
		Category:     "Informatique",
		Confidence:   0.92,
		Alternatives: []string{"Électronique", "Audiovisuel"},
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestSuggest_HappyPath(t *testing.T) {
	tagger := &fakeTagger{suggestion: sampleSuggestion()}
	getter := &fakeMissionGetter{m: sampleMission()}
	h := autotag.NewHandler(getter, tagger)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/macbook-pro/suggest-category", nil)
	w := httptest.NewRecorder()
	mountAutotagRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got autotag.CategorySuggestion
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Category != "Informatique" {
		t.Errorf("expected category Informatique, got %s", got.Category)
	}
	if got.Confidence != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", got.Confidence)
	}
	if len(got.Alternatives) != 2 {
		t.Errorf("expected 2 alternatives, got %d", len(got.Alternatives))
	}
	if tagger.called != 1 {
		t.Errorf("expected tagger called once, got %d", tagger.called)
	}
}

func TestSuggest_MissionNotFound(t *testing.T) {
	tagger := &fakeTagger{}
	getter := &fakeMissionGetter{m: nil}
	h := autotag.NewHandler(getter, tagger)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/nonexistent/suggest-category", nil)
	w := httptest.NewRecorder()
	mountAutotagRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if tagger.called != 0 {
		t.Errorf("tagger should not be called when mission not found, called %d times", tagger.called)
	}
}

func TestSuggest_MissionGetError(t *testing.T) {
	tagger := &fakeTagger{}
	getter := &fakeMissionGetter{err: errors.New("db error")}
	h := autotag.NewHandler(getter, tagger)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/broken/suggest-category", nil)
	w := httptest.NewRecorder()
	mountAutotagRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSuggest_FallbackWhenTaggerFails(t *testing.T) {
	// When the LLM tagger fails, the handler falls back to keyword-based suggestion.
	tagger := &fakeTagger{err: errors.New("llm unavailable")}
	getter := &fakeMissionGetter{m: sampleMission()}
	h := autotag.NewHandler(getter, tagger)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/macbook-pro/suggest-category", nil)
	w := httptest.NewRecorder()
	mountAutotagRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with fallback, got %d: %s", w.Code, w.Body.String())
	}

	var got autotag.CategorySuggestion
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Category == "" {
		t.Errorf("fallback must return a non-empty category")
	}
}

func TestSuggest_EmptyAlternativesNeverNull(t *testing.T) {
	// Alternatives must be an empty array, not null, in JSON.
	sug := &autotag.CategorySuggestion{
		Category:     "Autre",
		Confidence:   0.5,
		Alternatives: nil, // intentionally nil — handler must convert to []
	}
	tagger := &fakeTagger{suggestion: sug}
	getter := &fakeMissionGetter{m: sampleMission()}
	h := autotag.NewHandler(getter, tagger)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/macbook-pro/suggest-category", nil)
	w := httptest.NewRecorder()
	mountAutotagRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Check that JSON output has [] not null for alternatives.
	var raw map[string]any
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	alts, ok := raw["alternatives"]
	if !ok {
		t.Fatal("alternatives key missing from response")
	}
	if alts == nil {
		t.Error("alternatives must not be null in JSON output")
	}
	if arr, ok := alts.([]any); !ok || len(arr) != 0 {
		t.Errorf("expected alternatives to be an empty array, got %v", alts)
	}
}
