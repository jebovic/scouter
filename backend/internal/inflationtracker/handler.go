package inflationtracker

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	impact    InflationImpact
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/inflation-impact.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

// NewHandler creates a Handler wired to the given database pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// itemRow holds raw columns from the shopping_items query.
type itemRow struct {
	id           string
	name         string
	price        float64
	costCategory string
	status       string
}

// GetInflationImpact handles GET /api/missions/{missionId}/inflation-impact.
func (h *Handler) GetInflationImpact(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "missionId")
	if idParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing missionId")
		return
	}

	// Validate that the param is a UUID; if not, return 400.
	missionID, err := uuid.Parse(idParam)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "missionId must be a UUID")
		return
	}

	cacheKey := missionID.String()
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.impact)
			return
		}
	}

	ctx := r.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT id, name, price, COALESCE(cost_category, '') AS cost_category, status
		FROM shopping_items
		WHERE mission_id = $1`, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query shopping items")
		return
	}
	defer rows.Close()

	var rawItems []itemRow
	for rows.Next() {
		var it itemRow
		if scanErr := rows.Scan(&it.id, &it.name, &it.price, &it.costCategory, &it.status); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan shopping item row")
			return
		}
		rawItems = append(rawItems, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate shopping item rows")
		return
	}

	impact := compute(missionID.String(), rawItems)

	h.cache.Store(cacheKey, cacheEntry{impact: impact, expiresAt: time.Now().Add(cacheTTL)})

	httputil.WriteJSON(w, http.StatusOK, impact)
}

// compute derives InflationImpact from raw item rows. Pure function.
func compute(missionID string, rawItems []itemRow) InflationImpact {
	items := make([]InflationItem, 0, len(rawItems))

	var totalWaitCost, sumRate float64

	for _, raw := range rawItems {
		rate := rateForCategory(raw.costCategory)
		monthly := raw.price * rate / 12
		wait3 := raw.price * (1 + rate*3/12) - raw.price

		items = append(items, InflationItem{
			ItemID:              raw.id,
			Name:                raw.name,
			Price:               raw.price,
			AnnualInflationRate: rate,
			InflationPer30Days:  monthly,
			WaitCost3Months:     wait3,
			Recommendation:      recommendationFor(rate),
		})

		totalWaitCost += wait3
		sumRate += rate
	}

	var avgRate float64
	if len(rawItems) > 0 {
		avgRate = sumRate / float64(len(rawItems))
	}

	if items == nil {
		items = []InflationItem{}
	}

	return InflationImpact{
		MissionID:            missionID,
		Items:                items,
		TotalWaitCost3Months: totalWaitCost,
		AvgInflationRate:     avgRate,
	}
}
