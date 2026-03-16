package summary_test

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
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
	"github.com/jibei/scouter/internal/summary"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeBriefer struct {
	result *summary.MissionSummary
	err    error
	calls  int
}

func (f *fakeBriefer) Brief(_ context.Context, _ mission.Mission, _ []option.Option, _ []shopping.Item) (*summary.MissionSummary, error) {
	f.calls++
	return f.result, f.err
}

type fakeMissionGetter struct {
	m   *mission.Mission
	err error
}

func (f *fakeMissionGetter) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return f.m, f.err
}

type fakeOptionLister struct {
	opts []option.Option
	err  error
}

func (f *fakeOptionLister) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return f.opts, f.err
}

type fakeShoppingLister struct {
	items []shopping.Item
	err   error
}

func (f *fakeShoppingLister) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return f.items, f.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

const testSlug = "test-mission"

func goodMission() *mission.Mission {
	return &mission.Mission{
		ID:       uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Slug:     testSlug,
		Name:     "Casque Audio",
		Category: "electronics",
		Budget:   300.0,
		Currency: "EUR",
	}
}

func makeSummary(verdict string) *summary.MissionSummary {
	return &summary.MissionSummary{
		MissionSlug: testSlug,
		Bullets: []string{
			"✅ Meilleure option: Sony WH-1000XM5 à 279€",
			"⚠️ Budget: 85% utilisé",
			"🎯 Prochaine action: comparer avec Bose QC45",
		},
		Verdict:  verdict,
		CachedAt: time.Now().Unix(),
	}
}

func mountRouter(h *summary.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/missions/{slug}/summary", h.Get)
	return r
}

func buildHandler(mget *fakeMissionGetter, olist *fakeOptionLister, slist *fakeShoppingLister, br *fakeBriefer, cache *summary.Cache) *summary.Handler {
	return summary.NewHandler(mget, olist, slist, br, cache)
}

// ── GET /api/missions/{slug}/summary — success verdicts ───────────────────────

func TestHandler_Get_GoVerdict(t *testing.T) {
	br := &fakeBriefer{result: makeSummary("go")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got summary.MissionSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Verdict != "go" {
		t.Errorf("expected verdict 'go', got %q", got.Verdict)
	}
	if got.MissionSlug != testSlug {
		t.Errorf("expected missionSlug %q, got %q", testSlug, got.MissionSlug)
	}
	if len(got.Bullets) < 3 {
		t.Errorf("expected at least 3 bullets, got %d", len(got.Bullets))
	}
	if got.CachedAt == 0 {
		t.Error("expected cachedAt to be set")
	}
}

func TestHandler_Get_WaitVerdict(t *testing.T) {
	br := &fakeBriefer{result: makeSummary("wait")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got summary.MissionSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Verdict != "wait" {
		t.Errorf("expected verdict 'wait', got %q", got.Verdict)
	}
}

func TestHandler_Get_ResearchMoreVerdict(t *testing.T) {
	br := &fakeBriefer{result: makeSummary("research_more")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got summary.MissionSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Verdict != "research_more" {
		t.Errorf("expected verdict 'research_more', got %q", got.Verdict)
	}
}

// ── cache hit ─────────────────────────────────────────────────────────────────

func TestHandler_Get_CacheHit_ReturnsSameResult(t *testing.T) {
	br := &fakeBriefer{result: makeSummary("go")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)
	router := mountRouter(h)

	// First call: generates.
	r1 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first GET failed: %d", w1.Code)
	}

	// Second call: must come from cache (briefer not called again).
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second GET failed: %d", w2.Code)
	}

	if br.calls != 1 {
		t.Errorf("expected briefer called once (cache hit on 2nd), got %d calls", br.calls)
	}

	var got1, got2 summary.MissionSummary
	_ = json.NewDecoder(w1.Body).Decode(&got1)
	_ = json.NewDecoder(w2.Body).Decode(&got2)
	if got1.CachedAt != got2.CachedAt {
		t.Errorf("expected same cachedAt on cache hit: %d vs %d", got1.CachedAt, got2.CachedAt)
	}
}

// ── 404 for unknown slug ──────────────────────────────────────────────────────

func TestHandler_Get_UnknownSlug_Returns404(t *testing.T) {
	br := &fakeBriefer{}
	mget := &fakeMissionGetter{m: nil, err: nil} // nil mission = not found
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/unknown-slug/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if br.calls != 0 {
		t.Errorf("briefer should not be called when mission not found, got %d calls", br.calls)
	}
}

// ── error paths ───────────────────────────────────────────────────────────────

func TestHandler_Get_MissionRepoError_Returns500(t *testing.T) {
	br := &fakeBriefer{}
	mget := &fakeMissionGetter{err: errors.New("db error")}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_BrieferError_ReturnsDegradedDTO(t *testing.T) {
	br := &fakeBriefer{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 degraded DTO, got %d: %s", w.Code, w.Body.String())
	}
	var got summary.MissionSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode degraded DTO: %v", err)
	}
	if got.Verdict != "research_more" {
		t.Errorf("expected verdict 'research_more', got %q", got.Verdict)
	}
	if len(got.Bullets) == 0 {
		t.Error("expected at least one bullet in degraded DTO")
	}
	if got.CachedAt <= 0 {
		t.Error("expected cachedAt > 0 in degraded DTO")
	}
	if got.MissionSlug != testSlug {
		t.Errorf("expected missionSlug %q, got %q", testSlug, got.MissionSlug)
	}
}

func TestHandler_Get_BrieferError_DegradedResponseNotCached(t *testing.T) {
	br := &fakeBriefer{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(time.Hour)
	h := buildHandler(mget, olist, slist, br, cache)
	router := mountRouter(h)

	// First GET: briefer fails, returns degraded DTO.
	r1 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first GET: expected 200, got %d", w1.Code)
	}

	// Second GET: briefer should be called again (degraded response was not cached).
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/summary", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second GET: expected 200, got %d", w2.Code)
	}

	// Both responses must be valid degraded DTOs.
	var got1, got2 summary.MissionSummary
	_ = json.NewDecoder(w1.Body).Decode(&got1)
	_ = json.NewDecoder(w2.Body).Decode(&got2)
	if got1.Verdict != "research_more" || got2.Verdict != "research_more" {
		t.Errorf("expected both responses to be degraded: got %q, %q", got1.Verdict, got2.Verdict)
	}

	// Briefer must have been called twice — degraded DTO was NOT served from cache.
	if br.calls != 2 {
		t.Errorf("expected briefer called 2 times (not cached), got %d", br.calls)
	}
}

// ── cache unit tests ──────────────────────────────────────────────────────────

func TestCache_SetAndGet(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s := makeSummary("go")
	s.MissionSlug = "my-mission"
	c.Set("my-mission", s)

	got, ok := c.Get("my-mission")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Verdict != "go" {
		t.Errorf("expected verdict 'go', got %q", got.Verdict)
	}
}

func TestCache_MissOnUnknownKey(t *testing.T) {
	c := summary.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := summary.NewCache(1 * time.Millisecond)
	s := makeSummary("wait")
	c.Set("key", s)
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s := makeSummary("go")
	c.Set("key", s)
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after Delete")
	}
}

func TestCache_OverwriteExisting(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s1 := makeSummary("go")
	s2 := makeSummary("wait")
	c.Set("key", s1)
	c.Set("key", s2)

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cache hit after overwrite")
	}
	if got.Verdict != "wait" {
		t.Errorf("expected 'wait', got %q", got.Verdict)
	}
}
