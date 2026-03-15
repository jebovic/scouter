package digest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/digest"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

type stubRepository struct {
	items []digest.PriceDropItem
	err   error
}

func (s *stubRepository) GetPriceDrops(ctx context.Context, days int) ([]digest.PriceDropItem, error) {
	return s.items, s.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mountRouter(h *digest.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/digest/weekly", h.Weekly)
	return r
}

func decodeDigest(t *testing.T, body *httptest.ResponseRecorder) digest.WeeklyDigest {
	t.Helper()
	var d digest.WeeklyDigest
	if err := json.NewDecoder(body.Body).Decode(&d); err != nil {
		t.Fatalf("decode WeeklyDigest: %v", err)
	}
	return d
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestWeekly_EmptyResult_Returns200WithZeroDrops(t *testing.T) {
	repo := &stubRepository{items: []digest.PriceDropItem{}}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	d := decodeDigest(t, w)
	if d.TotalDrops != 0 {
		t.Errorf("expected 0 drops, got %d", d.TotalDrops)
	}
	if d.TotalSavings != 0 {
		t.Errorf("expected 0 savings, got %f", d.TotalSavings)
	}
	if d.PeriodDays != 7 {
		t.Errorf("expected default period 7, got %d", d.PeriodDays)
	}
	if d.Items == nil {
		t.Error("expected non-nil Items slice")
	}
}

func TestWeekly_WithDrops_ReturnsCorrectCalculations(t *testing.T) {
	items := []digest.PriceDropItem{
		{
			MissionID:   "m1",
			MissionName: "Laptop Mission",
			MissionSlug: "laptop-mission",
			ItemID:      "i1",
			ItemName:    "Dell XPS 15",
			Merchant:    "Amazon",
			PriceBefore: 1200.0,
			PriceNow:    960.0,
			Currency:    "EUR",
		},
		{
			MissionID:   "m1",
			MissionName: "Laptop Mission",
			MissionSlug: "laptop-mission",
			ItemID:      "i2",
			ItemName:    "Logitech MX Keys",
			Merchant:    "Fnac",
			PriceBefore: 100.0,
			PriceNow:    80.0,
			Currency:    "EUR",
		},
	}
	repo := &stubRepository{items: items}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	d := decodeDigest(t, w)

	if d.TotalDrops != 2 {
		t.Errorf("expected 2 drops, got %d", d.TotalDrops)
	}

	// Total savings: (1200-960) + (100-80) = 240 + 20 = 260
	expectedSavings := 260.0
	if d.TotalSavings != expectedSavings {
		t.Errorf("expected savings %.2f, got %.2f", expectedSavings, d.TotalSavings)
	}

	// dropPct for item 1: (1200 - 960) / 1200 * 100 = 20.0
	if len(d.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(d.Items))
	}
	item1 := d.Items[0]
	expectedDrop := 20.0
	if item1.DropPct != expectedDrop {
		t.Errorf("expected dropPct %.2f, got %.2f", expectedDrop, item1.DropPct)
	}
}

func TestWeekly_InvalidDaysParam_Returns400(t *testing.T) {
	repo := &stubRepository{}
	h := digest.NewHandler(repo)

	for _, tc := range []struct {
		name string
		days string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too_large", "31"},
		{"not_a_number", "abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly?days="+tc.days, nil)
			w := httptest.NewRecorder()
			mountRouter(h).ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for days=%q, got %d", tc.days, w.Code)
			}
		})
	}
}

func TestWeekly_ValidDaysParam_IsRespected(t *testing.T) {
	repo := &stubRepository{items: []digest.PriceDropItem{}}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly?days=14", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	d := decodeDigest(t, w)
	if d.PeriodDays != 14 {
		t.Errorf("expected periodDays=14, got %d", d.PeriodDays)
	}
}

func TestWeekly_RepoError_Returns500(t *testing.T) {
	repo := &stubRepository{err: errors.New("db down")}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestWeekly_ResponseStructure_HasAllFields(t *testing.T) {
	items := []digest.PriceDropItem{
		{
			MissionID:   "m1",
			MissionName: "Test Mission",
			MissionSlug: "test-mission",
			ItemID:      "i1",
			ItemName:    "Test Item",
			Merchant:    "Amazon",
			PriceBefore: 200.0,
			PriceNow:    150.0,
			Currency:    "EUR",
		},
	}
	repo := &stubRepository{items: items}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var raw map[string]any
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	requiredFields := []string{"generatedAt", "periodDays", "totalDrops", "totalSavings", "currency", "items"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing field %q in response", field)
		}
	}
}

func TestWeekly_GeneratedAtIsRecent(t *testing.T) {
	repo := &stubRepository{items: []digest.PriceDropItem{}}
	h := digest.NewHandler(repo)

	before := time.Now().Add(-2 * time.Second)
	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)
	after := time.Now().Add(2 * time.Second)

	d := decodeDigest(t, w)
	if d.GeneratedAt.Before(before) || d.GeneratedAt.After(after) {
		t.Errorf("generatedAt %v not within expected range [%v, %v]", d.GeneratedAt, before, after)
	}
}

func TestWeekly_DropPctCalculation_Precise(t *testing.T) {
	// 300 -> 225 = 25% drop
	items := []digest.PriceDropItem{
		{
			MissionID:   "m1",
			MissionName: "M",
			MissionSlug: "m",
			ItemID:      "i1",
			ItemName:    "Widget",
			Merchant:    "Store",
			PriceBefore: 300.0,
			PriceNow:    225.0,
			Currency:    "EUR",
		},
	}
	repo := &stubRepository{items: items}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	d := decodeDigest(t, w)
	if len(d.Items) == 0 {
		t.Fatal("expected at least one item")
	}
	got := d.Items[0].DropPct
	expected := 25.0
	if got != expected {
		t.Errorf("expected dropPct=%.2f, got=%.2f", expected, got)
	}
}

func TestWeekly_CurrencyQueryParam_IsRespected(t *testing.T) {
	repo := &stubRepository{items: []digest.PriceDropItem{}}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly?currency=USD", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	d := decodeDigest(t, w)
	if d.Currency != "USD" {
		t.Errorf("expected currency=USD, got=%q", d.Currency)
	}
}

func TestWeekly_DefaultCurrencyIsEUR(t *testing.T) {
	repo := &stubRepository{items: []digest.PriceDropItem{}}
	h := digest.NewHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/digest/weekly", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	d := decodeDigest(t, w)
	if d.Currency != "EUR" {
		t.Errorf("expected default currency=EUR, got=%q", d.Currency)
	}
}
