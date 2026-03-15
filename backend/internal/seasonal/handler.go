package seasonal

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the AI seasonal discount prediction endpoint over HTTP.
type Handler struct {
	agent SeasonalAgentInterface
	cache *Cache
}

// NewHandler returns a Handler wired to the given agent and cache.
func NewHandler(agent SeasonalAgentInterface, cache *Cache) *Handler {
	return &Handler{
		agent: agent,
		cache: cache,
	}
}

// GetPrediction handles GET /api/seasonal?itemName=...&category=...
// Checks the cache first; on miss delegates to the Agent, caches and returns the result.
func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	itemName := strings.TrimSpace(r.URL.Query().Get("itemName"))
	if itemName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "itemName is required")
		return
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))

	cacheKey := itemName + "|" + category
	if cached, ok := h.cache.Get(cacheKey); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	prediction, err := h.agent.Predict(ctx, itemName, category)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "seasonal prediction unavailable")
		return
	}

	h.cache.Set(cacheKey, prediction)
	httputil.WriteJSON(w, http.StatusOK, prediction)
}
