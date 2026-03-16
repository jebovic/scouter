package pricecomp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/pricecomp"
)

// ── fakes ────────────────────────────────────────────────────────────────────

// fakeComparator implements pricecomp.Comparator so tests do not depend on an LLM.
type fakeComparator struct {
	comparison *pricecomp.PriceComparison
	err        error
	called     int
}

func (f *fakeComparator) Compare(_ context.Context, productName, optionID string) (*pricecomp.PriceComparison, error) {
	f.called++
	return f.comparison, f.err
}

// fakeNameGetter implements pricecomp.OptionNameGetter.
type fakeNameGetter struct {
	name string
	err  error
}

func (f *fakeNameGetter) GetOptionName(_ context.Context, optionID string) (string, error) {
	return f.name, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func goodComparison(optionID string) *pricecomp.PriceComparison {
	now := time.Now().UTC().Format(time.RFC3339)
	return &pricecomp.PriceComparison{
		OptionID: optionID,
		Product:  "Sony WH-1000XM5",
		Offers: []pricecomp.RetailerOffer{
			{Retailer: "Fnac", Price: 299.99, Currency: "EUR", Available: true, URL: "https://fnac.com/test", UpdatedAt: now},
			{Retailer: "Amazon FR", Price: 319.99, Currency: "EUR", Available: true, URL: "https://amazon.fr/test", UpdatedAt: now},
			{Retailer: "Darty", Price: 349.99, Currency: "EUR", Available: false, URL: "", UpdatedAt: now},
		},
		LowestIdx: 0,
		FetchedAt: now,
	}
}

func mountPriceCompRouter(h *pricecomp.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/options/{id}/price-comparison", h.GetComparison)
	r.Post("/api/options/{id}/price-comparison/refresh", h.RefreshComparison)
	return r
}

// ── GET /api/options/{id}/price-comparison ───────────────────────────────────

func TestGetComparison_HappyPath(t *testing.T) {
	const optID = "11111111-1111-1111-1111-111111111111"
	agent := &fakeComparator{comparison: goodComparison(optID)}
	getter := &fakeNameGetter{name: "Sony WH-1000XM5"}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got pricecomp.PriceComparison
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.OptionID != optID {
		t.Errorf("expected optionId %s, got %s", optID, got.OptionID)
	}
	if got.Product != "Sony WH-1000XM5" {
		t.Errorf("expected product Sony WH-1000XM5, got %s", got.Product)
	}
	if len(got.Offers) != 3 {
		t.Errorf("expected 3 offers, got %d", len(got.Offers))
	}
	if got.LowestIdx != 0 {
		t.Errorf("expected lowestIdx 0, got %d", got.LowestIdx)
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetComparison_CacheHit(t *testing.T) {
	const optID = "22222222-2222-2222-2222-222222222222"
	agent := &fakeComparator{comparison: goodComparison(optID)}
	getter := &fakeNameGetter{name: "Apple AirPods Pro"}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)
	router := mountPriceCompRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache; agent not called again.
	r2 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called exactly once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetComparison_AgentError(t *testing.T) {
	const optID = "33333333-3333-3333-3333-333333333333"
	agent := &fakeComparator{err: errors.New("llm unavailable")}
	getter := &fakeNameGetter{name: "Bose QC45"}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded, got %d: %s", w.Code, w.Body.String())
	}
	var got pricecomp.PriceComparison
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode degraded response: %v", err)
	}
	if got.OptionID != optID {
		t.Errorf("expected optionId %s, got %s", optID, got.OptionID)
	}
	if len(got.Offers) != 0 {
		t.Errorf("expected empty offers in degraded response, got %d", len(got.Offers))
	}
}

func TestGetComparison_OptionNotFound(t *testing.T) {
	const optID = "44444444-4444-4444-4444-444444444444"
	agent := &fakeComparator{}
	// getter returns empty name → option does not exist
	getter := &fakeNameGetter{name: ""}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when option not found, called %d times", agent.called)
	}
}

func TestGetComparison_NameGetterError(t *testing.T) {
	const optID = "55555555-5555-5555-5555-555555555555"
	agent := &fakeComparator{}
	getter := &fakeNameGetter{err: errors.New("db error")}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when getter fails, called %d times", agent.called)
	}
}

// ── POST /api/options/{id}/price-comparison/refresh ──────────────────────────

func TestRefreshComparison_ClearsCacheAndRecalls(t *testing.T) {
	const optID = "66666666-6666-6666-6666-666666666666"
	now := time.Now().UTC().Format(time.RFC3339)

	first := goodComparison(optID)
	first.Offers[0].Price = 280.00

	second := goodComparison(optID)
	second.Offers[0].Price = 260.00

	agent := &fakeComparator{comparison: first}
	getter := &fakeNameGetter{name: "Samsung Galaxy Buds2 Pro"}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)
	router := mountPriceCompRouter(h)

	// Populate cache via GET.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/price-comparison", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET failed: %d", w1.Code)
	}

	// Change agent response for next call.
	agent.comparison = second

	// POST refresh — must clear cache and call agent again.
	r2 := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/price-comparison/refresh", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST refresh failed: %d: %s", w2.Code, w2.Body.String())
	}

	var got pricecomp.PriceComparison
	if err := json.NewDecoder(w2.Body).Decode(&got); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	if got.Offers[0].Price != 260.00 {
		t.Errorf("expected refreshed price 260.00, got %f", got.Offers[0].Price)
	}
	if agent.called != 2 {
		t.Errorf("expected agent called twice (GET + refresh), got %d", agent.called)
	}
	_ = now
}

func TestRefreshComparison_AgentError(t *testing.T) {
	const optID = "77777777-7777-7777-7777-777777777777"
	agent := &fakeComparator{err: errors.New("timeout")}
	getter := &fakeNameGetter{name: "Jabra Evolve2 85"}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/price-comparison/refresh", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded, got %d: %s", w.Code, w.Body.String())
	}
	var got pricecomp.PriceComparison
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode degraded response: %v", err)
	}
	if len(got.Offers) != 0 {
		t.Errorf("expected empty offers in degraded response, got %d", len(got.Offers))
	}
}

func TestRefreshComparison_OptionNotFound(t *testing.T) {
	const optID = "88888888-8888-8888-8888-888888888888"
	agent := &fakeComparator{}
	getter := &fakeNameGetter{name: ""}
	cache := pricecomp.NewCache(30 * time.Minute)
	h := pricecomp.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID+"/price-comparison/refresh", nil)
	w := httptest.NewRecorder()
	mountPriceCompRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ── cache unit tests ──────────────────────────────────────────────────────────

func TestCache_SetAndGet(t *testing.T) {
	c := pricecomp.NewCache(time.Hour)
	comp := goodComparison("abc")
	c.Set("abc", comp)

	got, ok := c.Get("abc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.OptionID != comp.OptionID {
		t.Errorf("expected %s, got %s", comp.OptionID, got.OptionID)
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	c := pricecomp.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := pricecomp.NewCache(1 * time.Millisecond)
	c.Set("key", goodComparison("key"))
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := pricecomp.NewCache(time.Hour)
	c.Set("key", goodComparison("key"))
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after Delete")
	}
}
