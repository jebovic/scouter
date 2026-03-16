package carbon

import (
	"net/http"
	"strings"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the AI carbon footprint estimation endpoint over HTTP.
type Handler struct {
	agent CarbonAgentInterface
	cache *Cache
}

// NewHandler returns a Handler wired to the given agent and cache.
func NewHandler(agent CarbonAgentInterface, cache *Cache) *Handler {
	return &Handler{
		agent: agent,
		cache: cache,
	}
}

// GetEstimate handles GET /api/carbon?itemName=...&category=...
// Checks the cache first; on miss delegates to the CarbonAgent, caches and returns the result.
func (h *Handler) GetEstimate(w http.ResponseWriter, r *http.Request) {
	itemName := strings.TrimSpace(r.URL.Query().Get("itemName"))
	if itemName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "itemName is required")
		return
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))

	cacheKey := itemName + ":" + category
	if cached, ok := h.cache.Get(cacheKey); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	estimate, err := h.agent.Estimate(r.Context(), itemName, category)
	if err != nil {
		degraded := &CarbonEstimate{
			ItemName:   itemName,
			Category:   category,
			KgCO2:      0,
			Label:      "Inconnu",
			Tips:       []string{},
			Confidence: "unavailable",
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	h.cache.Set(cacheKey, estimate)
	httputil.WriteJSON(w, http.StatusOK, estimate)
}
