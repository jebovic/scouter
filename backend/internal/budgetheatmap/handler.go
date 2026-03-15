package budgetheatmap

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 10 * time.Minute

// Handler exposes the budget heatmap endpoint.
type Handler struct {
	pool     *pgxpool.Pool
	mu       sync.Mutex
	cached   *BudgetHeatmap
	cachedAt time.Time
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetHeatmap handles GET /api/analytics/budget-heatmap.
// Results are cached for 10 minutes.
func (h *Handler) GetHeatmap(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) compute(r *http.Request) (*BudgetHeatmap, error) {
	ctx := r.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT si.cost_category,
		       TO_CHAR(ph.created_at, 'YYYY-MM') AS month,
		       SUM(ph.price)                      AS total,
		       COUNT(*)                           AS count
		FROM price_history ph
		JOIN shopping_items si ON si.id = ph.item_id
		WHERE ph.created_at > NOW() - INTERVAL '6 months'
		GROUP BY si.cost_category, month
		ORDER BY month, si.cost_category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []HeatmapCell
	maxTotal := 0.0

	catSet := map[string]struct{}{}
	monSet := map[string]struct{}{}

	for rows.Next() {
		var c HeatmapCell
		if err := rows.Scan(&c.Category, &c.Month, &c.Total, &c.Count); err != nil {
			return nil, err
		}
		cells = append(cells, c)
		catSet[c.Category] = struct{}{}
		monSet[c.Month] = struct{}{}
		if c.Total > maxTotal {
			maxTotal = c.Total
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Normalise intensity
	if maxTotal > 0 {
		for i := range cells {
			cells[i].Intensity = cells[i].Total / maxTotal
		}
	}

	// Build sorted slices
	categories := make([]string, 0, len(catSet))
	for k := range catSet {
		categories = append(categories, k)
	}
	sort.Strings(categories)

	months := make([]string, 0, len(monSet))
	for k := range monSet {
		months = append(months, k)
	}
	sort.Strings(months)

	if cells == nil {
		cells = []HeatmapCell{}
	}

	return &BudgetHeatmap{
		Cells:       cells,
		Categories:  categories,
		Months:      months,
		MaxTotal:    maxTotal,
		GeneratedAt: time.Now().UTC(),
	}, nil
}
