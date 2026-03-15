package listsorter

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 15 * time.Minute

type cacheEntry struct {
	resp      SortedResponse
	expiresAt time.Time
}

// Handler exposes the sorted-items endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: missionID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetSortedItems handles GET /api/missions/{missionId}/sorted-items.
func (h *Handler) GetSortedItems(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "missionId")
	missionID, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	// Serve from 15-min in-memory cache.
	if v, ok := h.cache.Load(rawID); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(rawID)
	}

	items, err := h.fetchItems(r, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	ranked := rankItems(items)
	resp := SortedResponse{
		Items:    ranked,
		SortedBy: "urgence",
	}

	h.cache.Store(rawID, &cacheEntry{
		resp:      resp,
		expiresAt: time.Now().Add(cacheTTL),
	})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// fetchItems loads all shopping items for missionID from the DB.
func (h *Handler) fetchItems(r *http.Request, missionID uuid.UUID) ([]rawItem, error) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, name, price, status, target_price, merchant, cost_category
		FROM shopping_items
		WHERE mission_id = $1`, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []rawItem
	for rows.Next() {
		var it rawItem
		if err := rows.Scan(
			&it.ID, &it.Name, &it.Price, &it.Status,
			&it.TargetPrice, &it.Merchant, &it.CostCategory,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// rankItems scores, sorts and assigns ranks to a slice of raw items.
func rankItems(items []rawItem) []RankedItem {
	type scored struct {
		item    rawItem
		score   int
		reasons []string
	}

	results := make([]scored, 0, len(items))

	for _, it := range items {
		s := scored{item: it}

		// Status-based score.
		switch it.Status {
		case "flash-sale":
			s.score += 100
			s.reasons = append(s.reasons, "Vente flash")
		case "crisis":
			s.score += 80
			s.reasons = append(s.reasons, "Urgence")
		case "buy":
			s.score += 50
			s.reasons = append(s.reasons, "À acheter")
		case "watch":
			s.score += 20
			s.reasons = append(s.reasons, "Surveiller")
		case "preorder":
			s.score += 10
			s.reasons = append(s.reasons, "Précommande")
		case "defer":
			s.score -= 10
			s.reasons = append(s.reasons, "Reporté")
		}

		// Target-price proximity bonus.
		if it.TargetPrice != nil && *it.TargetPrice > 0 {
			tp := *it.TargetPrice
			if it.Price <= tp {
				s.score += 60
				s.reasons = append(s.reasons, "objectif atteint")
			} else if it.Price <= tp*1.05 {
				s.score += 30
				s.reasons = append(s.reasons, "proche de l'objectif")
			}
		}

		// High-value item (needs attention).
		if it.Price > 500 {
			s.score += 15
			s.reasons = append(s.reasons, "article coûteux")
		}

		// No price set.
		if it.Price == 0 {
			s.score -= 20
			s.reasons = append(s.reasons, "prix non défini")
		}

		results = append(results, s)
	}

	// Sort descending by score; stable ties by name for determinism.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].item.Name < results[j].item.Name
	})

	ranked := make([]RankedItem, 0, len(results))
	for idx, s := range results {
		reason := strings.Join(s.reasons, " + ")
		if reason == "" {
			reason = "standard"
		}
		ranked = append(ranked, RankedItem{
			ID:     s.item.ID,
			Name:   s.item.Name,
			Price:  s.item.Price,
			Status: s.item.Status,
			Score:  s.score,
			Reason: reason,
			Rank:   idx + 1,
		})
	}
	return ranked
}
