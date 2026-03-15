package itemtagger

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 24 * time.Hour

type cacheEntry struct {
	tags      ItemTags
	expiresAt time.Time
}

// Handler exposes the item tagger endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetTags handles GET /api/missions/{missionId}/items/{itemId}/tags.
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Cache hit.
	if v, ok := h.cache.Load(itemID.String()); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.tags)
			return
		}
		h.cache.Delete(itemID.String())
	}

	// Load item name from DB.
	var itemName string
	err = h.pool.QueryRow(r.Context(),
		`SELECT name FROM shopping_items WHERE id = $1`, itemID,
	).Scan(&itemName)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	tags := Tag(itemID.String(), itemName)

	h.cache.Store(itemID.String(), cacheEntry{
		tags:      tags,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, tags)
}
