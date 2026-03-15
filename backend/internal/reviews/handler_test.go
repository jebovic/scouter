package reviews_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/reviews"
)

// ── fakes ────────────────────────────────────────────────────────────────────

// fakeReviewer implements reviews.Reviewer so tests do not depend on an LLM or DB.
type fakeReviewer struct {
	summary *reviews.ReviewSummary
	err     error
	called  int
}

func (f *fakeReviewer) Summarise(_ context.Context, optionName string) (*reviews.ReviewSummary, error) {
	f.called++
	return f.summary, f.err
}

// fakeNameGetter implements reviews.OptionNameGetter.
type fakeNameGetter struct {
	name string
	err  error
}

func (f *fakeNameGetter) GetOptionName(_ context.Context, optionID string) (string, error) {
	return f.name, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func goodSummary(optionID string) *reviews.ReviewSummary {
	return &reviews.ReviewSummary{
		OptionID:     optionID,
		Sentiment:    0.72,
		PositivePct:  78,
		NeutralPct:   14,
		NegativePct:  8,
		Pros:         []string{"great battery", "fast charging"},
		Cons:         []string{"expensive"},
		TotalReviews: 1240,
		Source:       "llm-simulated",
		RefreshedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func mountReviewsRouter(h *reviews.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/options/{id}/reviews", h.GetReviews)
	r.Post("/api/options/{id}/reviews/refresh", h.RefreshReviews)
	return r
}

// ── GET /api/options/{id}/reviews ────────────────────────────────────────────

func TestGetReviews_HappyPath(t *testing.T) {
	const optID = "11111111-1111-1111-1111-111111111111"
	agent := &fakeReviewer{summary: goodSummary(optID)}
	getter := &fakeNameGetter{name: "Sony WH-1000XM5"}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got reviews.ReviewSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.OptionID != optID {
		t.Errorf("expected optionId %s, got %s", optID, got.OptionID)
	}
	if got.Sentiment != 0.72 {
		t.Errorf("expected sentiment 0.72, got %f", got.Sentiment)
	}
	if len(got.Pros) != 2 {
		t.Errorf("expected 2 pros, got %d", len(got.Pros))
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetReviews_CacheHit(t *testing.T) {
	const optID = "22222222-2222-2222-2222-222222222222"
	agent := &fakeReviewer{summary: goodSummary(optID)}
	getter := &fakeNameGetter{name: "Apple AirPods Pro"}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)
	router := mountReviewsRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache; agent not called again.
	r2 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called exactly once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetReviews_AgentError(t *testing.T) {
	const optID = "33333333-3333-3333-3333-333333333333"
	agent := &fakeReviewer{err: errors.New("llm unavailable")}
	getter := &fakeNameGetter{name: "Bose QC45"}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetReviews_OptionNotFound(t *testing.T) {
	const optID = "44444444-4444-4444-4444-444444444444"
	agent := &fakeReviewer{}
	// getter returns empty name → option does not exist
	getter := &fakeNameGetter{name: ""}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when option not found, called %d times", agent.called)
	}
}

func TestGetReviews_NameGetterError(t *testing.T) {
	const optID = "55555555-5555-5555-5555-555555555555"
	agent := &fakeReviewer{}
	getter := &fakeNameGetter{err: errors.New("db error")}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when getter fails, called %d times", agent.called)
	}
}

// ── POST /api/options/{id}/reviews/refresh ────────────────────────────────────

func TestRefreshReviews_ClearsCacheAndRecalls(t *testing.T) {
	const optID = "66666666-6666-6666-6666-666666666666"
	first := goodSummary(optID)
	first.Sentiment = 0.50

	second := goodSummary(optID)
	second.Sentiment = 0.90

	callCount := 0
	agent := &fakeReviewer{}
	agent.summary = first
	getter := &fakeNameGetter{name: "Samsung Galaxy Buds2 Pro"}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)
	router := mountReviewsRouter(h)

	// Populate cache via GET.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/reviews", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET failed: %d", w1.Code)
	}
	callCount++ // agent.called == 1

	// Change agent response for next call.
	agent.summary = second

	// POST refresh — must clear cache and call agent again.
	r2 := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/reviews/refresh", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST refresh failed: %d: %s", w2.Code, w2.Body.String())
	}

	var got reviews.ReviewSummary
	if err := json.NewDecoder(w2.Body).Decode(&got); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	if got.Sentiment != 0.90 {
		t.Errorf("expected refreshed sentiment 0.90, got %f", got.Sentiment)
	}
	if agent.called != 2 {
		t.Errorf("expected agent called twice (GET + refresh), got %d", agent.called)
	}
	_ = callCount
}

func TestRefreshReviews_AgentError(t *testing.T) {
	const optID = "77777777-7777-7777-7777-777777777777"
	agent := &fakeReviewer{err: errors.New("timeout")}
	getter := &fakeNameGetter{name: "Jabra Evolve2 85"}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/reviews/refresh", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRefreshReviews_OptionNotFound(t *testing.T) {
	const optID = "88888888-8888-8888-8888-888888888888"
	agent := &fakeReviewer{}
	getter := &fakeNameGetter{name: ""}
	cache := reviews.NewCache(6 * time.Hour)
	h := reviews.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/reviews/refresh", nil)
	w := httptest.NewRecorder()
	mountReviewsRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ── cache unit tests ─────────────────────────────────────────────────────────

func TestCache_SetAndGet(t *testing.T) {
	c := reviews.NewCache(time.Hour)
	sum := goodSummary("abc")
	c.Set("abc", sum)

	got, ok := c.Get("abc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.OptionID != sum.OptionID {
		t.Errorf("expected %s, got %s", sum.OptionID, got.OptionID)
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	c := reviews.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := reviews.NewCache(1 * time.Millisecond)
	c.Set("key", goodSummary("key"))
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := reviews.NewCache(time.Hour)
	c.Set("key", goodSummary("key"))
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after Delete")
	}
}
