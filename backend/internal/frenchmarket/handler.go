package frenchmarket

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

type cacheEntry struct {
	insight   MarketInsight
	expiresAt time.Time
}

// Handler serves GET /api/market/insight.
type Handler struct {
	mu    sync.Mutex
	cache sync.Map // key string → *cacheEntry
}

// NewHandler returns a ready-to-use Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// GetInsight handles GET /api/market/insight?item={name}&category={cat}
// Results are cached in-process for 1 hour.
func (h *Handler) GetInsight(w http.ResponseWriter, r *http.Request) {
	item := strings.TrimSpace(r.URL.Query().Get("item"))
	if item == "" {
		httputil.WriteError(w, http.StatusBadRequest, "item query parameter is required")
		return
	}
	category := strings.TrimSpace(r.URL.Query().Get("category"))

	cacheKey := item + "|" + category

	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.insight)
			return
		}
	}

	insight := GetInsight(item, category)

	h.cache.Store(cacheKey, &cacheEntry{
		insight:   insight,
		expiresAt: time.Now().Add(time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, insight)
}
