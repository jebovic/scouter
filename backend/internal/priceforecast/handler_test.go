package priceforecast_test

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
	"github.com/jibei/scouter/internal/priceforecast"
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

type fakePriceHistoryGetter struct {
	pts []shopping.PriceHistoryPoint
	err error
}

func (f *fakePriceHistoryGetter) GetPriceHistory(_ context.Context, _ uuid.UUID, _ int) ([]shopping.PriceHistoryPoint, error) {
	return f.pts, f.err
}

type fakeForecastAgent struct {
	result *priceforecast.PriceForecastResult
	err    error
	called int
}

func (f *fakeForecastAgent) Forecast(_ context.Context, req priceforecast.ForecastRequest) (*priceforecast.PriceForecastResult, error) {
	f.called++
	return f.result, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func mountRouter(h *priceforecast.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/missions/{missionID}/shopping/{itemID}/forecast", h.GetForecast)
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

func sampleHistory() []shopping.PriceHistoryPoint {
	base := time.Now().Add(-30 * 24 * time.Hour)
	return []shopping.PriceHistoryPoint{
		{RecordedAt: base, Price: 320.0, Source: "manual"},
		{RecordedAt: base.Add(10 * 24 * time.Hour), Price: 299.0, Source: "manual"},
		{RecordedAt: base.Add(20 * 24 * time.Hour), Price: 279.99, Source: "manual"},
	}
}

func sampleForecast(itemID string) *priceforecast.PriceForecastResult {
	now := time.Now()
	return &priceforecast.PriceForecastResult{
		ItemID:   itemID,
		ItemName: "Sony WH-1000XM5",
		Currency: "EUR",
		Historical: []priceforecast.DataPoint{
			{Date: now.Add(-30 * 24 * time.Hour), Price: 320.0},
		},
		Forecast: []priceforecast.ForecastPoint{
			{Date: now.Add(7 * 24 * time.Hour), Low: 260.0, Mid: 270.0, High: 285.0},
			{Date: now.Add(14 * 24 * time.Hour), Low: 255.0, Mid: 265.0, High: 280.0},
		},
		Confidence:    0.75,
		Summary:       "Les prix devraient baisser dans les prochaines semaines.",
		BestBuyWindow: "dans 2-3 semaines",
		GeneratedAt:   now,
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetForecast_Success(t *testing.T) {
	const itemID = "11111111-1111-1111-1111-111111111111"
	const missionID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	agent := &fakeForecastAgent{result: sampleForecast(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	histGetter := &fakePriceHistoryGetter{pts: sampleHistory()}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got priceforecast.PriceForecastResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemID != itemID {
		t.Errorf("expected itemId %s, got %s", itemID, got.ItemID)
	}
	if len(got.Forecast) == 0 {
		t.Errorf("expected non-empty forecast points")
	}
	if got.Confidence <= 0 || got.Confidence > 1 {
		t.Errorf("expected confidence in (0,1], got %f", got.Confidence)
	}
	if got.Summary == "" {
		t.Errorf("expected non-empty summary")
	}
	if got.BestBuyWindow == "" {
		t.Errorf("expected non-empty best_buy_window")
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

func TestGetForecast_CacheHit(t *testing.T) {
	const itemID = "22222222-2222-2222-2222-222222222222"
	const missionID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	agent := &fakeForecastAgent{result: sampleForecast(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	histGetter := &fakePriceHistoryGetter{pts: sampleHistory()}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)
	router := mountRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d: %s", w1.Code, w1.Body.String())
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetForecast_ItemNotFound(t *testing.T) {
	const itemID = "33333333-3333-3333-3333-333333333333"
	const missionID = "cccccccc-cccc-cccc-cccc-cccccccccccc"

	agent := &fakeForecastAgent{}
	getter := &fakeItemGetter{item: nil}
	histGetter := &fakePriceHistoryGetter{}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when item not found, called %d times", agent.called)
	}
}

func TestGetForecast_InvalidItemID(t *testing.T) {
	const missionID = "dddddddd-dddd-dddd-dddd-dddddddddddd"

	agent := &fakeForecastAgent{}
	getter := &fakeItemGetter{}
	histGetter := &fakePriceHistoryGetter{}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/not-a-uuid/forecast", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetForecast_AgentError(t *testing.T) {
	const itemID = "44444444-4444-4444-4444-444444444444"
	const missionID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"

	agent := &fakeForecastAgent{err: errors.New("llm unavailable")}
	getter := &fakeItemGetter{item: sampleItem()}
	histGetter := &fakePriceHistoryGetter{pts: sampleHistory()}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetForecast_HistoryError(t *testing.T) {
	const itemID = "55555555-5555-5555-5555-555555555555"
	const missionID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	agent := &fakeForecastAgent{}
	getter := &fakeItemGetter{item: sampleItem()}
	histGetter := &fakePriceHistoryGetter{err: errors.New("db error")}
	cache := priceforecast.NewCache(2 * time.Hour)
	h := priceforecast.NewHandler(agent, getter, histGetter, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID+"/shopping/"+itemID+"/forecast", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCache_GetSet(t *testing.T) {
	c := priceforecast.NewCache(2 * time.Hour)

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected cache miss for key1")
	}

	now := time.Now()
	result := &priceforecast.PriceForecastResult{
		ItemID:      "key1",
		GeneratedAt: now,
	}
	c.Set("key1", result)

	got, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit for key1")
	}
	if got.ItemID != "key1" {
		t.Errorf("expected itemId key1, got %s", got.ItemID)
	}
}

func TestCache_Expiry(t *testing.T) {
	c := priceforecast.NewCache(1 * time.Millisecond)

	result := &priceforecast.PriceForecastResult{ItemID: "exp"}
	c.Set("exp", result)

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("exp")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}
