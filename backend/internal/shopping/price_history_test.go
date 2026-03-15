package shopping

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
)

// ---------------------------------------------------------------------------
// mock helpers for GetPriceHistory
// ---------------------------------------------------------------------------

// priceHistoryPoints stored in mockRepo keyed by itemID.
var mockPriceHistoryPoints = map[uuid.UUID][]PriceHistoryPoint{}

// Extend mockRepo with GetPriceHistory.
func (m *mockRepo) GetPriceHistory(_ context.Context, itemID uuid.UUID, limit int) ([]PriceHistoryPoint, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	pts := mockPriceHistoryPoints[itemID]
	if limit > 0 && len(pts) > limit {
		pts = pts[len(pts)-limit:]
	}
	return pts, nil
}

// ---------------------------------------------------------------------------
// Service tests
// ---------------------------------------------------------------------------

func TestService_GetPriceHistory_ReturnsEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	pts, err := svc.GetPriceHistory(context.Background(), uuid.New(), 50)
	if err != nil {
		t.Fatalf("GetPriceHistory() unexpected error: %v", err)
	}
	if pts == nil {
		t.Error("GetPriceHistory() = nil; want empty slice")
	}
	if len(pts) != 0 {
		t.Errorf("GetPriceHistory() len = %d; want 0", len(pts))
	}
}

func TestService_GetPriceHistory_ReturnsPoints(t *testing.T) {
	repo := newMockRepo()
	itemID := uuid.New()
	now := time.Now().UTC()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{
		{RecordedAt: now.Add(-2 * time.Hour), Price: 100.0, Source: "manual"},
		{RecordedAt: now.Add(-1 * time.Hour), Price: 90.0, Source: "scheduler"},
		{RecordedAt: now, Price: 85.0, Source: "manual"},
	}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	svc := NewService(repo)
	pts, err := svc.GetPriceHistory(context.Background(), itemID, 50)
	if err != nil {
		t.Fatalf("GetPriceHistory() unexpected error: %v", err)
	}
	if len(pts) != 3 {
		t.Errorf("GetPriceHistory() len = %d; want 3", len(pts))
	}
}

func TestService_GetPriceHistory_RespectsLimit(t *testing.T) {
	repo := newMockRepo()
	itemID := uuid.New()
	now := time.Now().UTC()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{
		{RecordedAt: now.Add(-3 * time.Hour), Price: 110.0, Source: "s"},
		{RecordedAt: now.Add(-2 * time.Hour), Price: 100.0, Source: "s"},
		{RecordedAt: now.Add(-1 * time.Hour), Price: 90.0, Source: "s"},
		{RecordedAt: now, Price: 80.0, Source: "s"},
	}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	svc := NewService(repo)
	pts, err := svc.GetPriceHistory(context.Background(), itemID, 2)
	if err != nil {
		t.Fatalf("GetPriceHistory() unexpected error: %v", err)
	}
	if len(pts) != 2 {
		t.Errorf("GetPriceHistory() len = %d; want 2", len(pts))
	}
}

func TestService_GetPriceHistory_PropagatesRepoError(t *testing.T) {
	repo := newMockRepo()
	repo.nextErr = errors.New("db failure")
	svc := NewService(repo)

	_, err := svc.GetPriceHistory(context.Background(), uuid.New(), 50)
	if err == nil {
		t.Error("GetPriceHistory() should propagate repo error")
	}
}

// ---------------------------------------------------------------------------
// Handler tests
// ---------------------------------------------------------------------------

func newHandlerWithMock() (*Handler, *mockRepo) {
	repo := newMockRepo()
	svc := NewService(repo)
	return NewHandler(svc), repo
}

func routerWithItemID(itemID string, h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})
	return r
}

func TestHandler_PriceHistory_InvalidUUID(t *testing.T) {
	h, _ := newHandlerWithMock()
	router := routerWithItemID("", h)

	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/not-a-uuid/price-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_PriceHistory_DefaultLimit(t *testing.T) {
	h, repo := newHandlerWithMock()
	itemID := uuid.New()
	now := time.Now().UTC()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{
		{RecordedAt: now.Add(-time.Hour), Price: 99.0, Source: "manual"},
	}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })
	_ = repo // referenced to ensure mock is used

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}

	var result []PriceHistoryPoint
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("len(result) = %d; want 1", len(result))
	}
}

func TestHandler_PriceHistory_CustomLimit(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	now := time.Now().UTC()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{
		{RecordedAt: now.Add(-4 * time.Hour), Price: 120, Source: "s"},
		{RecordedAt: now.Add(-3 * time.Hour), Price: 110, Source: "s"},
		{RecordedAt: now.Add(-2 * time.Hour), Price: 100, Source: "s"},
		{RecordedAt: now.Add(-1 * time.Hour), Price: 90, Source: "s"},
		{RecordedAt: now, Price: 80, Source: "s"},
	}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history?limit=3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}

	var result []PriceHistoryPoint
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("len(result) = %d; want 3", len(result))
	}
}

func TestHandler_PriceHistory_LimitClampsToMax(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	router := routerWithItemID(itemID.String(), h)
	// Request limit above max (200) — handler should clamp to 200
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history?limit=9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed (not error) — limit clamping is silent
	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_PriceHistory_InvalidLimitFallsBackToDefault(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history?limit=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_PriceHistory_ZeroLimitFallsBackToDefault(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history?limit=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_PriceHistory_RepoError(t *testing.T) {
	h, repo := newHandlerWithMock()
	repo.nextErr = errors.New("db explosion")

	itemID := uuid.New()
	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_PriceHistory_EmptyResultIsArray(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	// No points registered — expect empty JSON array, not null

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}

	// Body should be a JSON array (not null)
	body := w.Body.String()
	if len(body) == 0 || body[0] != '[' {
		t.Errorf("body = %q; want JSON array starting with '['", body)
	}
}

func TestHandler_PriceHistory_ResponseShape(t *testing.T) {
	h, _ := newHandlerWithMock()
	itemID := uuid.New()
	ts := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	mockPriceHistoryPoints[itemID] = []PriceHistoryPoint{
		{RecordedAt: ts, Price: 149.99, Source: "scheduler"},
	}
	t.Cleanup(func() { delete(mockPriceHistoryPoints, itemID) })

	router := routerWithItemID(itemID.String(), h)
	req := httptest.NewRequest(http.MethodGet, "/missions/"+uuid.New().String()+"/shopping/"+itemID.String()+"/price-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var result []PriceHistoryPoint
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d; want 1", len(result))
	}
	if result[0].Price != 149.99 {
		t.Errorf("Price = %f; want 149.99", result[0].Price)
	}
	if result[0].Source != "scheduler" {
		t.Errorf("Source = %q; want %q", result[0].Source, "scheduler")
	}
}
