package benchmark_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/benchmark"
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

type fakeBenchmarkAgent struct {
	result *benchmark.BenchmarkResult
	err    error
	called int
}

func (f *fakeBenchmarkAgent) Benchmark(_ context.Context, itemID, itemName string, currentPrice float64) (*benchmark.BenchmarkResult, error) {
	f.called++
	return f.result, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func mountRouter(h *benchmark.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/shopping-items/{id}/benchmark", h.Get)
	return r
}

func sampleItem() *shopping.Item {
	return &shopping.Item{
		ID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:     "Sony WH-1000XM5",
		Merchant: "Fnac",
		Price:    279.99,
	}
}

func sampleGoodDeal(itemID string) *benchmark.BenchmarkResult {
	return &benchmark.BenchmarkResult{
		ItemID:       itemID,
		ItemName:     "Sony WH-1000XM5",
		CurrentPrice: 279.99,
		Currency:     "EUR",
		MarketRange: benchmark.PriceRange{
			Low:     299.0,
			Average: 349.0,
			High:    399.0,
		},
		Verdict:     "good_deal",
		DiffPercent: -19.77,
		Explanation: "Ce prix est 20% en dessous de la moyenne du marché français.",
		CachedAt:    time.Now().Unix(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestBenchmark_GoodDeal(t *testing.T) {
	const itemID = "11111111-1111-1111-1111-111111111111"

	agent := &fakeBenchmarkAgent{result: sampleGoodDeal(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := benchmark.NewCache(2 * time.Hour)
	h := benchmark.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got benchmark.BenchmarkResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemID != itemID {
		t.Errorf("expected itemId %s, got %s", itemID, got.ItemID)
	}
	if got.Verdict != "good_deal" {
		t.Errorf("expected verdict good_deal, got %s", got.Verdict)
	}
	if got.DiffPercent >= 0 {
		t.Errorf("expected negative diffPercent for good deal, got %f", got.DiffPercent)
	}
	if got.Explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestBenchmark_Overpriced(t *testing.T) {
	const itemID = "22222222-2222-2222-2222-222222222222"

	item := &shopping.Item{
		ID:       uuid.MustParse(itemID),
		Name:     "Casque Audio Generic",
		Merchant: "Amazon",
		Price:    450.0,
	}
	result := &benchmark.BenchmarkResult{
		ItemID:       itemID,
		ItemName:     "Casque Audio Generic",
		CurrentPrice: 450.0,
		Currency:     "EUR",
		MarketRange: benchmark.PriceRange{
			Low:     280.0,
			Average: 350.0,
			High:    420.0,
		},
		Verdict:     "overpriced",
		DiffPercent: 28.57,
		Explanation: "Ce prix est 29% au-dessus de la moyenne du marché français.",
		CachedAt:    time.Now().Unix(),
	}

	agent := &fakeBenchmarkAgent{result: result}
	getter := &fakeItemGetter{item: item}
	cache := benchmark.NewCache(2 * time.Hour)
	h := benchmark.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got benchmark.BenchmarkResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Verdict != "overpriced" {
		t.Errorf("expected verdict overpriced, got %s", got.Verdict)
	}
	if got.DiffPercent <= 10 {
		t.Errorf("expected diffPercent > 10 for overpriced, got %f", got.DiffPercent)
	}
}

func TestBenchmark_Average(t *testing.T) {
	const itemID = "33333333-3333-3333-3333-333333333333"

	item := &shopping.Item{
		ID:       uuid.MustParse(itemID),
		Name:     "Clavier Mécanique",
		Merchant: "Darty",
		Price:    120.0,
	}
	result := &benchmark.BenchmarkResult{
		ItemID:       itemID,
		ItemName:     "Clavier Mécanique",
		CurrentPrice: 120.0,
		Currency:     "EUR",
		MarketRange: benchmark.PriceRange{
			Low:     90.0,
			Average: 115.0,
			High:    150.0,
		},
		Verdict:     "average",
		DiffPercent: 4.35,
		Explanation: "Ce prix est dans la moyenne du marché français pour ce type de produit.",
		CachedAt:    time.Now().Unix(),
	}

	agent := &fakeBenchmarkAgent{result: result}
	getter := &fakeItemGetter{item: item}
	cache := benchmark.NewCache(2 * time.Hour)
	h := benchmark.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got benchmark.BenchmarkResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Verdict != "average" {
		t.Errorf("expected verdict average, got %s", got.Verdict)
	}
	if got.DiffPercent < -10 || got.DiffPercent > 10 {
		t.Errorf("expected diffPercent in [-10, 10] for average, got %f", got.DiffPercent)
	}
}

func TestBenchmark_NotFound(t *testing.T) {
	const itemID = "44444444-4444-4444-4444-444444444444"

	agent := &fakeBenchmarkAgent{}
	getter := &fakeItemGetter{item: nil}
	cache := benchmark.NewCache(2 * time.Hour)
	h := benchmark.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when item not found, called %d times", agent.called)
	}
}

func TestBenchmark_CachedResponse(t *testing.T) {
	const itemID = "55555555-5555-5555-5555-555555555555"

	agent := &fakeBenchmarkAgent{result: sampleGoodDeal(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := benchmark.NewCache(2 * time.Hour)
	h := benchmark.NewHandler(getter, agent, cache)
	router := mountRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/benchmark", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

// errSentinel is a typed test error.
var errAgentDown = errors.New("agent unavailable")
