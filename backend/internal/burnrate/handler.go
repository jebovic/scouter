package burnrate

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves GET /api/missions/{missionId}/burn-rate.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// dayRow holds a single aggregated day from the purchases query.
type dayRow struct {
	day    string
	amount float64
}

// GetBurnRate handles GET /api/missions/{missionId}/burn-rate.
func (h *Handler) GetBurnRate(w http.ResponseWriter, r *http.Request) {
	missionParam := chi.URLParam(r, "missionId")
	if missionParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId required")
		return
	}

	ctx := r.Context()

	// Resolve mission by UUID or slug, fetch budget and creation date.
	var missionUUID string
	var budget float64
	var createdAt time.Time
	err := h.pool.QueryRow(ctx,
		`SELECT id::text, budget, created_at FROM missions WHERE id::text = $1 OR slug = $1`,
		missionParam,
	).Scan(&missionUUID, &budget, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query mission")
		return
	}

	// Fetch daily purchase totals.
	rows, err := h.pool.Query(ctx, `
		SELECT DATE(purchased_at)::text AS day, SUM(final_price) AS amount
		FROM purchase_records
		WHERE mission_id::text = $1 AND purchased_at IS NOT NULL
		GROUP BY DATE(purchased_at)
		ORDER BY day ASC`,
		missionUUID,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query purchases")
		return
	}
	defer rows.Close()

	var days []dayRow
	for rows.Next() {
		var dr dayRow
		if scanErr := rows.Scan(&dr.day, &dr.amount); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan purchase row")
			return
		}
		days = append(days, dr)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate purchase rows")
		return
	}

	resp := buildResponse(missionUUID, budget, createdAt, days)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildResponse computes the BurnRate Response from raw data. Pure function, testable without DB.
func buildResponse(missionID string, budget float64, createdAt time.Time, days []dayRow) Response {
	daysElapsed := int(math.Max(1, time.Since(createdAt).Hours()/24))

	// Build cumulative burn points and compute total spent.
	burnPoints := make([]DataPoint, 0, len(days))
	var cumulative float64
	for _, d := range days {
		cumulative += d.amount
		burnPoints = append(burnPoints, DataPoint{Date: d.day, Amount: math.Round(cumulative*100) / 100})
	}
	totalSpent := cumulative

	dailyRate := totalSpent / float64(daysElapsed)
	projectedTotal := dailyRate * 30

	var status string
	switch {
	case projectedTotal > budget*1.1:
		status = "over_budget"
	case projectedTotal > budget*0.9:
		status = "at_risk"
	default:
		status = "on_track"
	}

	return Response{
		MissionID:      missionID,
		TotalBudget:    budget,
		TotalSpent:     math.Round(totalSpent*100) / 100,
		DailyRate:      math.Round(dailyRate*100) / 100,
		ProjectedTotal: math.Round(projectedTotal*100) / 100,
		DaysElapsed:    daysElapsed,
		BurnPoints:     burnPoints,
		Status:         status,
	}
}
