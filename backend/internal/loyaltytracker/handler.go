package loyaltytracker

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 15 * time.Minute

type merchantTotal struct {
	merchant string
	total    float64
}

type cacheEntry struct {
	summary  *LoyaltySummary
	cachedAt time.Time
}

// Handler serves GET /api/missions/{missionId}/loyalty-summary.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a Handler backed by pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

const merchantTotalsQuery = `
SELECT merchant, SUM(price) AS total
FROM shopping_items
WHERE mission_id = $1
  AND merchant IS NOT NULL
  AND merchant <> ''
GROUP BY merchant
ORDER BY total DESC
`

// GetLoyaltySummary handles GET /api/missions/{missionId}/loyalty-summary.
func (h *Handler) GetLoyaltySummary(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId is required")
		return
	}

	h.mu.Lock()
	if entry, ok := h.cache[missionID]; ok && time.Since(entry.cachedAt) < cacheTTL {
		result := entry.summary
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, result)
		return
	}
	h.mu.Unlock()

	rows, err := h.pool.Query(r.Context(), merchantTotalsQuery, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var totals []merchantTotal
	for rows.Next() {
		var mt merchantTotal
		if err := rows.Scan(&mt.merchant, &mt.total); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		totals = append(totals, mt)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	summary := buildSummary(totals)

	h.mu.Lock()
	h.cache[missionID] = cacheEntry{summary: summary, cachedAt: time.Now()}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, summary)
}

func buildSummary(rows []merchantTotal) *LoyaltySummary {
	merchants := make([]MerchantSummary, 0, len(rows))
	var totalCashback float64

	for _, rr := range rows {
		prog, ok := merchantRegistry[rr.merchant]
		if !ok {
			prog = defaultProgram
		}

		cashback := rr.total * prog.CashbackPct / 100
		points := int64(math.Round(rr.total * prog.PointRate))

		merchants = append(merchants, MerchantSummary{
			Merchant:          rr.merchant,
			Total:             rr.total,
			Program:           prog.Program,
			Tier:              prog.Tier,
			EstimatedCashback: cashback,
			EstimatedPoints:   points,
		})
		totalCashback += cashback
	}

	return &LoyaltySummary{
		Merchants:              merchants,
		TotalEstimatedCashback: totalCashback,
	}
}
