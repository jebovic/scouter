// Package stockcheck provides a deterministic heuristic stock-availability
// indicator for French retailers. It does not call any real retailer API —
// instead it hashes the product+retailer combination to produce a stable,
// pseudo-random status, backed by a short-lived in-memory cache.
package stockcheck

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// StockStatus represents the availability of a product at a retailer.
type StockStatus string

const (
	StatusInStock    StockStatus = "in_stock"
	StatusLimited    StockStatus = "limited"
	StatusOutOfStock StockStatus = "out_of_stock"
)

// StockResponse is the JSON payload returned by the handler.
type StockResponse struct {
	Status    StockStatus `json:"status"`
	Retailer  string      `json:"retailer"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

// DetermineStatus hashes the product+retailer combination to yield a
// deterministic stock status without any external API call.
//
// Distribution (mod 3):
//
//	0 → in_stock
//	1 → limited
//	2 → out_of_stock
func DetermineStatus(product, retailer string) StockStatus {
	h := fnv.New32a()
	_, _ = fmt.Fprintf(h, "%s|%s", strings.ToLower(strings.TrimSpace(product)), strings.ToLower(strings.TrimSpace(retailer)))
	switch h.Sum32() % 3 {
	case 0:
		return StatusInStock
	case 1:
		return StatusLimited
	default:
		return StatusOutOfStock
	}
}

// ── in-memory cache ───────────────────────────────────────────────────────────

type cacheEntry struct {
	status    StockStatus
	expiresAt time.Time
}

// Cache is a simple TTL-based in-memory store backed by sync.Map.
type Cache struct {
	m   sync.Map
	ttl time.Duration
}

// NewCache creates a Cache with the given per-entry TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl}
}

// Get returns the cached status and true, or zero-value and false when the key
// is absent or expired.
func (c *Cache) Get(key string) (StockStatus, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return "", false
	}
	e := v.(cacheEntry)
	if time.Now().After(e.expiresAt) {
		c.m.Delete(key)
		return "", false
	}
	return e.status, true
}

// Set stores status under key for the cache TTL duration.
func (c *Cache) Set(key string, status StockStatus) {
	c.m.Store(key, cacheEntry{
		status:    status,
		expiresAt: time.Now().Add(c.ttl),
	})
}

// ── HTTP handler ──────────────────────────────────────────────────────────────

// Handler serves GET /api/stock-check?product=<name>&retailer=<id>.
type Handler struct {
	cache *Cache
}

// NewHandler creates a Handler backed by the provided Cache.
func NewHandler(cache *Cache) *Handler {
	return &Handler{cache: cache}
}

// Check handles GET /api/stock-check.
// Query params: product (required), retailer (required).
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	product := strings.TrimSpace(r.URL.Query().Get("product"))
	retailer := strings.TrimSpace(r.URL.Query().Get("retailer"))

	if product == "" {
		httputil.WriteError(w, http.StatusBadRequest, "product query param is required")
		return
	}
	if retailer == "" {
		httputil.WriteError(w, http.StatusBadRequest, "retailer query param is required")
		return
	}

	key := product + "|" + retailer
	status, hit := h.cache.Get(key)
	if !hit {
		status = DetermineStatus(product, retailer)
		h.cache.Set(key, status)
	}

	httputil.WriteJSON(w, http.StatusOK, StockResponse{
		Status:    status,
		Retailer:  retailer,
		UpdatedAt: time.Now().UTC(),
	})
}
