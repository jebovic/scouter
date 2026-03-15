package priceanalytics

import (
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// priceRow holds a single price observation queried from price_history.
type priceRow struct {
	price      float64
	recordedAt time.Time
}

// Handler exposes the price-stats endpoint.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetAnalytics handles GET /api/missions/{missionID}/price-analytics.
// It aggregates price history for all items in the mission and returns computed stats.
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT si.id, si.name, ph.price, ph.recorded_at
		FROM shopping_items si
		LEFT JOIN price_history ph ON ph.item_id = si.id
		WHERE si.mission_id = $1
		ORDER BY si.id, ph.recorded_at ASC`, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	type itemKey struct {
		id   string
		name string
	}
	type itemData struct {
		name   string
		prices []float64
	}
	// Use ordered slice to preserve deterministic output.
	var order []string
	byID := make(map[string]*itemData)

	for rows.Next() {
		var (
			itemID   uuid.UUID
			itemName string
			rawPrice *float64
			rawRecAt *time.Time
		)
		// price and recorded_at are nullable due to LEFT JOIN.
		if err := rows.Scan(&itemID, &itemName, &rawPrice, &rawRecAt); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		_ = rawRecAt
		id := itemID.String()
		if _, exists := byID[id]; !exists {
			byID[id] = &itemData{name: itemName}
			order = append(order, id)
		}
		if rawPrice != nil {
			byID[id].prices = append(byID[id].prices, *rawPrice)
		}
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	result := MissionPriceAnalytics{
		MissionID: missionID.String(),
		Items:     make([]PriceStats, 0, len(order)),
	}

	bestDealRatio := 0.0
	bestDealSet := false
	mostVolPct := 0.0

	for _, id := range order {
		d := byID[id]
		ps := computeAnalytics(id, d.name, d.prices)
		result.Items = append(result.Items, ps)

		if ps.DataPoints > 0 && ps.Average > 0 {
			ratio := (ps.Current - ps.Average) / ps.Average
			if !bestDealSet || ratio < bestDealRatio {
				bestDealRatio = ratio
				result.BestDealItem = ps.ItemName
				bestDealSet = true
			}
			if ps.VolatilityPct > mostVolPct {
				mostVolPct = ps.VolatilityPct
				result.MostVolatile = ps.ItemName
			}
		}
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// computeAnalytics derives the mission-level PriceStats from raw price values.
func computeAnalytics(itemID, itemName string, prices []float64) PriceStats {
	n := len(prices)
	ps := PriceStats{
		ItemID:   itemID,
		ItemName: itemName,
		Trend:    TrendStable,
	}
	if n == 0 {
		return ps
	}

	minVal, maxVal := prices[0], prices[0]
	sum := 0.0
	for _, p := range prices {
		if p < minVal {
			minVal = p
		}
		if p > maxVal {
			maxVal = p
		}
		sum += p
	}
	avg := sum / float64(n)

	varianceSum := 0.0
	for _, p := range prices {
		d := p - avg
		varianceSum += d * d
	}
	stdDev := 0.0
	if n > 1 {
		stdDev = math.Sqrt(varianceSum / float64(n))
	}

	volPct := 0.0
	if avg > 0 {
		volPct = stdDev / avg * 100
	}

	trend := TrendStable
	if n >= 2 {
		last := prices[n-1]
		secondLast := prices[n-2]
		if secondLast > 0 {
			ratio := last / secondLast
			if ratio > 1.01 {
				trend = TrendRising
			} else if ratio < 0.99 {
				trend = TrendFalling
			}
		}
	}

	ps.Min = minVal
	ps.Max = maxVal
	ps.Average = avg
	ps.Current = prices[n-1]
	ps.VolatilityPct = volPct
	ps.Trend = trend
	ps.DataPoints = n
	return ps
}

// GetStats handles GET /api/missions/{missionID}/items/{itemID}/price-stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT price, recorded_at
		FROM price_history
		WHERE item_id = $1
		ORDER BY recorded_at ASC`, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var pts []priceRow
	for rows.Next() {
		var p priceRow
		if err := rows.Scan(&p.price, &p.recordedAt); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		pts = append(pts, p)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(pts) == 0 {
		httputil.WriteJSON(w, http.StatusOK, PriceStats{ItemID: itemID.String()})
		return
	}

	stats := compute(itemID.String(), pts)
	httputil.WriteJSON(w, http.StatusOK, stats)
}

// compute derives PriceStats from raw price observations.
func compute(itemID string, pts []priceRow) PriceStats {
	n := len(pts)
	if n == 0 {
		return PriceStats{ItemID: itemID}
	}
	prices := make([]float64, n)
	for i, p := range pts {
		prices[i] = p.price
	}

	// Min / max / average
	minVal, maxVal := prices[0], prices[0]
	sum := 0.0
	for _, p := range prices {
		if p < minVal {
			minVal = p
		}
		if p > maxVal {
			maxVal = p
		}
		sum += p
	}
	avg := sum / float64(n)

	// Standard deviation (population)
	varianceSum := 0.0
	for _, p := range prices {
		d := p - avg
		varianceSum += d * d
	}
	stdDev := 0.0
	if n > 1 {
		stdDev = math.Sqrt(varianceSum / float64(n))
	}

	// Linear regression slope (x = index, y = price)
	slope := 0.0
	if n >= 2 {
		// Use simple index as x so we don't need to convert timestamps.
		xMean := float64(n-1) / 2.0
		num, den := 0.0, 0.0
		for i, p := range prices {
			dx := float64(i) - xMean
			num += dx * (p - avg)
			den += dx * dx
		}
		if den != 0 {
			slope = num / den
		}
	}

	// Price change vs 7 days ago
	drop7d := 0.0
	cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
	// Find the latest snapshot that is at least 7 days old.
	baseline := -1
	for i := n - 1; i >= 0; i-- {
		if !pts[i].recordedAt.After(cutoff) {
			baseline = i
			break
		}
	}
	if baseline >= 0 {
		// Positive drop7d means the current price is higher than the baseline (more expensive).
		// Negative drop7d means the price fell (cheaper).
		drop7d = prices[n-1] - prices[baseline]
	}

	return PriceStats{
		ItemID:      itemID,
		Min:         minVal,
		Max:         maxVal,
		Average:     avg,
		Volatility:  stdDev,
		DataPoints:  n,
		TrendSlope:  slope,
		PriceDrop7d: drop7d,
	}
}
