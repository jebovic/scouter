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
