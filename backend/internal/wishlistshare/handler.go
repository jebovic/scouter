package wishlistshare

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 5 * time.Minute

// cacheEntry holds a cached wishlist card and its expiry time.
type cacheEntry struct {
	data   WishlistShareCard
	expiry time.Time
}

// Handler serves the wishlist card endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

// NewHandler creates a Handler backed by the given pgx pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

const wishlistCardQuery = `
SELECT
    m.id::text      AS mission_id,
    m.name          AS mission_name,
    m.budget,
    m.currency,
    si.name         AS item_name,
    si.price,
    si.status,
    si.merchant
FROM missions m
LEFT JOIN shopping_items si ON si.mission_id = m.id
WHERE m.id = $1
  AND m.archived_at IS NULL
ORDER BY si.price DESC NULLS LAST
`

// GetWishlistCard handles GET /api/missions/{id}/wishlist-card.
// It returns a shareable card data structure for the mission's shopping list.
func (h *Handler) GetWishlistCard(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing mission id")
		return
	}

	// Check cache first.
	if v, ok := h.cache.Load(missionID); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiry) {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		h.cache.Delete(missionID)
	}

	rows, err := h.pool.Query(r.Context(), wishlistCardQuery, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var (
		mID      string
		mName    string
		budget   float64
		currency string
		items    []WishlistItem
		found    bool
	)

	for rows.Next() {
		var (
			rowMissionID   string
			rowMissionName string
			rowBudget      float64
			rowCurrency    string
			itemName       *string
			itemPrice      *float64
			itemStatus     *string
			itemMerchant   *string
		)
		if err := rows.Scan(
			&rowMissionID,
			&rowMissionName,
			&rowBudget,
			&rowCurrency,
			&itemName,
			&itemPrice,
			&itemStatus,
			&itemMerchant,
		); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if !found {
			mID = rowMissionID
			mName = rowMissionName
			budget = rowBudget
			currency = rowCurrency
			found = true
		}

		// itemName is NULL when there are no shopping items (LEFT JOIN with no rows).
		if itemName != nil {
			name := *itemName
			price := 0.0
			if itemPrice != nil {
				price = *itemPrice
			}
			status := ""
			if itemStatus != nil {
				status = *itemStatus
			}
			merchant := ""
			if itemMerchant != nil {
				merchant = *itemMerchant
			}
			items = append(items, WishlistItem{
				Name:     name,
				Price:    price,
				Status:   status,
				Merchant: merchant,
			})
		}
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if !found {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	shareURL := buildShareURL(r, mID)

	card := WishlistShareCard{
		MissionID:   mID,
		MissionName: mName,
		TotalBudget: budget,
		Currency:    currency,
		Items:       items,
		ShareURL:    shareURL,
		GeneratedAt: time.Now().UTC(),
	}

	h.cache.Store(missionID, cacheEntry{
		data:   card,
		expiry: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, card)
}

// buildShareURL constructs a share URL from the incoming request's host.
func buildShareURL(r *http.Request, missionID string) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	return fmt.Sprintf("%s://%s/missions/%s/shopping?share=wishlist", scheme, host, missionID)
}
