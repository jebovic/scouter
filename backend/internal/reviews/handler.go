package reviews

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// OptionNameGetter is the narrow interface used by Handler to resolve an option
// ID to its product name. Tests can inject a fake; production wires the option
// repository directly.
type OptionNameGetter interface {
	// GetOptionName returns the name for the given option ID.
	// It returns ("", nil) when the option does not exist.
	GetOptionName(ctx context.Context, optionID string) (string, error)
}

// Handler exposes the reviews endpoints over HTTP.
type Handler struct {
	getter  OptionNameGetter
	cache   *Cache
	reviewer Reviewer
}

// NewHandler returns a Handler wired to the given name getter, cache, and reviewer.
func NewHandler(getter OptionNameGetter, cache *Cache, reviewer Reviewer) *Handler {
	return &Handler{getter: getter, cache: cache, reviewer: reviewer}
}

// GetReviews handles GET /api/options/{id}/reviews.
// It serves from cache when possible; otherwise calls the ReviewAgent and
// caches the result.
func (h *Handler) GetReviews(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "id")

	name, err := h.getter.GetOptionName(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if name == "" {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	if cached, ok := h.cache.Get(optionID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	summary, err := h.reviewer.Summarise(r.Context(), name)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "review service unavailable")
		return
	}
	summary.OptionID = optionID

	h.cache.Set(optionID, summary)
	httputil.WriteJSON(w, http.StatusOK, summary)
}

// RefreshReviews handles POST /api/options/{id}/reviews/refresh.
// It clears the cache entry for the option and re-runs the ReviewAgent,
// returning the freshly generated summary.
func (h *Handler) RefreshReviews(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "id")

	name, err := h.getter.GetOptionName(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if name == "" {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	h.cache.Delete(optionID)

	summary, err := h.reviewer.Summarise(r.Context(), name)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "review service unavailable")
		return
	}
	summary.OptionID = optionID

	h.cache.Set(optionID, summary)
	httputil.WriteJSON(w, http.StatusOK, summary)
}
