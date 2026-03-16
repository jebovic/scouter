package pricecomp

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

// Comparator is the narrow interface consumed by Handler so tests can inject
// a fake without depending on the real LLM.
type Comparator interface {
	Compare(ctx context.Context, productName, optionID string) (*PriceComparison, error)
}

// Handler exposes the price comparison endpoints over HTTP.
type Handler struct {
	getter     OptionNameGetter
	cache      *Cache
	comparator Comparator
}

// NewHandler returns a Handler wired to the given name getter, cache, and comparator.
func NewHandler(getter OptionNameGetter, cache *Cache, comparator Comparator) *Handler {
	return &Handler{getter: getter, cache: cache, comparator: comparator}
}

// GetComparison handles GET /api/options/{id}/price-comparison.
// It serves from cache when possible; otherwise calls the PriceCompAgent and
// caches the result.
func (h *Handler) GetComparison(w http.ResponseWriter, r *http.Request) {
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

	comparison, err := h.comparator.Compare(r.Context(), name, optionID)
	if err != nil {
		degraded := &PriceComparison{
			OptionID:  optionID,
			Product:   name,
			Offers:    []RetailerOffer{},
			LowestIdx: 0,
			FetchedAt: "",
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	h.cache.Set(optionID, comparison)
	httputil.WriteJSON(w, http.StatusOK, comparison)
}

// RefreshComparison handles POST /api/options/{id}/price-comparison/refresh.
// It clears the cache entry for the option and re-runs the PriceCompAgent,
// returning the freshly generated comparison.
func (h *Handler) RefreshComparison(w http.ResponseWriter, r *http.Request) {
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

	comparison, err := h.comparator.Compare(r.Context(), name, optionID)
	if err != nil {
		degraded := &PriceComparison{
			OptionID:  optionID,
			Product:   name,
			Offers:    []RetailerOffer{},
			LowestIdx: 0,
			FetchedAt: "",
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	h.cache.Set(optionID, comparison)
	httputil.WriteJSON(w, http.StatusOK, comparison)
}
