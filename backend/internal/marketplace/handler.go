package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves the French marketplace price comparison endpoint.
type Handler struct{}

// NewHandler returns a new Handler.
func NewHandler() *Handler { return &Handler{} }

// RegisterRoutes mounts the marketplace route on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/marketplace", h.handleSearch)
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		httputil.WriteError(w, http.StatusBadRequest, "query too short")
		return
	}

	results := make([]MarketplaceResult, 0, len(Sites))
	for _, site := range Sites {
		results = append(results, MarketplaceResult{
			Site:      site.Name,
			URL:       site.SearchFn(query),
			Available: true,
			Label:     site.Label,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, MarketplaceResponse{
		ItemName: query,
		Results:  results,
	})
}
