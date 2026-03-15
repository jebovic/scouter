package purchase

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// CategoryStat represents aggregate spend stats for a mission category.
type CategoryStat struct {
	Category        string   `json:"category"`
	Count           int      `json:"count"`
	TotalSpent      float64  `json:"totalSpent"`
	AvgSatisfaction *float64 `json:"avgSatisfaction,omitempty"`
}

// StatsResponse is the response payload for GET /api/stats.
type StatsResponse struct {
	TotalSpent        float64        `json:"totalSpent"`
	TotalBudget       float64        `json:"totalBudget"`
	Savings           float64        `json:"savings"`
	PurchaseCount     int            `json:"purchaseCount"`
	CategoryBreakdown []CategoryStat `json:"categoryBreakdown"`
}

// StatsHandler serves GET /api/stats.
type StatsHandler struct {
	pool *pgxpool.Pool
}

// NewStatsHandler creates a stats handler backed by a pgx pool.
func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

// Stats handles GET /api/stats.
func (h *StatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var resp StatsResponse
	err := h.pool.QueryRow(ctx, `
		SELECT
		    COALESCE(SUM(pr.final_price), 0)::float8,
		    COALESCE(SUM(m.budget), 0)::float8,
		    COUNT(pr.id)::int
		FROM purchase_records pr
		JOIN missions m ON m.id = pr.mission_id
	`).Scan(&resp.TotalSpent, &resp.TotalBudget, &resp.PurchaseCount)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to compute stats")
		return
	}
	resp.Savings = resp.TotalBudget - resp.TotalSpent

	rows, err := h.pool.Query(ctx, `
		SELECT
		    COALESCE(m.category, 'other') AS category,
		    COUNT(pr.id)::int,
		    COALESCE(SUM(pr.final_price), 0)::float8,
		    AVG(pr.satisfaction::float8)
		FROM purchase_records pr
		JOIN missions m ON m.id = pr.mission_id
		GROUP BY m.category
		ORDER BY SUM(pr.final_price) DESC
	`)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to compute category breakdown")
		return
	}
	defer rows.Close()

	resp.CategoryBreakdown = []CategoryStat{}
	for rows.Next() {
		var cs CategoryStat
		if err := rows.Scan(&cs.Category, &cs.Count, &cs.TotalSpent, &cs.AvgSatisfaction); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan category row")
			return
		}
		resp.CategoryBreakdown = append(resp.CategoryBreakdown, cs)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read category breakdown")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// MonthlyStatPoint holds aggregated spend data for a single calendar month.
type MonthlyStatPoint struct {
	Month         string  `json:"month"`         // "YYYY-MM"
	TotalSpent    float64 `json:"totalSpent"`
	TotalBudget   float64 `json:"totalBudget"`
	PurchaseCount int     `json:"purchaseCount"`
}

// MonthlyStats handles GET /api/stats/monthly.
// It returns the last 12 months of aggregated spending data, ordered oldest-first.
func (h *StatsHandler) MonthlyStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT
		    TO_CHAR(DATE_TRUNC('month', pr.purchased_at), 'YYYY-MM') AS month,
		    COALESCE(SUM(pr.final_price), 0)::float8,
		    COALESCE(SUM(m.budget), 0)::float8,
		    COUNT(pr.id)::int
		FROM purchase_records pr
		JOIN missions m ON m.id = pr.mission_id
		WHERE pr.purchased_at >= NOW() - INTERVAL '12 months'
		GROUP BY DATE_TRUNC('month', pr.purchased_at)
		ORDER BY DATE_TRUNC('month', pr.purchased_at) ASC
	`)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query monthly stats")
		return
	}
	defer rows.Close()

	points := []MonthlyStatPoint{}
	for rows.Next() {
		var p MonthlyStatPoint
		if err := rows.Scan(&p.Month, &p.TotalSpent, &p.TotalBudget, &p.PurchaseCount); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan monthly stat row")
			return
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to read monthly stats")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, points)
}
