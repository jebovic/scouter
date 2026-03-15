package crossmission

import (
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 15 * time.Minute

// Handler serves the cross-mission comparison endpoint.
type Handler struct {
	pool     *pgxpool.Pool
	mu       sync.Mutex
	cached   *CrossMissionResponse
	cachedAt time.Time
}

// NewHandler creates a new cross-mission handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetComparison handles GET /api/analytics/cross-mission.
func (h *Handler) GetComparison(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.cached != nil && time.Since(h.cachedAt) < cacheTTL {
		cached := h.cached
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}
	h.mu.Unlock()

	result, err := h.compute(r)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.mu.Lock()
	h.cached = result
	h.cachedAt = time.Now()
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) compute(r *http.Request) (*CrossMissionResponse, error) {
	ctx := r.Context()

	// 1. Fetch up to 10 active missions ordered by most recent.
	missionRows, err := h.pool.Query(ctx,
		`SELECT id, name, budget, slug
		 FROM missions
		 WHERE archived_at IS NULL
		 ORDER BY created_at DESC
		 LIMIT 10`,
	)
	if err != nil {
		return nil, err
	}
	defer missionRows.Close()

	type missionMeta struct {
		id     string
		name   string
		budget float64
		slug   string
	}
	var missions []missionMeta
	for missionRows.Next() {
		var m missionMeta
		if err := missionRows.Scan(&m.id, &m.name, &m.budget, &m.slug); err != nil {
			return nil, err
		}
		missions = append(missions, m)
	}
	if err := missionRows.Err(); err != nil {
		return nil, err
	}

	// 2. For each mission, query aggregate shopping data.
	comparisons := make([]MissionComparison, 0, len(missions))
	for _, m := range missions {
		var itemCount, boughtCount int
		var totalPrice, totalEstimate float64

		err := h.pool.QueryRow(ctx,
			`SELECT
				COUNT(*) AS item_count,
				COALESCE(SUM(price), 0) AS total_price,
				COALESCE(SUM(original_estimate), 0) AS total_estimate,
				COUNT(CASE WHEN status = 'buy' OR status = 'flash-sale' THEN 1 END) AS bought_count
			 FROM shopping_items
			 WHERE mission_id = $1`,
			m.id,
		).Scan(&itemCount, &totalPrice, &totalEstimate, &boughtCount)
		if err != nil {
			return nil, err
		}

		cmp := buildComparison(m.id, m.name, m.slug, m.budget, itemCount, totalPrice, totalEstimate, boughtCount)
		comparisons = append(comparisons, cmp)
	}

	// 3. Find best mission (highest savingsPct among missions with estimates).
	var best *MissionComparison
	var totalSavings float64
	for i := range comparisons {
		c := &comparisons[i]
		if c.SavingsAmount > 0 {
			totalSavings += c.SavingsAmount
		}
		if c.Efficiency != "no data" {
			if best == nil || c.SavingsPct > best.SavingsPct {
				best = c
			}
		}
	}

	resp := &CrossMissionResponse{
		Missions:               comparisons,
		BestMission:            best,
		TotalSavingsAllMissions: totalSavings,
		GeneratedAt:            time.Now().UTC(),
	}
	return resp, nil
}

// buildComparison computes derived fields for a single mission.
func buildComparison(id, name, slug string, budget float64, itemCount int, totalPrice, totalEstimate float64, boughtCount int) MissionComparison {
	cmp := MissionComparison{
		MissionID:     id,
		Name:          name,
		Slug:          slug,
		Budget:        budget,
		ItemCount:     itemCount,
		TotalPrice:    totalPrice,
		TotalEstimate: totalEstimate,
		BoughtCount:   boughtCount,
	}

	if totalEstimate > 0 {
		cmp.SavingsAmount = totalEstimate - totalPrice
		cmp.SavingsPct = cmp.SavingsAmount / totalEstimate * 100
		switch {
		case cmp.SavingsPct > 10:
			cmp.Efficiency = "excellent"
		case cmp.SavingsPct > 0:
			cmp.Efficiency = "good"
		default:
			cmp.Efficiency = "overspent"
		}
	} else {
		cmp.Efficiency = "no data"
	}

	if budget > 0 {
		cmp.BudgetUsedPct = totalPrice / budget * 100
	}

	return cmp
}
