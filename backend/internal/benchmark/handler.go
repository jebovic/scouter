package benchmark

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// ItemGetter retrieves a shopping item by UUID.
// Tests inject a fake; production wires the shopping repository.
type ItemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

// Handler exposes the price-benchmark endpoint over HTTP.
type Handler struct {
	getter ItemGetter
	agent  BenchmarkAgentInterface
	cache  *Cache
}

// NewHandler returns a Handler wired to the given item getter, agent, and cache.
func NewHandler(getter ItemGetter, agent BenchmarkAgentInterface, cache *Cache) *Handler {
	return &Handler{getter: getter, agent: agent, cache: cache}
}

// Get handles GET /api/shopping-items/{id}/benchmark.
// Returns a BenchmarkResult with market range, verdict, and diff percentage.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from cache.
	if cached, ok := h.cache.Get(rawID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Resolve item details.
	item, err := h.getter.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	result, err := h.agent.Benchmark(r.Context(), rawID, item.Name, item.Price)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "benchmark unavailable")
		return
	}

	h.cache.Set(rawID, result)
	httputil.WriteJSON(w, http.StatusOK, result)
}
