package decisionmatrix

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

const cacheTTL = 15 * time.Minute

type cacheEntry struct {
	resp      MatrixResponse
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/decision-matrix.
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
	id           string
	name         string
	price        float64
	targetPrice  *float64
	status       string
	costCategory string
	merchant     string
}

// GetMatrix handles GET /api/missions/{missionId}/decision-matrix.
// The {missionId} param may be a UUID or a slug.
func (h *Handler) GetMatrix(w http.ResponseWriter, r *http.Request) {
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
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
	}

	rows, err := h.pool.Query(ctx, `
		SELECT id, name, price, target_price, status, cost_category, merchant
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
		if scanErr := rows.Scan(&it.id, &it.name, &it.price, &it.targetPrice, &it.status, &it.costCategory, &it.merchant); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan shopping item row")
			return
		}
		items = append(items, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate shopping item rows")
		return
	}

	resp := compute(items)
	h.cache.Store(cacheKey, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// compute derives the decision matrix from raw item data. Pure function, testable without DB.
func compute(items []itemRow) MatrixResponse {
	if len(items) == 0 {
		return MatrixResponse{Items: []RankedItem{}, TopRecommendation: nil}
	}

	// Find min/max price for valeur (value) linear scaling.
	minPrice := items[0].price
	maxPrice := items[0].price
	for _, it := range items[1:] {
		if it.price < minPrice {
			minPrice = it.price
		}
		if it.price > maxPrice {
			maxPrice = it.price
		}
	}

	ranked := make([]RankedItem, 0, len(items))
	for _, it := range items {
		scores := ItemScores{
			Prix:          scorePrix(it.price, it.targetPrice),
			Urgence:       scoreUrgence(it.status),
			Valeur:        scoreValeur(it.price, minPrice, maxPrice),
			Disponibilite: scoreDisponibilite(it.status),
			Confiance:     scoreConfiance(it.targetPrice, it.merchant, it.costCategory),
		}
		total := scores.Prix + scores.Urgence + scores.Valeur + scores.Disponibilite + scores.Confiance
		ranked = append(ranked, RankedItem{
			ItemID:         it.id,
			Name:           it.name,
			Scores:         scores,
			TotalScore:     total,
			Recommendation: recommendation(total),
		})
	}

	// Sort descending by total score.
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].TotalScore > ranked[j].TotalScore
	})

	// Assign ranks (1-based).
	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	var top *RankedItem
	if len(ranked) > 0 {
		t := ranked[0]
		top = &t
	}

	return MatrixResponse{
		Items:             ranked,
		TopRecommendation: top,
	}
}

// scorePrix scores the price criterion (0–20).
func scorePrix(price float64, targetPrice *float64) int {
	if targetPrice == nil {
		return 10
	}
	tp := *targetPrice
	switch {
	case price <= tp:
		return 20
	case price <= tp*1.1:
		return 15
	case price <= tp*1.2:
		return 10
	default:
		return 5
	}
}

// scoreUrgence scores the urgency criterion from status (0–20).
func scoreUrgence(status string) int {
	switch status {
	case "flash-sale":
		return 20
	case "crisis":
		return 18
	case "buy":
		return 15
	case "preorder":
		return 12
	case "watch":
		return 8
	case "defer":
		return 3
	default:
		return 5
	}
}

// scoreValeur scores value via inverse price scaling within the mission (0–20).
func scoreValeur(price, minPrice, maxPrice float64) int {
	if maxPrice == minPrice {
		return 12 // single item or all same price — neutral
	}
	// Cheapest = 20, most expensive = 5, linear interpolation.
	ratio := (price - minPrice) / (maxPrice - minPrice) // 0.0 (cheapest) .. 1.0 (most expensive)
	score := 20.0 - ratio*15.0                          // 20 .. 5
	return int(score + 0.5)
}

// scoreDisponibilite scores availability from status (0–20).
func scoreDisponibilite(status string) int {
	switch status {
	case "flash-sale":
		return 20
	case "preorder":
		return 18
	case "buy":
		return 15
	case "watch":
		return 10
	case "defer":
		return 5
	default:
		return 8
	}
}

// scoreConfiance scores confidence based on data completeness (0–20).
func scoreConfiance(targetPrice *float64, merchant, costCategory string) int {
	score := 0
	if targetPrice != nil {
		score += 10
	}
	if merchant != "" {
		score += 5
	}
	if costCategory != "" {
		score += 5
	}
	return score
}
