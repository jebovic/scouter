package volatilitycalendar

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 2 * time.Hour

type cacheEntry struct {
	cal       VolatilityCalendar
	expiresAt time.Time
}

// Handler exposes the volatility-calendar endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetCalendar handles GET /api/missions/{missionId}/items/{itemId}/volatility-calendar.
func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.cal)
			return
		}
	}

	cal, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		cal:       cal,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, cal)
}

type priceRecord struct {
	price      float64
	recordedAt time.Time
}

// compute fetches 90 days of price history and builds the volatility calendar.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (VolatilityCalendar, error) {
	ctx := r.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT price, recorded_at
		FROM price_history
		WHERE item_id = $1
		  AND recorded_at >= NOW() - INTERVAL '90 days'
		ORDER BY recorded_at ASC`, itemID)
	if err != nil {
		return VolatilityCalendar{}, err
	}
	defer rows.Close()

	var records []priceRecord
	for rows.Next() {
		var rec priceRecord
		if err := rows.Scan(&rec.price, &rec.recordedAt); err != nil {
			return VolatilityCalendar{}, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return VolatilityCalendar{}, err
	}

	cal := VolatilityCalendar{
		ItemID: itemID.String(),
		Weeks:  []WeekData{},
	}

	if len(records) < 2 {
		return cal, nil
	}

	// Group by (year, ISO week).
	type weekKey struct {
		year int
		week int
	}
	type weekAgg struct {
		min float64
		max float64
	}

	agg := make(map[weekKey]*weekAgg)
	order := []weekKey{} // preserve insertion order

	for _, rec := range records {
		year, week := rec.recordedAt.ISOWeek()
		k := weekKey{year: year, week: week}
		if _, exists := agg[k]; !exists {
			agg[k] = &weekAgg{min: rec.price, max: rec.price}
			order = append(order, k)
		} else {
			if rec.price < agg[k].min {
				agg[k].min = rec.price
			}
			if rec.price > agg[k].max {
				agg[k].max = rec.price
			}
		}
	}

	// Compute raw ranges and max range for normalization.
	type rawWeek struct {
		key   weekKey
		agg   *weekAgg
		rangeV float64
	}
	rawWeeks := make([]rawWeek, 0, len(order))
	maxRange := 0.0

	for _, k := range order {
		a := agg[k]
		rng := a.max - a.min
		rawWeeks = append(rawWeeks, rawWeek{key: k, agg: a, rangeV: rng})
		if rng > maxRange {
			maxRange = rng
		}
	}

	// Build week data with normalized intensity.
	weeks := make([]WeekData, 0, len(rawWeeks))
	for _, rw := range rawWeeks {
		intensity := 0.0
		if maxRange > 0 {
			intensity = rw.rangeV / maxRange
			intensity = math.Round(intensity*100) / 100
		}
		wd := WeekData{
			Week:      rw.key.week,
			Year:      rw.key.year,
			Intensity: intensity,
			MinPrice:  math.Round(rw.agg.min*100) / 100,
			MaxPrice:  math.Round(rw.agg.max*100) / 100,
			Label:     fmt.Sprintf("Sem. %d", rw.key.week),
		}
		weeks = append(weeks, wd)
	}

	cal.Weeks = weeks

	// Find most volatile week (intensity == 1.0 means maxRange).
	var mostVolatile *WeekData
	for i := range weeks {
		if mostVolatile == nil || weeks[i].Intensity > mostVolatile.Intensity {
			mostVolatile = &weeks[i]
		}
	}

	if mostVolatile != nil && mostVolatile.Intensity > 0 {
		wCopy := *mostVolatile
		cal.MostVolatileWeek = &wCopy
		cal.Recommendation = fmt.Sprintf(
			"Évitez d'acheter en semaine %d (forte volatilité)",
			mostVolatile.Week,
		)
	}

	return cal, nil
}
