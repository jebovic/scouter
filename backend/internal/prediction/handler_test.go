package prediction_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/prediction"
	"github.com/jibei/scouter/internal/shopping"
)

// ── stub repository ──────────────────────────────────────────────────────────

type stubShoppingRepo struct {
	item     *shopping.Item
	history  []shopping.PriceSnapshot
	itemErr  error
	histErr  error
}

func (s *stubShoppingRepo) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return s.item, s.itemErr
}

func (s *stubShoppingRepo) ListPriceHistory(_ context.Context, _ uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return s.history, s.histErr
}

// ── helpers ──────────────────────────────────────────────────────────────────

func makeItem(id uuid.UUID, price float64) *shopping.Item {
	return &shopping.Item{
		ID:        id,
		MissionID: uuid.New(),
		Name:      "Test Item",
		Price:     price,
		Status:    "watch",
		CreatedAt: time.Now(),
	}
}

func makeSnapshots(prices []float64) []shopping.PriceSnapshot {
	snaps := make([]shopping.PriceSnapshot, len(prices))
	base := time.Now().Add(-time.Duration(len(prices)) * 24 * time.Hour)
	for i, p := range prices {
		snaps[i] = shopping.PriceSnapshot{
			ID:         uuid.New(),
			ItemID:     uuid.New(),
			Price:      p,
			RecordedAt: base.Add(time.Duration(i) * 24 * time.Hour),
		}
	}
	return snaps
}

func buildRouter(repo prediction.ShoppingRepository) http.Handler {
	h := prediction.NewHandler(repo)
	r := chi.NewRouter()
	r.Get("/api/shopping-items/{id}/prediction", h.GetPrediction)
	return r
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestGetPrediction_OK(t *testing.T) {
	itemID := uuid.New()
	repo := &stubShoppingRepo{
		item:    makeItem(itemID, 150),
		history: makeSnapshots([]float64{100, 110, 120, 130, 140}),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID.String()+"/prediction", nil)
	w := httptest.NewRecorder()

	buildRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var got prediction.PricePrediction
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.ItemID == "" {
		t.Error("itemId must not be empty")
	}
	if got.ForecastDays != 30 {
		t.Errorf("forecastDays: want 30, got %d", got.ForecastDays)
	}
	if got.DataPoints != 5 {
		t.Errorf("dataPoints: want 5, got %d", got.DataPoints)
	}
	if got.ConfidenceLevel != "high" {
		t.Errorf("confidenceLevel: want high, got %s", got.ConfidenceLevel)
	}
	if got.Trend != "rising" {
		t.Errorf("trend: want rising, got %s", got.Trend)
	}
}

func TestGetPrediction_LessThanTwoPoints(t *testing.T) {
	itemID := uuid.New()
	repo := &stubShoppingRepo{
		item:    makeItem(itemID, 99),
		history: makeSnapshots([]float64{99}), // single point
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID.String()+"/prediction", nil)
	w := httptest.NewRecorder()

	buildRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var got prediction.PricePrediction
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.ConfidenceLevel != "low" {
		t.Errorf("want low confidence, got %s", got.ConfidenceLevel)
	}
	if got.Trend != "stable" {
		t.Errorf("want stable trend, got %s", got.Trend)
	}
}

func TestGetPrediction_NoHistory(t *testing.T) {
	itemID := uuid.New()
	repo := &stubShoppingRepo{
		item:    makeItem(itemID, 75),
		history: nil, // no history at all
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID.String()+"/prediction", nil)
	w := httptest.NewRecorder()

	buildRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var got prediction.PricePrediction
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.CurrentPrice != 75 {
		t.Errorf("currentPrice: want 75, got %f", got.CurrentPrice)
	}
	if got.ConfidenceLevel != "low" {
		t.Errorf("want low confidence, got %s", got.ConfidenceLevel)
	}
	if got.Trend != "stable" {
		t.Errorf("want stable trend, got %s", got.Trend)
	}
}

func TestGetPrediction_InvalidID(t *testing.T) {
	repo := &stubShoppingRepo{}

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/not-a-uuid/prediction", nil)
	w := httptest.NewRecorder()

	buildRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestGetPrediction_ItemNotFound(t *testing.T) {
	itemID := uuid.New()
	repo := &stubShoppingRepo{
		item: nil, // not found
	}

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID.String()+"/prediction", nil)
	w := httptest.NewRecorder()

	buildRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}
