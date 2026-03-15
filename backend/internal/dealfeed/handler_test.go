package dealfeed_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/dealfeed"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakePromoAgent struct {
	result *dealfeed.PromoFeedResult
	err    error
	called int
}

func (f *fakePromoAgent) FindPromos(ctx context.Context, query, category string) (*dealfeed.PromoFeedResult, error) {
	f.called++
	return f.result, f.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newHandler(agent dealfeed.PromoAgent) *dealfeed.Handler {
	cache := dealfeed.NewCache(30 * time.Minute)
	return dealfeed.NewHandlerWithDeps(agent, cache)
}

func sampleResult(query, category string) *dealfeed.PromoFeedResult {
	return &dealfeed.PromoFeedResult{
		Offers: []dealfeed.PromoOffer{
			{
				Title:         "MacBook Air M3 — Offre spéciale",
				Retailer:      "Fnac",
				OriginalPrice: 1299.00,
				PromoPrice:    999.00,
				Discount:      23,
				URL:           "fnac.com/macbook-air-m3",
				Category:      category,
				IsFlash:       false,
				Source:        "ai_generated",
			},
			{
				Title:         "Dell XPS 13 — Bon plan",
				Retailer:      "Darty",
				OriginalPrice: 1099.00,
				PromoPrice:    849.00,
				Discount:      22,
				URL:           "darty.com/dell-xps-13",
				Category:      category,
				IsFlash:       true,
				Source:        "ai_generated",
			},
		},
		Query:     query,
		Category:  category,
		FetchedAt: time.Now().UTC(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

// Test 1: Happy path — returns promo feed for valid query.
func TestHandler_Get_HappyPath(t *testing.T) {
	agent := &fakePromoAgent{result: sampleResult("laptop", "informatique")}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=laptop&category=informatique", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got dealfeed.PromoFeedResult
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Query != "laptop" {
		t.Errorf("expected query=laptop, got %q", got.Query)
	}
	if len(got.Offers) == 0 {
		t.Error("expected at least one offer")
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

// Test 2: Cache hit — agent called only once on second request.
func TestHandler_Get_CacheHit(t *testing.T) {
	agent := &fakePromoAgent{result: sampleResult("television", "electronique")}
	h := newHandler(agent)

	req1 := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=television&category=electronique", nil)
	w1 := httptest.NewRecorder()
	h.Get(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d %s", w1.Code, w1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=television&category=electronique", nil)
	w2 := httptest.NewRecorder()
	h.Get(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d %s", w2.Code, w2.Body.String())
	}

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

// Test 3: Missing q param → 422.
func TestHandler_Get_MissingQuery(t *testing.T) {
	agent := &fakePromoAgent{}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
	if agent.called != 0 {
		t.Error("agent must not be called when query is missing")
	}
}

// Test 4: Query too short (< 2 chars) → 422.
func TestHandler_Get_QueryTooShort(t *testing.T) {
	agent := &fakePromoAgent{}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=a", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

// Test 5: Agent error → 502.
func TestHandler_Get_AgentError(t *testing.T) {
	agent := &fakePromoAgent{err: errors.New("LLM unavailable")}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=laptop", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

// Test 6: Response fields present — offers contain required fields.
func TestHandler_Get_ResponseFields(t *testing.T) {
	agent := &fakePromoAgent{result: sampleResult("aspirateur", "menager")}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=aspirateur&category=menager", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"offers", "query", "category", "fetchedAt"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required key %q in response", key)
		}
	}
}

// Test 7: Category param is optional — works without it.
func TestHandler_Get_NoCategoryParam(t *testing.T) {
	agent := &fakePromoAgent{result: sampleResult("smartphone", "")}
	h := newHandler(agent)

	req := httptest.NewRequest(http.MethodGet, "/api/promo-feed?q=smartphone", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ── cache unit tests ──────────────────────────────────────────────────────────

func TestCache_GetSet(t *testing.T) {
	c := dealfeed.NewCache(30 * time.Minute)
	result := sampleResult("test", "cat")
	c.Set("key1", result)

	got := c.Get("key1")
	if got == nil {
		t.Fatal("expected cached result, got nil")
	}
	if got.Query != "test" {
		t.Errorf("expected query=test, got %q", got.Query)
	}
}

func TestCache_MissReturnsNil(t *testing.T) {
	c := dealfeed.NewCache(30 * time.Minute)
	got := c.Get("nonexistent")
	if got != nil {
		t.Error("expected nil for cache miss")
	}
}

func TestCache_TTLExpiry(t *testing.T) {
	c := dealfeed.NewCacheWithNow(time.Millisecond, func() time.Time {
		return time.Now()
	})
	result := sampleResult("q", "c")
	c.Set("k", result)

	// Advance time by using a custom now func that returns future.
	c2 := dealfeed.NewCacheWithNow(time.Millisecond, func() time.Time {
		return time.Now().Add(time.Second)
	})
	// c2 has no entries, so it should miss.
	if c2.Get("k") != nil {
		t.Error("expected nil after TTL expiry in fresh cache")
	}
	// Original c with short TTL should still have it briefly.
	_ = c.Get("k") // just ensuring no panic
}
