package priceinsights

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

var frenchDays = [7]string{
	"dimanche", // 0 = Sunday
	"lundi",    // 1 = Monday
	"mardi",    // 2 = Tuesday
	"mercredi", // 3 = Wednesday
	"jeudi",    // 4 = Thursday
	"vendredi", // 5 = Friday
	"samedi",   // 6 = Saturday
}

type cacheEntry struct {
	insights  PriceInsights
	expiresAt time.Time
}

// Handler exposes the price-insights endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetInsights handles GET /api/missions/{missionId}/items/{itemId}/price-insights.
func (h *Handler) GetInsights(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	// Check cache.
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.insights)
			return
		}
	}

	insights, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		insights:  insights,
		expiresAt: time.Now().Add(2 * time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, insights)
}

// compute queries price_history and derives PriceInsights.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (PriceInsights, error) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT price, EXTRACT(DOW FROM created_at)::int AS dow, EXTRACT(DAY FROM created_at)::int AS dom
		FROM price_history
		WHERE item_id = $1
		ORDER BY created_at ASC`, itemID)
	if err != nil {
		return PriceInsights{}, err
	}
	defer rows.Close()

	type record struct {
		price float64
		dow   int // 0=Sunday … 6=Saturday
		dom   int // 1–31
	}

	var records []record
	for rows.Next() {
		var rec record
		if err := rows.Scan(&rec.price, &rec.dow, &rec.dom); err != nil {
			return PriceInsights{}, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return PriceInsights{}, err
	}

	n := len(records)
	ins := PriceInsights{
		ItemID:      itemID.String(),
		DataPoints:  n,
		GeneratedAt: time.Now().UTC(),
	}

	if n < 3 {
		ins.Confidence = "low"
		ins.BuyRecommendation = "Pas assez de données pour déterminer le meilleur moment d'achat. Continuez à surveiller les prix."
		ins.PriceByDay = make([]DayStats, 0)
		return ins, nil
	}

	// Aggregate by day of week.
	type agg struct {
		sum   float64
		count int
	}
	byDow := [7]agg{}
	byPeriod := [3]agg{} // 0=début, 1=milieu, 2=fin

	for _, rec := range records {
		if rec.dow >= 0 && rec.dow < 7 {
			byDow[rec.dow].sum += rec.price
			byDow[rec.dow].count++
		}
		period := domToPeriod(rec.dom)
		byPeriod[period].sum += rec.price
		byPeriod[period].count++
	}

	// Build PriceByDay — only include days that have data.
	dayStats := make([]DayStats, 0, 7)
	bestDow, worstDow := -1, -1
	bestAvg, worstAvg := 0.0, 0.0

	for i := 0; i < 7; i++ {
		if byDow[i].count == 0 {
			continue
		}
		avg := byDow[i].sum / float64(byDow[i].count)
		dayStats = append(dayStats, DayStats{
			Day:      frenchDays[i],
			AvgPrice: avg,
			Count:    byDow[i].count,
		})
		if bestDow == -1 || avg < bestAvg {
			bestAvg = avg
			bestDow = i
		}
		if worstDow == -1 || avg > worstAvg {
			worstAvg = avg
			worstDow = i
		}
	}
	ins.PriceByDay = dayStats

	if bestDow >= 0 {
		ins.BestDayOfWeek = frenchDays[bestDow]
	}
	if worstDow >= 0 {
		ins.WorstDayOfWeek = frenchDays[worstDow]
	}

	// Best period of month.
	periodNames := [3]string{"début", "milieu", "fin"}
	bestPeriod := 0
	bestPeriodAvg := -1.0
	for i := 0; i < 3; i++ {
		if byPeriod[i].count == 0 {
			continue
		}
		avg := byPeriod[i].sum / float64(byPeriod[i].count)
		if bestPeriodAvg < 0 || avg < bestPeriodAvg {
			bestPeriodAvg = avg
			bestPeriod = i
		}
	}
	ins.BestTimeOfMonth = periodNames[bestPeriod]

	// Confidence level.
	switch {
	case n >= 7:
		ins.Confidence = "high"
	case n >= 5:
		ins.Confidence = "medium"
	default:
		ins.Confidence = "low"
	}

	// French buy recommendation.
	ins.BuyRecommendation = buildRecommendation(ins.BestDayOfWeek, ins.BestTimeOfMonth, ins.Confidence)

	return ins, nil
}

// domToPeriod maps a day-of-month (1–31) to a period index: 0=début, 1=milieu, 2=fin.
func domToPeriod(dom int) int {
	switch {
	case dom <= 10:
		return 0
	case dom <= 20:
		return 1
	default:
		return 2
	}
}

// buildRecommendation returns a French buy recommendation string.
func buildRecommendation(bestDay, bestPeriod, confidence string) string {
	if confidence == "low" {
		return "Continuez à surveiller les prix pour obtenir une recommandation plus précise."
	}
	return fmt.Sprintf(
		"Le meilleur moment pour acheter est généralement le %s, en %s de mois. Les prix sont statistiquement plus bas à ce moment-là.",
		bestDay, bestPeriod,
	)
}
