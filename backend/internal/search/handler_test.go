package search_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/search"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeSearchRepo struct {
	results []search.Result
	err     error
}

func (f *fakeSearchRepo) Search(_ context.Context, _ []float64, _ int) ([]search.Result, error) {
	return f.results, f.err
}
func (f *fakeSearchRepo) Similar(_ context.Context, _ uuid.UUID, _ int) ([]search.Result, error) {
	return f.results, f.err
}

type fakeEmbedder struct {
	vec []float64
	err error
}

func (f *fakeEmbedder) Embed(_ context.Context, _ string) ([]float64, error) {
	return f.vec, f.err
}
func (f *fakeEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i := range texts {
		out[i] = f.vec
	}
	return out, f.err
}

// ── helpers ──────────────────────────────────────────────────────────────────

func newHandler(repo search.Repository, embedder *fakeEmbedder) *search.Handler {
	return search.NewHandler(repo, embedder, nil, nil)
}

func doRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestSearch_MissingQuery(t *testing.T) {
	h := newHandler(&fakeSearchRepo{}, &fakeEmbedder{vec: []float64{1}})
	rr := doRequest(t, http.HandlerFunc(h.Search), http.MethodGet, "/api/search")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestSearch_EmbedFailReturnsEmpty(t *testing.T) {
	embedder := &fakeEmbedder{err: context.DeadlineExceeded}
	h := newHandler(&fakeSearchRepo{results: []search.Result{{Name: "X"}}}, embedder)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=laptop", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results := body["results"].([]any)
	if len(results) != 0 {
		t.Errorf("expected empty results on embed failure, got %d", len(results))
	}
}

func TestSearch_ReturnsResults(t *testing.T) {
	id := uuid.New()
	mid := uuid.New()
	embedder := &fakeEmbedder{vec: []float64{0.1, 0.2}}
	repo := &fakeSearchRepo{results: []search.Result{
		{ID: id, MissionID: mid, Name: "Sony WH-1000XM5", Similarity: 0.92},
	}}
	h := newHandler(repo, embedder)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=headphones", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results := body["results"].([]any)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSimilar_InvalidID(t *testing.T) {
	h := newHandler(&fakeSearchRepo{}, &fakeEmbedder{})

	r := chi.NewRouter()
	r.Get("/api/options/{optionID}/similar", h.Similar)

	req := httptest.NewRequest(http.MethodGet, "/api/options/not-a-uuid/similar", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestSimilar_ReturnsResults(t *testing.T) {
	targetID := uuid.New()
	similarID := uuid.New()
	mid := uuid.New()

	h := newHandler(
		&fakeSearchRepo{results: []search.Result{{ID: similarID, MissionID: mid, Name: "Bose QC45", Similarity: 0.88}}},
		&fakeEmbedder{vec: []float64{0.1}},
	)

	r := chi.NewRouter()
	r.Get("/api/options/{optionID}/similar", h.Similar)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+targetID.String()+"/similar", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	results := body["results"].([]any)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestReindex_NoWorker(t *testing.T) {
	h := newHandler(&fakeSearchRepo{}, &fakeEmbedder{})
	rr := doRequest(t, http.HandlerFunc(h.Reindex), http.MethodPost, "/api/search/reindex")
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rr.Code)
	}
}
