package pricefloor

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

type cacheEntry struct {
	floor     PriceFloor
	expiresAt time.Time
}

// Handler exposes the price-floor endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetFloor handles GET /api/missions/{missionId}/items/{itemId}/price-floor.
func (h *Handler) GetFloor(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	// Check cache (2h TTL).
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.floor)
			return
		}
	}

	floor, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		floor:     floor,
		expiresAt: time.Now().Add(2 * time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, floor)
}

// compute queries price_history, fetches the current price from shopping_items,
// and returns a PriceFloor prediction.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (PriceFloor, error) {
	ctx := r.Context()

	// Fetch current price from shopping_items.
	var currentPrice float64
	err := h.pool.QueryRow(ctx,
		`SELECT price FROM shopping_items WHERE id = $1`, itemID,
	).Scan(&currentPrice)
	if err != nil {
		// Fall back to 0 — predict will still work with just history.
		currentPrice = 0
	}

	// Load price history ordered by time.
	rows, err := h.pool.Query(ctx,
		`SELECT price FROM price_history WHERE item_id = $1 ORDER BY created_at ASC`, itemID)
	if err != nil {
		return PriceFloor{}, err
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			return PriceFloor{}, err
		}
		prices = append(prices, p)
	}
	if err := rows.Err(); err != nil {
		return PriceFloor{}, err
	}

	// If we have history and current price is 0, use the last known price.
	if currentPrice == 0 && len(prices) > 0 {
		currentPrice = prices[len(prices)-1]
	}

	return predict(itemID.String(), currentPrice, prices), nil
}
