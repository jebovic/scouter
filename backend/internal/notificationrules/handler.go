package notificationrules

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 5 * time.Minute

// cacheEntry holds a cached response with its expiry time.
type cacheEntry struct {
	response  AlertsResponse
	expiresAt time.Time
}

// Handler serves smart alert rule endpoints.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a new Handler with the given database pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

// GetAlerts handles GET /api/missions/{missionId}/smart-alerts.
// It evaluates all shopping items for the given mission against the smart alert
// rules and returns triggered alerts sorted by severity (critical first).
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId is required")
		return
	}

	// Check cache
	h.mu.Lock()
	if entry, ok := h.cache[missionID]; ok && time.Now().Before(entry.expiresAt) {
		resp := entry.response
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, resp)
		return
	}
	h.mu.Unlock()

	resp, err := h.evaluate(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Store in cache
	h.mu.Lock()
	h.cache[missionID] = cacheEntry{response: resp, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, resp)
}

const itemsQuery = `
	SELECT id::text, name, price, target_price, status, cost_category
	FROM shopping_items
	WHERE mission_id = $1
`

const historyQuery = `
	SELECT price
	FROM price_history
	WHERE item_id = $1::uuid
	ORDER BY recorded_at DESC
	LIMIT 10
`

func (h *Handler) evaluate(ctx context.Context, missionID string) (AlertsResponse, error) {
	rows, err := h.pool.Query(ctx, itemsQuery, missionID)
	if err != nil {
		return AlertsResponse{}, err
	}
	defer rows.Close()

	var items []shoppingItem
	for rows.Next() {
		var it shoppingItem
		if err := rows.Scan(&it.id, &it.name, &it.price, &it.targetPrice, &it.status, &it.costCategory); err != nil {
			return AlertsResponse{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return AlertsResponse{}, err
	}

	now := time.Now().UTC()
	var alerts []Alert

	for _, it := range items {
		// Fetch up to 10 price history entries (most recent first)
		hrows, err := h.pool.Query(ctx, historyQuery, it.id)
		if err != nil {
			return AlertsResponse{}, err
		}
		var history []pricePoint
		for hrows.Next() {
			var pp pricePoint
			if err := hrows.Scan(&pp.price); err != nil {
				hrows.Close()
				return AlertsResponse{}, err
			}
			history = append(history, pp)
		}
		hrows.Close()
		if err := hrows.Err(); err != nil {
			return AlertsResponse{}, err
		}

		triggered := evaluateRules(it, history, now)
		alerts = append(alerts, triggered...)
	}

	// Sort: critical first, then warning; within each severity preserve order
	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].Severity == alerts[j].Severity {
			return false
		}
		return alerts[i].Severity == "critical"
	})

	criticalCount := 0
	warningCount := 0
	for _, a := range alerts {
		if a.Severity == "critical" {
			criticalCount++
		} else {
			warningCount++
		}
	}

	if alerts == nil {
		alerts = []Alert{}
	}

	return AlertsResponse{
		MissionID:     missionID,
		Alerts:        alerts,
		CriticalCount: criticalCount,
		WarningCount:  warningCount,
	}, nil
}

// evaluateRules applies all smart alert rules to a single item and returns triggered alerts.
func evaluateRules(it shoppingItem, history []pricePoint, now time.Time) []Alert {
	var alerts []Alert

	latestPrice := it.price
	if len(history) > 0 {
		latestPrice = history[0].price
	}

	// Rule 1: TARGET_HIT — price <= target_price
	if it.targetPrice != nil && latestPrice <= *it.targetPrice {
		alerts = append(alerts, Alert{
			ItemID:      it.id,
			ItemName:    it.name,
			RuleID:      "TARGET_HIT",
			Severity:    "critical",
			Message:     "🎯 Objectif atteint pour " + it.name + " (" + formatPrice(latestPrice) + "€)",
			TriggeredAt: now,
		})
	} else if it.targetPrice != nil && latestPrice <= *it.targetPrice*1.05 {
		// Rule 2: NEAR_TARGET — price <= target_price * 1.05 (only if TARGET_HIT not triggered)
		alerts = append(alerts, Alert{
			ItemID:      it.id,
			ItemName:    it.name,
			RuleID:      "NEAR_TARGET",
			Severity:    "warning",
			Message:     "⚡ " + it.name + " à 5% de votre objectif",
			TriggeredAt: now,
		})
	}

	// Rule 3: FLASH_SALE — status == "flash-sale"
	if it.status == "flash-sale" {
		alerts = append(alerts, Alert{
			ItemID:      it.id,
			ItemName:    it.name,
			RuleID:      "FLASH_SALE",
			Severity:    "critical",
			Message:     "🔥 Vente flash sur " + it.name + " — agissez vite !",
			TriggeredAt: now,
		})
	}

	// History-based rules require at least 2 entries
	if len(history) >= 2 {
		currentPrice := history[0].price
		prevPrice := history[1].price

		// Rule 4: PRICE_DROP_10 — latest < price 7 entries ago * 0.90
		if len(history) >= 8 {
			ref := history[7].price
			if currentPrice < ref*0.90 {
				alerts = append(alerts, Alert{
					ItemID:      it.id,
					ItemName:    it.name,
					RuleID:      "PRICE_DROP_10",
					Severity:    "warning",
					Message:     "📉 " + it.name + " a baissé de >10% cette semaine",
					TriggeredAt: now,
				})
			}
		}

		// Rule 5: ALL_TIME_LOW — latest == min of all 10
		minPrice := history[0].price
		for _, p := range history[1:] {
			if p.price < minPrice {
				minPrice = p.price
			}
		}
		if currentPrice <= minPrice {
			alerts = append(alerts, Alert{
				ItemID:      it.id,
				ItemName:    it.name,
				RuleID:      "ALL_TIME_LOW",
				Severity:    "critical",
				Message:     "🏆 " + it.name + " au prix le plus bas observé",
				TriggeredAt: now,
			})
		}

		// Rule 6: PRICE_SPIKE — latest > previous * 1.15
		if currentPrice > prevPrice*1.15 {
			alerts = append(alerts, Alert{
				ItemID:      it.id,
				ItemName:    it.name,
				RuleID:      "PRICE_SPIKE",
				Severity:    "warning",
				Message:     "⬆️ " + it.name + " a augmenté de >15%",
				TriggeredAt: now,
			})
		}

		// Rule 7: DEFER_OPPORTUNITY — status == "defer" AND latest < avg * 0.95
		if it.status == "defer" {
			sum := 0.0
			for _, p := range history {
				sum += p.price
			}
			avg := sum / float64(len(history))
			if currentPrice < avg*0.95 {
				alerts = append(alerts, Alert{
					ItemID:      it.id,
					ItemName:    it.name,
					RuleID:      "DEFER_OPPORTUNITY",
					Severity:    "warning",
					Message:     "💡 " + it.name + " différé mais prix bas — reconsidérez ?",
					TriggeredAt: now,
				})
			}
		}
	}

	return alerts
}

// formatPrice formats a float64 price with up to 2 decimal places, trimming trailing zeros.
func formatPrice(v float64) string {
	// Use sprintf with %g to avoid unnecessary decimal places
	s := ""
	if v == float64(int64(v)) {
		// Whole number
		n := int64(v)
		s = formatInt(n)
	} else {
		// Format with 2 decimal places, trim trailing zeros
		s = trimTrailingZeros(v)
	}
	return s
}

func formatInt(n int64) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	if n < 1000 {
		return itoa(n)
	}
	return formatInt(n/1000) + "," + pad3(n%1000)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

func pad3(n int64) string {
	s := itoa(n)
	for len(s) < 3 {
		s = "0" + s
	}
	return s
}

func trimTrailingZeros(v float64) string {
	// Format to 2 decimal places
	buf := make([]byte, 0, 16)
	buf = appendFloat(buf, v)
	return string(buf)
}

func appendFloat(buf []byte, v float64) []byte {
	// Simple 2-decimal formatting
	neg := v < 0
	if neg {
		v = -v
		buf = append(buf, '-')
	}
	intPart := int64(v)
	fracPart := int64((v-float64(intPart))*100 + 0.5)
	if fracPart >= 100 {
		intPart++
		fracPart -= 100
	}
	buf = append(buf, []byte(itoa(intPart))...)
	if fracPart != 0 {
		buf = append(buf, '.')
		if fracPart < 10 {
			buf = append(buf, '0')
		}
		buf = append(buf, []byte(itoa(fracPart))...)
	}
	return buf
}
