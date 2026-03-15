package wishlistprioritizer

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 15 * time.Minute

// PrioritizedItem is a single wishlist item with priority scores.
type PrioritizedItem struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
	Verdict    string  `json:"verdict"`
	Urgency    float64 `json:"urgency"`
	Trend      float64 `json:"trend"`
	BudgetFit  float64 `json:"budgetFit"`
}

// TopPick summarizes the highest-priority item.
type TopPick struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PrioritizedList is the full response.
type PrioritizedList struct {
	Items   []PrioritizedItem `json:"items"`
	TopPick *TopPick          `json:"topPick"`
	Summary string            `json:"summary"`
}

type cacheEntry struct {
	data      *PrioritizedList
	expiresAt time.Time
}

// Handler serves the smart wishlist prioritizer endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a new Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

// GetPrioritized handles GET /api/wishlist/prioritized.
func (h *Handler) GetPrioritized(w http.ResponseWriter, r *http.Request) {
	const cacheKey = "global"

	h.mu.Lock()
	if entry, ok := h.cache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		result := entry.data
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, result)
		return
	}
	h.mu.Unlock()

	result, err := h.load(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.mu.Lock()
	h.cache[cacheKey] = cacheEntry{data: result, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, result)
}

type rawItem struct {
	id               string
	name             string
	lastCheckedPrice *float64
	targetPrice      *float64
	alertEnabled     bool
	updatedAt        time.Time
}

func (h *Handler) load(ctx context.Context) (*PrioritizedList, error) {
	rows, err := h.pool.Query(ctx,
		`SELECT id::text, name, last_checked_price, target_price, alert_enabled, updated_at
		 FROM wish_list_items
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []rawItem
	for rows.Next() {
		var it rawItem
		if err := rows.Scan(&it.id, &it.name, &it.lastCheckedPrice, &it.targetPrice, &it.alertEnabled, &it.updatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildPrioritizedList(items), nil
}

func buildPrioritizedList(items []rawItem) *PrioritizedList {
	if len(items) == 0 {
		return &PrioritizedList{
			Items:   []PrioritizedItem{},
			TopPick: nil,
			Summary: "Votre liste de souhaits est vide.",
		}
	}

	// Compute average last_checked_price for budgetFit baseline
	var sum float64
	var count int
	for _, it := range items {
		if it.lastCheckedPrice != nil {
			sum += *it.lastCheckedPrice
			count++
		}
	}
	avgPrice := 0.0
	if count > 0 {
		avgPrice = sum / float64(count)
	}

	prioritized := make([]PrioritizedItem, 0, len(items))
	for _, it := range items {
		urgency := computeUrgency(it)
		trend := computeTrend(it.id)
		budgetFit := computeBudgetFit(it.lastCheckedPrice, avgPrice)

		score := urgency*0.4 + trend*0.35 + budgetFit*0.25

		prioritized = append(prioritized, PrioritizedItem{
			ID:        it.id,
			Name:      it.name,
			Score:     score,
			Urgency:   urgency,
			Trend:     trend,
			BudgetFit: budgetFit,
			Verdict:   verdictFor(score),
		})
	}

	// Sort by score descending
	sort.Slice(prioritized, func(i, j int) bool {
		return prioritized[i].Score > prioritized[j].Score
	})

	// Assign ranks
	for i := range prioritized {
		prioritized[i].Rank = i + 1
	}

	var topPick *TopPick
	if len(prioritized) > 0 {
		topPick = &TopPick{ID: prioritized[0].ID, Name: prioritized[0].Name}
	}

	summary := fmt.Sprintf("%d article(s) priorisé(s). Meilleure opportunité : %s (score %.0f/100).",
		len(prioritized), prioritized[0].Name, prioritized[0].Score)

	return &PrioritizedList{
		Items:   prioritized,
		TopPick: topPick,
		Summary: summary,
	}
}

func computeUrgency(it rawItem) float64 {
	score := 0.0

	// alert enabled: base urgency
	if it.alertEnabled {
		score += 50
	}

	// price near target (within 10%) → high urgency
	if it.lastCheckedPrice != nil && it.targetPrice != nil && *it.targetPrice > 0 {
		diff := (*it.lastCheckedPrice - *it.targetPrice) / *it.targetPrice
		if diff >= -0.1 && diff <= 0.1 {
			score += 40
		} else if diff < -0.1 {
			// already below target — urgent to buy
			score += 50
		}
	}

	// decay for items not updated in 30 days
	if time.Since(it.updatedAt) > 30*24*time.Hour {
		score -= 20
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// computeTrend uses a deterministic FNV-32a hash of the item ID to simulate
// a price trend direction (0-100).
func computeTrend(id string) float64 {
	h := fnv.New32a()
	h.Write([]byte(id))
	return float64(h.Sum32()%101)
}

// computeBudgetFit scores items with lower price relative to average higher.
func computeBudgetFit(price *float64, avg float64) float64 {
	if price == nil || avg <= 0 {
		return 50
	}
	// Lower price relative to average → higher fit
	ratio := *price / avg
	score := (2 - ratio) * 50
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func verdictFor(score float64) string {
	switch {
	case score >= 75:
		return "Priorité haute — acheter maintenant"
	case score >= 50:
		return "Priorité moyenne — surveiller"
	default:
		return "Priorité basse — peut attendre"
	}
}
