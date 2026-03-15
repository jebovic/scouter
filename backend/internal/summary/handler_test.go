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

// ── fakes ────────────────────────────────────────────────────────────────────

// fakeGenerator implements summary.Generator interface for handler tests.
type fakeGenerator struct {
	content string
	err     error
	called  int
}

func (f *fakeGenerator) Generate(_ context.Context, _ mission.Mission, _ []option.Option, _ []shopping.Item) (string, error) {
	f.called++
	return f.content, f.err
}

// fakeMissionGetter implements summary.MissionGetter.
type fakeMissionGetter struct {
	m   *mission.Mission
	err error
}

func (f *fakeMissionGetter) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return f.m, f.err
}

// fakeOptionLister implements summary.OptionLister.
type fakeOptionLister struct {
	opts []option.Option
	err  error
}

func (f *fakeOptionLister) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return f.opts, f.err
}

// fakeShoppingLister implements summary.ShoppingLister.
type fakeShoppingLister struct {
	items []shopping.Item
	err   error
}

func (f *fakeShoppingLister) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return f.items, f.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

const testMissionID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func goodMission() *mission.Mission {
	id := uuid.MustParse(testMissionID)
	return &mission.Mission{
		ID:       id,
		Name:     "Test Mission",
		Category: "computing",
		Budget:   1500.0,
		Currency: "EUR",
	}
}

func mountSummaryRouter(h *summary.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/missions/{missionID}/summary", h.Generate)
	r.Get("/api/missions/{missionID}/summary", h.Get)
	return r
}

// ── POST /api/missions/{missionID}/summary — generate ────────────────────────

func TestHandler_Generate_HappyPath(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{opts: makeOptions()}
	slist := &fakeShoppingLister{items: makeShoppingItems()}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+testMissionID+"/summary", nil)
	w := httptest.NewRecorder()
	mountSummaryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var got summary.ShoppingSummary
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Content != sampleMarkdown {
		t.Errorf("expected content to match generated markdown")
	}
	if got.MissionID.String() != testMissionID {
		t.Errorf("expected missionId %s, got %s", testMissionID, got.MissionID)
	}
	if got.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
	if got.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if gen.called != 1 {
		t.Errorf("expected generator called once, got %d", gen.called)
	}
}

func TestHandler_Generate_MissionNotFound(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{m: nil, err: nil} // nil means not found
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+testMissionID+"/summary", nil)
	w := httptest.NewRecorder()
	mountSummaryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if gen.called != 0 {
		t.Errorf("generator should not be called when mission not found, got %d calls", gen.called)
	}
}

func TestHandler_Generate_MissionRepoError(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{err: errors.New("db error")}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+testMissionID+"/summary", nil)
	w := httptest.NewRecorder()
	mountSummaryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Generate_GeneratorError(t *testing.T) {
	gen := &fakeGenerator{err: errors.New("llm unavailable")}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{opts: makeOptions()}
	slist := &fakeShoppingLister{items: makeShoppingItems()}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+testMissionID+"/summary", nil)
	w := httptest.NewRecorder()
	mountSummaryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Generate_InvalidMissionID(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	r := chi.NewRouter()
	r.Post("/api/missions/{missionID}/summary", h.Generate)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/not-a-uuid/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

// ── GET /api/missions/{missionID}/summary ─────────────────────────────────────

func TestHandler_Get_CacheHit(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{opts: makeOptions()}
	slist := &fakeShoppingLister{items: makeShoppingItems()}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)
	router := mountSummaryRouter(h)

	// First: POST to generate and cache.
	r1 := httptest.NewRequest(http.MethodPost, "/api/missions/"+testMissionID+"/summary", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d %s", w1.Code, w1.Body.String())
	}

	// Second: GET should return cached value.
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testMissionID+"/summary", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on cache hit, got %d: %s", w2.Code, w2.Body.String())
	}

	var got summary.ShoppingSummary
	if err := json.NewDecoder(w2.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Content != sampleMarkdown {
		t.Errorf("expected cached content to match")
	}
	// Generator should only have been called once (from POST).
	if gen.called != 1 {
		t.Errorf("expected generator called once total (GET served from cache), got %d", gen.called)
	}
}

func TestHandler_Get_CacheMiss_Returns404(t *testing.T) {
	gen := &fakeGenerator{content: sampleMarkdown}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testMissionID+"/summary", nil)
	w := httptest.NewRecorder()
	mountSummaryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when no cached summary, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidMissionID(t *testing.T) {
	gen := &fakeGenerator{}
	mget := &fakeMissionGetter{m: goodMission()}
	olist := &fakeOptionLister{}
	slist := &fakeShoppingLister{}
	cache := summary.NewCache(24 * time.Hour)
	h := summary.NewHandler(gen, mget, olist, slist, cache)

	r := chi.NewRouter()
	r.Get("/api/missions/{missionID}/summary", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/missions/not-a-uuid/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

// ── Cache unit tests ──────────────────────────────────────────────────────────

func TestSummaryCache_SetAndGet(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s := &summary.ShoppingSummary{
		ID:          uuid.New(),
		MissionID:   uuid.MustParse(testMissionID),
		Content:     "# Test",
		GeneratedAt: time.Now().UTC(),
	}
	c.Set(testMissionID, s)

	got, ok := c.Get(testMissionID)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Content != "# Test" {
		t.Errorf("expected content '# Test', got %s", got.Content)
	}
}

func TestSummaryCache_MissOnUnknownKey(t *testing.T) {
	c := summary.NewCache(time.Hour)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestSummaryCache_ExpiredEntry(t *testing.T) {
	c := summary.NewCache(1 * time.Millisecond)
	s := &summary.ShoppingSummary{ID: uuid.New(), Content: "test"}
	c.Set("key", s)
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestSummaryCache_Delete(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s := &summary.ShoppingSummary{ID: uuid.New(), Content: "test"}
	c.Set("key", s)
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after Delete")
	}
}

func TestSummaryCache_OverwriteExisting(t *testing.T) {
	c := summary.NewCache(time.Hour)
	s1 := &summary.ShoppingSummary{ID: uuid.New(), Content: "first"}
	s2 := &summary.ShoppingSummary{ID: uuid.New(), Content: "second"}
	c.Set("key", s1)
	c.Set("key", s2)

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cache hit after overwrite")
	}
	if got.Content != "second" {
		t.Errorf("expected 'second', got %s", got.Content)
	}
}
