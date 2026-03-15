package roicalculator

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	roi       MissionROI
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/roi.
type Handler struct {
	pool        *pgxpool.Pool
	missionRepo mission.Repository
	cache       sync.Map
}

// NewHandler creates a Handler wired to the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:        pool,
		missionRepo: mission.NewRepository(pool),
	}
}

// itemRow holds the raw columns from the shopping_items query.
type itemRow struct {
	price           float64
	originalEstimate *float64
	targetPrice      *float64
	status          string
}

// GetROI handles GET /api/missions/{missionId}/roi.
// The {missionId} param may be a UUID or a slug.
func (h *Handler) GetROI(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "missionId")
	if idParam == "" {
		idParam = chi.URLParam(r, "id")
	}
	if idParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing missionId")
		return
	}

	ctx := r.Context()

	var m *mission.Mission
	var err error

	if missionID, parseErr := uuid.Parse(idParam); parseErr == nil {
		m, err = h.missionRepo.GetByID(ctx, missionID)
	} else {
		m, err = h.missionRepo.GetBySlug(ctx, idParam)
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	cacheKey := m.ID.String()
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.roi)
			return
		}
	}

	// Query 1: shopping items
	rows, err := h.pool.Query(ctx, `
		SELECT price, original_estimate, target_price, status
		FROM shopping_items
		WHERE mission_id = $1`, m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query shopping items")
		return
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		if scanErr := rows.Scan(&it.price, &it.originalEstimate, &it.targetPrice, &it.status); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan shopping item row")
			return
		}
		items = append(items, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate shopping item rows")
		return
	}

	// Query 2: price check count
	var priceChecks int
	err = h.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM price_history ph
		JOIN shopping_items si ON ph.item_id = si.id
		WHERE si.mission_id = $1`, m.ID).Scan(&priceChecks)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to count price checks")
		return
	}

	roi := compute(m.ID.String(), m.CreatedAt, items, priceChecks)

	h.cache.Store(cacheKey, cacheEntry{roi: roi, expiresAt: time.Now().Add(cacheTTL)})

	httputil.WriteJSON(w, http.StatusOK, roi)
}

// compute derives the ROI metrics from raw data. Pure function, testable without DB.
func compute(missionID string, createdAt time.Time, items []itemRow, priceChecks int) MissionROI {
	var totalSpend, totalEstimate, totalTargetSavings float64

	for _, it := range items {
		isBought := it.status == "buy" || it.status == "flash-sale"

		if isBought {
			totalSpend += it.price
		}

		if it.originalEstimate != nil {
			totalEstimate += *it.originalEstimate
		}

		// Savings toward target: negative value = savings (price <= target).
		if it.targetPrice != nil {
			diff := it.price - *it.targetPrice
			totalTargetSavings += diff
		}
	}

	savingsVsEstimate := totalEstimate - totalSpend

	var savingsPercent float64
	if totalEstimate > 0 {
		savingsPercent = savingsVsEstimate / totalEstimate * 100
	}

	daysTracking := int(math.Max(1, time.Since(createdAt).Hours()/24))
	avgChecksPerDay := float64(priceChecks) / float64(daysTracking)

	roiScore := math.Min(100, savingsPercent*2)

	return MissionROI{
		MissionID:          missionID,
		TotalSpend:         totalSpend,
		TotalEstimate:      totalEstimate,
		TotalTargetSavings: totalTargetSavings,
		SavingsVsEstimate:  savingsVsEstimate,
		SavingsPercent:     savingsPercent,
		PriceChecks:        priceChecks,
		DaysTracking:       daysTracking,
		AvgChecksPerDay:    avgChecksPerDay,
		ROIScore:           roiScore,
		Verdict:            verdictFor(roiScore),
		GeneratedAt:        time.Now().UTC(),
	}
}
