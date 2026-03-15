package substitute

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// OptionInfoGetter resolves an option ID to its name, category, and budget.
// Tests inject a fake; production wires the option repository via OptionRepoGetter.
type OptionInfoGetter interface {
	GetOptionInfo(ctx context.Context, optionID string) (name, category string, budget float64, err error)
}

// Handler exposes the substitutes endpoint over HTTP.
type Handler struct {
	getter OptionInfoGetter
	cache  *Cache
	agent  SubstituteAgent
}

// NewHandler returns a Handler wired to the given option getter, cache, and agent.
func NewHandler(getter OptionInfoGetter, cache *Cache, agent SubstituteAgent) *Handler {
	return &Handler{getter: getter, cache: cache, agent: agent}
}

// GetSubstitutes handles GET /api/options/{id}/substitutes.
// It serves from cache when possible; otherwise calls the SubstituteAgent and
// caches the result for 2 hours.
func (h *Handler) GetSubstitutes(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "id")

	// Serve from cache if present.
	if cached, ok := h.cache.Get(optionID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Resolve option details.
	name, category, budget, err := h.getter.GetOptionInfo(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if name == "" {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	// Call the agent.
	subs, err := h.agent.FindSubstitutes(r.Context(), name, category, budget)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "substitute service unavailable")
		return
	}

	resp := &SubstituteResponse{
		OptionID:    optionID,
		ProductName: name,
		Substitutes: subs,
		GeneratedAt: time.Now().UTC(),
	}

	h.cache.Set(optionID, resp)
	httputil.WriteJSON(w, http.StatusOK, resp)
}
