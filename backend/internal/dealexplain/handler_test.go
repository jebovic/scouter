package dealexplain_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/dealexplain"
	"github.com/jibei/scouter/internal/shopping"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeItemGetter struct {
	item *shopping.Item
	err  error
}

func (f *fakeItemGetter) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return f.item, f.err
}

type fakeExplainAgent struct {
	result *dealexplain.DealExplanation
	err    error
	called int
}

func (f *fakeExplainAgent) Explain(_ context.Context, _ dealexplain.ExplainInput) (*dealexplain.DealExplanation, error) {
	f.called++
	return f.result, f.err
}

// ── router helper ─────────────────────────────────────────────────────────────

func mountExplainRouter(h *dealexplain.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/shopping-items/{id}/explain", h.Explain)
	return r
}

func sampleItem() *shopping.Item {
	return &shopping.Item{
		Name:     "MacBook Pro 14",
		Merchant: "Apple Store",
		Price:    1999.99,
	}
}

func sampleExplanation(itemID string) *dealexplain.DealExplanation {
	return &dealexplain.DealExplanation{
		ItemID:      itemID,
		Explanation: "Le prix actuel est 15% en dessous de la moyenne historique, ce qui en fait une excellente affaire.",
		Tags:        []string{"prix bas", "tendance baisse"},
		GeneratedAt: time.Now().UTC(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestExplain_HappyPath(t *testing.T) {
	const itemID = "11111111-1111-1111-1111-111111111111"

	agent := &fakeExplainAgent{result: sampleExplanation(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain?score=75&trend=falling", nil)
	w := httptest.NewRecorder()
	mountExplainRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got dealexplain.DealExplanation
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemID != itemID {
		t.Errorf("expected itemId %s, got %s", itemID, got.ItemID)
	}
	if got.Explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
	if len(got.Tags) == 0 {
		t.Errorf("expected at least one tag")
	}
	if got.GeneratedAt.IsZero() {
		t.Errorf("expected non-zero generatedAt")
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestExplain_FallbackWhenAgentFails(t *testing.T) {
	const itemID = "22222222-2222-2222-2222-222222222222"

	// Agent returns an error — handler must produce a deterministic fallback explanation.
	agent := &fakeExplainAgent{err: errLLMDown}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain?score=85&trend=stable", nil)
	w := httptest.NewRecorder()
	mountExplainRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (fallback), got %d: %s", w.Code, w.Body.String())
	}

	var got dealexplain.DealExplanation
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemID != itemID {
		t.Errorf("expected itemId %s, got %s", itemID, got.ItemID)
	}
	if got.Explanation == "" {
		t.Errorf("fallback explanation must not be empty")
	}
	if len(got.Tags) == 0 {
		t.Errorf("fallback tags must not be empty")
	}
}

func TestExplain_CacheHit(t *testing.T) {
	const itemID = "33333333-3333-3333-3333-333333333333"

	agent := &fakeExplainAgent{result: sampleExplanation(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)
	router := mountExplainRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain?score=70&trend=rising", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain?score=70&trend=rising", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called exactly once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestExplain_ItemNotFound(t *testing.T) {
	const itemID = "44444444-4444-4444-4444-444444444444"

	agent := &fakeExplainAgent{}
	getter := &fakeItemGetter{item: nil}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain", nil)
	w := httptest.NewRecorder()
	mountExplainRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when item not found, called %d times", agent.called)
	}
}

func TestExplain_InvalidID(t *testing.T) {
	agent := &fakeExplainAgent{}
	getter := &fakeItemGetter{}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/not-a-uuid/explain", nil)
	w := httptest.NewRecorder()
	mountExplainRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestExplain_ScoreOutOfRange(t *testing.T) {
	const itemID = "55555555-5555-5555-5555-555555555555"

	agent := &fakeExplainAgent{}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := dealexplain.NewCache(1 * time.Hour)
	h := dealexplain.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/explain?score=150", nil)
	w := httptest.NewRecorder()
	mountExplainRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range score, got %d", w.Code)
	}
}

// errLLMDown is a sentinel error used in tests.
var errLLMDown = errSentinel("llm unavailable")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
