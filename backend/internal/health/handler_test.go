package health_test

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
	"github.com/jibei/scouter/internal/health"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

type stubMissionRepo struct {
	m   *mission.Mission
	err error
}

func (s *stubMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return s.m, s.err
}

type stubOptionRepo struct {
	opts []option.Option
	err  error
}

func (s *stubOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return s.opts, s.err
}

type stubShoppingRepo struct {
	items []shopping.Item
	err   error
}

func (s *stubShoppingRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return s.items, s.err
}

type stubSummarizer struct {
	summary string
	advice  string
	err     error
	calls   int
}

func (s *stubSummarizer) BuildHealthSummary(_ context.Context, _ string, _ *health.HealthReport) (string, string, error) {
	s.calls++
	return s.summary, s.advice, s.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

const testSlug = "test-mission"
const testMissionIDStr = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func goodMissionForHealth() *mission.Mission {
	return &mission.Mission{
		ID:       uuid.MustParse(testMissionIDStr),
		Slug:     testSlug,
		Name:     "Test Mission",
		Category: "computing",
		Budget:   1500.0,
		Currency: "EUR",
		Phase:    "researching",
	}
}

func mountHealthRouter(h *health.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/missions/{slug}/health", h.GetHealth)
	return r
}

func newTestHandler(
	mRepo *stubMissionRepo,
	oRepo *stubOptionRepo,
	sRepo *stubShoppingRepo,
	sum *stubSummarizer,
) *health.Handler {
	cache := health.NewCache(30 * time.Minute)
	return health.NewHandler(mRepo, oRepo, sRepo, sum, cache)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestHandler_GetHealth_HappyPath(t *testing.T) {
	mRepo := &stubMissionRepo{m: goodMissionForHealth()}
	oRepo := &stubOptionRepo{opts: []option.Option{{}, {}, {}}} // 3 options
	sRepo := &stubShoppingRepo{}
	sum := &stubSummarizer{summary: "Mission en bonne progression.", advice: "Continuez la recherche."}
	h := newTestHandler(mRepo, oRepo, sRepo, sum)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/health", nil)
	w := httptest.NewRecorder()
	mountHealthRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var report health.HealthReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if report.MissionID == "" {
		t.Error("expected MissionID to be set")
	}
	if len(report.Signals) != 4 {
		t.Errorf("expected 4 signals, got %d", len(report.Signals))
	}
	if report.Grade == "" {
		t.Error("expected Grade to be set")
	}
	if report.Summary == "" {
		t.Error("expected Summary to be set")
	}
	if report.Advice == "" {
		t.Error("expected Advice to be set")
	}
	if report.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
}

func TestHandler_GetHealth_MissionNotFound(t *testing.T) {
	mRepo := &stubMissionRepo{m: nil, err: nil}
	oRepo := &stubOptionRepo{}
	sRepo := &stubShoppingRepo{}
	sum := &stubSummarizer{}
	h := newTestHandler(mRepo, oRepo, sRepo, sum)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/nonexistent/health", nil)
	w := httptest.NewRecorder()
	mountHealthRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetHealth_MissionRepoError(t *testing.T) {
	mRepo := &stubMissionRepo{err: errors.New("db error")}
	oRepo := &stubOptionRepo{}
	sRepo := &stubShoppingRepo{}
	sum := &stubSummarizer{}
	h := newTestHandler(mRepo, oRepo, sRepo, sum)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/health", nil)
	w := httptest.NewRecorder()
	mountHealthRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetHealth_SummaryCached(t *testing.T) {
	mRepo := &stubMissionRepo{m: goodMissionForHealth()}
	oRepo := &stubOptionRepo{}
	sRepo := &stubShoppingRepo{}
	sum := &stubSummarizer{summary: "Un résumé.", advice: "Un conseil."}
	h := newTestHandler(mRepo, oRepo, sRepo, sum)
	router := mountHealthRouter(h)

	// First request
	r1 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/health", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — summarizer should not be called again (cache hit)
	r2 := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/health", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if sum.calls > 1 {
		t.Errorf("expected summarizer called at most once (cache hit), got %d calls", sum.calls)
	}
}

func TestHandler_GetHealth_OverallScoreInRange(t *testing.T) {
	mRepo := &stubMissionRepo{m: goodMissionForHealth()}
	oRepo := &stubOptionRepo{opts: []option.Option{{}, {}, {}, {}, {}}} // 5 options
	sRepo := &stubShoppingRepo{}
	sum := &stubSummarizer{summary: "Bonne mission.", advice: "Continuez."}
	h := newTestHandler(mRepo, oRepo, sRepo, sum)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+testSlug+"/health", nil)
	w := httptest.NewRecorder()
	mountHealthRouter(h).ServeHTTP(w, req)

	var report health.HealthReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if report.OverallScore < 0 || report.OverallScore > 100 {
		t.Errorf("overall score %d out of range [0,100]", report.OverallScore)
	}
}
