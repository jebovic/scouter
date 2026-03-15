package listoptimizer

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler provides smart shopping list optimization.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // string -> cacheEntry
}

type cacheEntry struct {
	list      *OptimizedList
	expiresAt time.Time
}

// NewHandler creates a new list optimizer handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetOptimizedList handles GET /api/missions/{missionId}/optimized-list
func (h *Handler) GetOptimizedList(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")

	// Validate UUID
	_, err := uuid.Parse(missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission ID format")
		return
	}

	// Check cache
	if cached, ok := h.cache.Load(missionID); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.list)
			return
		}
		h.cache.Delete(missionID)
	}

	// Query shopping items
	rows, err := h.pool.Query(r.Context(),
		`SELECT id, name, price, merchant, status FROM shopping_items
		 WHERE mission_id = $1 AND status NOT IN ('defer')
		 ORDER BY price DESC`,
		missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	// Parse items into groups by merchant
	merchantMap := make(map[string][]ItemRef)
	totalPrice := 0.0
	totalItems := 0

	for rows.Next() {
		var id, name, merchant, status string
		var price float64
		if err := rows.Scan(&id, &name, &price, &merchant, &status); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		totalPrice += price
		totalItems++
		merchantLower := strings.ToLower(strings.TrimSpace(merchant))
		merchantMap[merchantLower] = append(merchantMap[merchantLower], ItemRef{
			ItemID:   id,
			Name:     name,
			Price:    price,
			Status:   status,
			Priority: 0, // will be assigned per group
		})
	}

	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Build merchant groups
	groups := make([]MerchantGroup, 0, len(merchantMap))
	for merchant, items := range merchantMap {
		// Sort items by price descending within the merchant group
		sort.Slice(items, func(i, j int) bool {
			return items[i].Price > items[j].Price
		})

		// Assign priorities 1..N by price descending
		for i := range items {
			items[i].Priority = i + 1
		}

		// Calculate merchant batch total price
		merchantTotal := 0.0
		for _, item := range items {
			merchantTotal += item.Price
		}

		// Trip saving: 5% of merchant batch price (conceptual savings from batching)
		tripSaving := merchantTotal * 0.05

		groups = append(groups, MerchantGroup{
			Merchant:   merchant,
			Items:      items,
			TotalPrice: merchantTotal,
			TripSaving: tripSaving,
		})
	}

	// Sort merchant groups by total price descending (most expensive batch first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalPrice > groups[j].TotalPrice
	})

	estimatedTrips := len(groups)

	// Generate French summary
	summary := fmt.Sprintf("🛒 Liste optimisée pour %d trajets | Total : %.2f€ | Économies estimées : %.2f€",
		estimatedTrips, totalPrice, totalPrice*0.05)

	result := &OptimizedList{
		MissionID:      missionID,
		MerchantGroups: groups,
		TotalItems:     totalItems,
		TotalPrice:     totalPrice,
		EstimatedTrips: estimatedTrips,
		Summary:        summary,
	}

	// Cache for 15 minutes
	h.cache.Store(missionID, cacheEntry{
		list:      result,
		expiresAt: time.Now().Add(15 * time.Minute),
	})

	httputil.WriteJSON(w, http.StatusOK, result)
}
