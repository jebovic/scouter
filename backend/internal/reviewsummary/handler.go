package reviewsummary

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

// cacheEntry pairs a summary with its expiry timestamp.
type cacheEntry struct {
	value     *ReviewSummary
	expiresAt time.Time
}

// Handler serves GET /api/items/{itemID}/review-summary.
// Summaries are generated deterministically from the item ID hash and cached
// in-memory for cacheTTL (30 minutes) to avoid repeated computation.
type Handler struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry
}

// NewHandler returns a ready-to-use Handler.
func NewHandler() *Handler {
	return &Handler{cache: make(map[string]cacheEntry)}
}

// GetReviewSummary handles GET /api/items/{itemID}/review-summary.
// The optional "itemName" query parameter is used for display purposes only;
// the summary data itself is seeded by itemID so it is always consistent.
func (h *Handler) GetReviewSummary(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing itemID")
		return
	}

	itemName := r.URL.Query().Get("itemName")

	if cached := h.getCached(itemID); cached != nil {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	summary := generate(itemID, itemName)

	h.setCached(itemID, summary)
	httputil.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) getCached(key string) *ReviewSummary {
	h.mu.RLock()
	e, ok := h.cache[key]
	h.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil
	}
	return e.value
}

func (h *Handler) setCached(key string, v *ReviewSummary) {
	h.mu.Lock()
	h.cache[key] = cacheEntry{value: v, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()
}
