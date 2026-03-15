package dealfeed

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
)

// Handler exposes the promo feed endpoint over HTTP.
type Handler struct {
	agent  PromoAgent
	cache  *Cache
	logger *slog.Logger
}

// NewHandler wires a Handler from real dependencies used in production.
func NewHandler(provider llm.Provider, logger *slog.Logger) *Handler {
	return &Handler{
		agent:  NewPromoFinderAgent(provider),
		cache:  NewCache(30 * time.Minute),
		logger: logger,
	}
}

// NewHandlerWithDeps creates a Handler with explicit dependencies (for testing).
func NewHandlerWithDeps(agent PromoAgent, cache *Cache) *Handler {
	return &Handler{
		agent:  agent,
		cache:  cache,
		logger: slog.Default(),
	}
}

// Get handles GET /api/promo-feed?q=laptop&category=informatique.
// Returns 422 when q is missing or shorter than 2 characters.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "Le paramètre q est requis (minimum 2 caractères)")
		return
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))

	// Cache hit?
	key := cacheKey(q, category)
	if cached := h.cache.Get(key); cached != nil {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Call agent.
	result, err := h.agent.FindPromos(r.Context(), q, category)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("promo finder agent error", "query", q, "category", category, "err", err)
		}
		httputil.WriteError(w, http.StatusBadGateway, "service de promotions indisponible")
		return
	}

	// Stamp category on each offer so they carry the right label.
	for i := range result.Offers {
		if result.Offers[i].Category == "" {
			result.Offers[i].Category = category
		}
	}

	h.cache.Set(key, result)
	httputil.WriteJSON(w, http.StatusOK, result)
}
