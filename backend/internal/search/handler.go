package search

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
)

const (
	defaultSearchLimit  = 10
	maxSearchLimit      = 50
	defaultSimilarLimit = 5
	maxSimilarLimit     = 20
)

// Handler serves semantic search endpoints.
type Handler struct {
	repo      Repository
	embedder  llm.EmbedProvider
	embedRepo embedding.Repository
	worker    *embedding.Worker
}

// NewHandler creates a search handler. worker may be nil (reindex endpoint returns 503).
func NewHandler(repo Repository, embedder llm.EmbedProvider, embedRepo embedding.Repository, worker *embedding.Worker) *Handler {
	return &Handler{repo: repo, embedder: embedder, embedRepo: embedRepo, worker: worker}
}

// GET /api/search?q=<text>&limit=<n>
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "q is required")
		return
	}
	limit := parseLimit(r, "limit", defaultSearchLimit, maxSearchLimit)

	vec, err := h.embedder.Embed(r.Context(), q)
	if err != nil {
		// Embed service unavailable — return empty results gracefully.
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"results": []Result{}})
		return
	}

	results, err := h.repo.Search(r.Context(), vec, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "search failed")
		return
	}
	if results == nil {
		results = []Result{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"results": results})
}

// GET /api/options/{optionID}/similar?limit=<n>
func (h *Handler) Similar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "optionID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option ID")
		return
	}
	limit := parseLimit(r, "limit", defaultSimilarLimit, maxSimilarLimit)

	results, err := h.repo.Similar(r.Context(), id, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "similar search failed")
		return
	}
	if results == nil {
		results = []Result{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"results": results})
}

// POST /api/search/reindex
// Enqueues all un-embedded options for (re)embedding. Returns the count enqueued.
func (h *Handler) Reindex(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "embedding worker not configured")
		return
	}
	n := h.worker.Backfill(r.Context(), 500)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"enqueued": n})
}

func parseLimit(r *http.Request, param string, def, max int) int {
	v := r.URL.Query().Get(param)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}
