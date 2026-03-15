package timing_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/timing"
)

// ── fakes ────────────────────────────────────────────────────────────────────

// fakeAdvisor implements timing.Advisor so tests do not depend on an LLM.
type fakeAdvisor struct {
	advice *timing.TimingAdvice
	err    error
	called int
}

func (f *fakeAdvisor) Advise(_ context.Context, _, _ string, _ float64, _ time.Time) (*timing.TimingAdvice, error) {
	f.called++
	return f.advice, f.err
}

// fakeMissionGetter implements timing.MissionInfoGetter.
type fakeMissionGetter struct {
	name      string
	category  string
	budgetEur float64
	createdAt time.Time
	err       error
}

func (f *fakeMissionGetter) GetMissionInfo(_ context.Context, _ string) (string, string, float64, time.Time, error) {
	return f.name, f.category, f.budgetEur, f.createdAt, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func goodAdvice(missionID string) *timing.TimingAdvice {
	return &timing.TimingAdvice{
		MissionID:      missionID,
		Recommendation: timing.RecommendWait,
		Rationale:      "Black Friday is 3 weeks away; electronics prices typically drop 15-30%.",
		WaitUntil:      "2026-11-28",
		NextSaleEvent:  "Black Friday 2026",
		Confidence:     0.82,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
	}
}

func goodGetter() *fakeMissionGetter {
	return &fakeMissionGetter{
		name:      "New Laptop",
		category:  "computing",
		budgetEur: 1200.0,
		createdAt: time.Now().Add(-30 * 24 * time.Hour),
	}
}

func mountTimingRouter(h *timing.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/missions/{id}/timing-advice", h.GetTimingAdvice)
	return r
}

// ── POST /api/missions/{id}/timing-advice ─────────────────────────────────────

func TestGetTimingAdvice_HappyPath(t *testing.T) {
	const missionID = "11111111-1111-1111-1111-111111111111"

	advisor := &fakeAdvisor{advice: goodAdvice(missionID)}
	getter := goodGetter()
	cache := timing.NewCache(24 * time.Hour)
	h := timing.NewHandler(getter, cache, advisor)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID+"/timing-advice", nil)
	w := httptest.NewRecorder()
	mountTimingRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var got timing.TimingAdvice
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.MissionID != missionID {
		t.Errorf("expected missionId %s, got %s", missionID, got.MissionID)
	}
	if got.Recommendation != timing.RecommendWait {
		t.Errorf("expected recommendation WAIT, got %s", got.Recommendation)
	}
	if got.Confidence != 0.82 {
		t.Errorf("expected confidence 0.82, got %f", got.Confidence)
	}
	if advisor.called != 1 {
		t.Errorf("expected advisor called once, got %d", advisor.called)
	}
}

func TestGetTimingAdvice_CacheHit(t *testing.T) {
	const missionID = "22222222-2222-2222-2222-222222222222"

	advisor := &fakeAdvisor{advice: goodAdvice(missionID)}
	getter := goodGetter()
	cache := timing.NewCache(24 * time.Hour)
	h := timing.NewHandler(getter, cache, advisor)
	router := mountTimingRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID+"/timing-advice", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache; advisor not called again.
	r2 := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID+"/timing-advice", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("cached request failed: %d: %s", w2.Code, w2.Body.String())
	}

	if advisor.called != 1 {
		t.Errorf("expected advisor called exactly once (cache hit on 2nd), got %d", advisor.called)
	}
}

func TestGetTimingAdvice_AgentError(t *testing.T) {
	const missionID = "33333333-3333-3333-3333-333333333333"

	advisor := &fakeAdvisor{err: errors.New("llm unavailable")}
	getter := goodGetter()
	cache := timing.NewCache(24 * time.Hour)
	h := timing.NewHandler(getter, cache, advisor)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID+"/timing-advice", nil)
	w := httptest.NewRecorder()
	mountTimingRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetTimingAdvice_MissionNotFound(t *testing.T) {
	const missionID = "44444444-4444-4444-4444-444444444444"

	advisor := &fakeAdvisor{}
	getter := &fakeMissionGetter{err: errors.New("mission not found")}
	cache := timing.NewCache(24 * time.Hour)
	h := timing.NewHandler(getter, cache, advisor)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID+"/timing-advice", nil)
	w := httptest.NewRecorder()
	mountTimingRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if advisor.called != 0 {
		t.Errorf("advisor should not be called when mission not found, called %d times", advisor.called)
	}
}

// ── cache unit tests ─────────────────────────────────────────────────────────

func TestCache_SetAndGet(t *testing.T) {
	c := timing.NewCache(time.Hour)
	adv := goodAdvice("abc")
	c.Set("abc", adv)

	got, ok := c.Get("abc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.MissionID != adv.MissionID {
		t.Errorf("expected %s, got %s", adv.MissionID, got.MissionID)
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	c := timing.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := timing.NewCache(1 * time.Millisecond)
	c.Set("key", goodAdvice("key"))
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := timing.NewCache(time.Hour)
	c.Set("key", goodAdvice("key"))
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after Delete")
	}
}

// ── model constants ───────────────────────────────────────────────────────────

func TestRecommendationConstants(t *testing.T) {
	if timing.RecommendBuyNow != "BUY_NOW" {
		t.Errorf("RecommendBuyNow = %q, want BUY_NOW", timing.RecommendBuyNow)
	}
	if timing.RecommendWait != "WAIT" {
		t.Errorf("RecommendWait = %q, want WAIT", timing.RecommendWait)
	}
	if timing.RecommendUnsure != "UNSURE" {
		t.Errorf("RecommendUnsure = %q, want UNSURE", timing.RecommendUnsure)
	}
}
