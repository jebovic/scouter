package pricealertdigest

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves the price alert digest endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]cachedDigest
	ttl   time.Duration
}

type cachedDigest struct {
	digest    *PriceAlertDigest
	timestamp time.Time
}

// NewHandler creates a new Handler backed by the given pgx pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
		ttl:  5 * time.Minute,
	}
}

// itemRow holds a raw row from the items and price history queries.
type itemRow struct {
	itemID      string
	name        string
	price       float64
	targetPrice *float64
	status      string
	minPrice    *float64
	avgPrice    *float64
	histCount   int64
}

// GetDigest handles GET /api/missions/{missionId}/price-alert-digest.
// It aggregates price alerts: target_hit, below_avg, trending_down.
func (h *Handler) GetDigest(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId required")
		return
	}

	// Parse UUID to validate format
	if _, err := uuid.Parse(missionID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid missionId")
		return
	}

	// Check cache
	if cached, ok := h.cache.Load(missionID); ok {
		c := cached.(cachedDigest)
		if time.Since(c.timestamp) < h.ttl {
			httputil.WriteJSON(w, http.StatusOK, c.digest)
			return
		}
	}

	// Query shopping items for the mission
	const itemsQuery = `
	SELECT
		si.id::text,
		si.name,
		si.price::float8,
		si.target_price::float8,
		si.status
	FROM shopping_items si
	WHERE si.mission_id = $1
	  AND si.status NOT IN ('buy', 'defer')
	ORDER BY si.created_at
	`

	rows, err := h.pool.Query(r.Context(), itemsQuery, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(&it.itemID, &it.name, &it.price, &it.targetPrice, &it.status); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// For each item, query price history stats
	const priceHistoryQuery = `
	SELECT
		MIN(price)::float8,
		AVG(price)::float8,
		COUNT(*)::bigint
	FROM price_history
	WHERE item_id = $1
	`

	for i := range items {
		err := h.pool.QueryRow(r.Context(), priceHistoryQuery, items[i].itemID).
			Scan(&items[i].minPrice, &items[i].avgPrice, &items[i].histCount)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// If no history, minPrice/avgPrice remain nil
	}

	// Generate alerts
	digest := buildAlertDigest(missionID, items)

	// Cache result
	h.cache.Store(missionID, cachedDigest{
		digest:    digest,
		timestamp: time.Now(),
	})

	httputil.WriteJSON(w, http.StatusOK, digest)
}

// buildAlertDigest classifies items into alerts and constructs the response.
func buildAlertDigest(missionID string, items []itemRow) *PriceAlertDigest {
	var alerts []AlertEntry
	alertTypes := make(map[string]bool) // Track which types we've seen

	for _, item := range items {
		// 1. target_hit: if targetPrice is set and current price <= target
		if item.targetPrice != nil && item.price <= *item.targetPrice {
			savings := *item.targetPrice - item.price
			alerts = append(alerts, AlertEntry{
				ItemID:    item.itemID,
				Name:      item.name,
				Price:     item.price,
				AlertType: "target_hit",
				Message:   fmt.Sprintf("🎯 Objectif atteint! Le prix est descendu à %.2f€ (cible: %.2f€)", item.price, *item.targetPrice),
				Savings:   savings,
			})
			alertTypes["target_hit"] = true
			continue
		}

		// 2. below_avg: if avgPrice is available and price < avgPrice * 0.9 (10% below average)
		if item.avgPrice != nil && *item.avgPrice > 0 && item.price < *item.avgPrice*0.9 {
			savings := *item.avgPrice - item.price
			alerts = append(alerts, AlertEntry{
				ItemID:    item.itemID,
				Name:      item.name,
				Price:     item.price,
				AlertType: "below_avg",
				Message:   fmt.Sprintf("📉 Prix 10%% sous la moyenne (%.2f€ vs %.2f€ en moyenne)", item.price, *item.avgPrice),
				Savings:   savings,
			})
			alertTypes["below_avg"] = true
			continue
		}

		// 3. trending_down: if histCount >= 3 and price <= minPrice * 1.05 (within 5% of all-time low)
		if item.histCount >= 3 && item.minPrice != nil && item.price <= *item.minPrice*1.05 {
			var savings float64
			if item.avgPrice != nil {
				savings = *item.avgPrice - item.price
			}
			alerts = append(alerts, AlertEntry{
				ItemID:    item.itemID,
				Name:      item.name,
				Price:     item.price,
				AlertType: "trending_down",
				Message:   fmt.Sprintf("⬇️ Proche du prix plancher historique (%.2f€)", *item.minPrice),
				Savings:   savings,
			})
			alertTypes["trending_down"] = true
		}
	}

	// Sort alerts: target_hit first, then below_avg, then trending_down
	sort.Slice(alerts, func(i, j int) bool {
		typeOrder := map[string]int{"target_hit": 0, "below_avg": 1, "trending_down": 2}
		ti := typeOrder[alerts[i].AlertType]
		tj := typeOrder[alerts[j].AlertType]
		if ti != tj {
			return ti < tj
		}
		// Within same type, sort by savings descending
		return alerts[i].Savings > alerts[j].Savings
	})

	// Build summary
	var summary string
	if len(alerts) > 0 {
		summary = fmt.Sprintf("🔔 %d alerte(s) de prix pour cette mission", len(alerts))
	} else {
		summary = "✅ Aucune alerte de prix active"
	}

	return &PriceAlertDigest{
		MissionID:   missionID,
		AlertCount:  len(alerts),
		Alerts:      alerts,
		Summary:     summary,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// Helper to safely format float64 with specific precision
func formatPrice(v float64) string {
	return fmt.Sprintf("%.2f", math.Round(v*100)/100)
}
