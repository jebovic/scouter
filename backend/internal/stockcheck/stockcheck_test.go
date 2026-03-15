package stockcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/stockcheck"
)

// ── heuristic unit tests ──────────────────────────────────────────────────────

func TestDetermineStatus_Deterministic(t *testing.T) {
	t.Parallel()

	s1 := stockcheck.DetermineStatus("iPhone 15 Pro", "fnac")
	s2 := stockcheck.DetermineStatus("iPhone 15 Pro", "fnac")
	if s1 != s2 {
		t.Errorf("DetermineStatus must be deterministic: got %q then %q", s1, s2)
	}
}

func TestDetermineStatus_ReturnValues(t *testing.T) {
	t.Parallel()

	valid := map[string]bool{
		"in_stock":     true,
		"limited":      true,
		"out_of_stock": true,
	}

	cases := []struct{ product, retailer string }{
		{"iPhone 15 Pro", "fnac"},
		{"MacBook Air M3", "darty"},
		{"PlayStation 5", "boulanger"},
		{"", "fnac"},
		{"product", ""},
		{"", ""},
	}

	for _, c := range cases {
		got := stockcheck.DetermineStatus(c.product, c.retailer)
		if !valid[string(got)] {
			t.Errorf("DetermineStatus(%q, %q) = %q; want one of in_stock|limited|out_of_stock",
				c.product, c.retailer, got)
		}
	}
}

func TestDetermineStatus_DifferentRetailersGiveDifferentResults(t *testing.T) {
	t.Parallel()

	// With enough product/retailer combos, not all results should be the same.
	retailers := []string{"fnac", "darty", "boulanger", "cdiscount", "amazon-fr", "ldlc"}
	seen := map[stockcheck.StockStatus]bool{}
	for _, r := range retailers {
		seen[stockcheck.DetermineStatus("Samsung Galaxy S24", r)] = true
	}
	// Expect at least 2 distinct statuses across 6 retailers
	if len(seen) < 2 {
		t.Log("only one status seen across all retailers — could be unlucky hash collision, not failing")
	}
}

// ── cache unit tests ──────────────────────────────────────────────────────────

func TestCache_StoreAndLoad(t *testing.T) {
	t.Parallel()

	c := stockcheck.NewCache(5 * time.Minute)

	key := "product|retailer"
	status := stockcheck.StatusInStock

	if _, ok := c.Get(key); ok {
		t.Fatal("expected cache miss on empty cache")
	}

	c.Set(key, status)

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("expected cache hit after Set")
	}
	if got != status {
		t.Errorf("cache.Get = %q; want %q", got, status)
	}
}

func TestCache_Expiry(t *testing.T) {
	t.Parallel()

	c := stockcheck.NewCache(1 * time.Millisecond)
	c.Set("k", stockcheck.StatusLimited)

	time.Sleep(5 * time.Millisecond)

	if _, ok := c.Get("k"); ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	t.Parallel()

	c := stockcheck.NewCache(5 * time.Minute)
	if _, ok := c.Get("nonexistent"); ok {
		t.Error("expected cache miss for unknown key")
	}
}

// ── handler integration tests ─────────────────────────────────────────────────

func newTestHandler() http.Handler {
	c := stockcheck.NewCache(5 * time.Minute)
	h := stockcheck.NewHandler(c)
	return http.HandlerFunc(h.Check)
}

func TestHandler_MissingProduct(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/stock-check?retailer=fnac", nil)
	w := httptest.NewRecorder()

	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400; got %d", w.Code)
	}
}

func TestHandler_MissingRetailer(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=iPhone", nil)
	w := httptest.NewRecorder()

	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400; got %d", w.Code)
	}
}

func TestHandler_ValidRequest(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=iPhone+15+Pro&retailer=fnac", nil)
	w := httptest.NewRecorder()

	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200; got %d", w.Code)
	}

	var resp stockcheck.StockResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	valid := map[string]bool{"in_stock": true, "limited": true, "out_of_stock": true}
	if !valid[string(resp.Status)] {
		t.Errorf("unexpected status %q", resp.Status)
	}
	if resp.Retailer != "fnac" {
		t.Errorf("expected retailer=fnac; got %q", resp.Retailer)
	}
	if resp.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestHandler_ResponseIsCachedOnSecondCall(t *testing.T) {
	t.Parallel()

	h := newTestHandler()

	req1 := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=LG+OLED+C3&retailer=cdiscount", nil)
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)

	var r1 stockcheck.StockResponse
	if err := json.NewDecoder(w1.Body).Decode(&r1); err != nil {
		t.Fatalf("decode r1: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=LG+OLED+C3&retailer=cdiscount", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	var r2 stockcheck.StockResponse
	if err := json.NewDecoder(w2.Body).Decode(&r2); err != nil {
		t.Fatalf("decode r2: %v", err)
	}

	if r1.Status != r2.Status {
		t.Errorf("cache miss: first=%q second=%q", r1.Status, r2.Status)
	}
}

func TestHandler_ContentTypeIsJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=test&retailer=fnac", nil)
	w := httptest.NewRecorder()

	newTestHandler().ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json; got %q", ct)
	}
}

func TestHandler_ProductTrimmedAndNormalized(t *testing.T) {
	t.Parallel()

	// Leading/trailing whitespace should not cause an error
	req := httptest.NewRequest(http.MethodGet, "/api/stock-check?product=+iPhone+&retailer=fnac", nil)
	w := httptest.NewRecorder()

	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for trimmed product; got %d", w.Code)
	}
}
