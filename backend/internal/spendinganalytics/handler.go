package spendinganalytics

import (
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 10 * time.Minute

// Handler exposes the spending analytics endpoint.
type Handler struct {
	pool      *pgxpool.Pool
	mu        sync.Mutex
	cached    *SpendingAnalytics
	cachedAt  time.Time
}

// NewHandler creates a new Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetAnalytics handles GET /api/analytics/spending.
// Results are cached for 10 minutes to avoid repeated heavy aggregate queries.
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) compute(r *http.Request) (*SpendingAnalytics, error) {
	ctx := r.Context()

	analytics := &SpendingAnalytics{
		GeneratedAt:    time.Now().UTC(),
		ByCategory:     []CategoryStats{},
		ByMerchant:     []MerchantStats{},
		TopItems:       []TopItem{},
		MonthlySummary: []MonthStats{},
	}

	// 1. Category breakdown
	catRows, err := h.pool.Query(ctx, `
		SELECT cost_category, SUM(price), COUNT(*), AVG(price)
		FROM shopping_items
		WHERE status = 'buy'
		GROUP BY cost_category
		ORDER BY SUM(price) DESC`)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	totalSpending := 0.0
	for catRows.Next() {
		var cs CategoryStats
		if err := catRows.Scan(&cs.Category, &cs.Total, &cs.ItemCount, &cs.AvgPrice); err != nil {
			return nil, err
		}
		analytics.ByCategory = append(analytics.ByCategory, cs)
		totalSpending += cs.Total
	}
	if err := catRows.Err(); err != nil {
		return nil, err
	}

	// Compute percentage share per category
	analytics.TotalSpending = totalSpending
	if totalSpending > 0 {
		for i := range analytics.ByCategory {
			analytics.ByCategory[i].Pct = analytics.ByCategory[i].Total / totalSpending * 100
		}
	}

	// 2. Merchant breakdown (top 10)
	merchantRows, err := h.pool.Query(ctx, `
		SELECT merchant, SUM(price), COUNT(*)
		FROM shopping_items
		WHERE status = 'buy'
		GROUP BY merchant
		ORDER BY SUM(price) DESC
		LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer merchantRows.Close()

	for merchantRows.Next() {
		var ms MerchantStats
		if err := merchantRows.Scan(&ms.Merchant, &ms.Total, &ms.ItemCount); err != nil {
			return nil, err
		}
		analytics.ByMerchant = append(analytics.ByMerchant, ms)
	}
	if err := merchantRows.Err(); err != nil {
		return nil, err
	}

	// 3. Top items (most expensive buy items, up to 5)
	topRows, err := h.pool.Query(ctx, `
		SELECT si.id, si.name, m.name, si.price, si.cost_category
		FROM shopping_items si
		JOIN missions m ON m.id = si.mission_id
		WHERE si.status = 'buy'
		ORDER BY si.price DESC
		LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer topRows.Close()

	for topRows.Next() {
		var ti TopItem
		if err := topRows.Scan(&ti.ItemID, &ti.ItemName, &ti.MissionName, &ti.Price, &ti.Category); err != nil {
			return nil, err
		}
		analytics.TopItems = append(analytics.TopItems, ti)
	}
	if err := topRows.Err(); err != nil {
		return nil, err
	}

	// 4. Monthly summary (last 6 months of price_history)
	monthRows, err := h.pool.Query(ctx, `
		SELECT TO_CHAR(recorded_at, 'YYYY-MM'), AVG(price), COUNT(*)
		FROM price_history
		WHERE recorded_at > NOW() - INTERVAL '6 months'
		GROUP BY 1
		ORDER BY 1`)
	if err != nil {
		return nil, err
	}
	defer monthRows.Close()

	for monthRows.Next() {
		var ms MonthStats
		if err := monthRows.Scan(&ms.Month, &ms.AvgPrice, &ms.DataPoints); err != nil {
			return nil, err
		}
		analytics.MonthlySummary = append(analytics.MonthlySummary, ms)
	}
	if err := monthRows.Err(); err != nil {
		return nil, err
	}

	// 5. Total budget across all missions
	var totalBudget *float64
	err = h.pool.QueryRow(ctx, `SELECT SUM(budget) FROM missions`).Scan(&totalBudget)
	if err != nil {
		return nil, err
	}
	if totalBudget != nil {
		analytics.TotalBudget = *totalBudget
	}
	if analytics.TotalBudget > 0 {
		analytics.BudgetUsagePct = analytics.TotalSpending / analytics.TotalBudget * 100
	}

	return analytics, nil
}
