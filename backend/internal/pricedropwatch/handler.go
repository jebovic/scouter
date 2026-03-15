package pricedropwatch

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const (
	cacheTTL = 20 * time.Minute
)

// cacheEntry holds a cached watchlist with its expiration time
type cacheEntry struct {
	data      *WatchlistResponse
	expiresAt time.Time
}

// Handler handles price drop watchlist requests
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]*cacheEntry
}

// shoppingItem holds item data for analysis
type shoppingItem struct {
	ID           string
	Name         string
	CurrentPrice float64
}

// NewHandler creates a new handler instance
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
	}
}

// GetWatchlist returns the price drop watchlist for a mission
func (h *Handler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")

	// Check cache
	if cached, ok := h.cache.Load(missionID); ok {
		entry := cached.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		h.cache.Delete(missionID)
	}

	// Compute watchlist
	response, err := h.compute(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Cache result
	h.cache.Store(missionID, &cacheEntry{
		data:      response,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, response)
}

// compute retrieves and analyzes shopping items
func (h *Handler) compute(ctx context.Context, missionID string) (*WatchlistResponse, error) {
	// Query shopping items (excluding 'defer' status)
	query := `
		SELECT
			si.id,
			si.name,
			si.price as current_price,
			STDDEV(ph.price) as price_stddev,
			AVG(ph.price) as price_avg,
			MAX(ph.price) as price_max,
			COUNT(ph.id) as history_count
		FROM shopping_items si
		LEFT JOIN price_history ph ON ph.item_id = si.id
		WHERE si.mission_id = $1 AND si.status != 'defer'
		GROUP BY si.id, si.name, si.price
		ORDER BY si.name
	`

	rows, err := h.pool.Query(ctx, query, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WatchItem
	var itemsForAnalysis []shoppingItem

	for rows.Next() {
		var (
			id            string
			name          string
			currentPrice  float64
			stddev        *float64
			avg           *float64
			max           *float64
			historyCount  int64
		)

		if err := rows.Scan(&id, &name, &currentPrice, &stddev, &avg, &max, &historyCount); err != nil {
			return nil, err
		}

		item := shoppingItem{
			ID:           id,
			Name:         name,
			CurrentPrice: currentPrice,
		}
		itemsForAnalysis = append(itemsForAnalysis, item)

		// Analyze the item
		analyzed := h.analyzeItem(item, stddev, avg, max, historyCount)
		items = append(items, analyzed)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build French summary
	summary := h.buildSummary(items)

	return &WatchlistResponse{
		Summary:    summary,
		WatchItems: items,
	}, nil
}

// analyzeItem analyzes a single item for price drop probability
func (h *Handler) analyzeItem(item shoppingItem, stddev, avg, max *float64, historyCount int64) WatchItem {
	result := WatchItem{
		ItemID: item.ID,
		Name:   item.Name,
		CurrentPrice: item.CurrentPrice,
	}

	// No price history
	if historyCount == 0 {
		result.Signal = "no_history"
		result.DropProbability = h.deterministicProbability(item.ID, 40, 70)
		result.ExpectedDrop = 15
		result.Reasoning = "Pas d'historique de prix pour analyser les tendances."
		return result
	}

	if avg == nil || *avg == 0 {
		result.Signal = "stable"
		result.DropProbability = 20
		result.ExpectedDrop = 5
		result.Reasoning = "Pas assez de données pour analyser."
		return result
	}

	avgPrice := *avg

	// Calculate volatility as percentage of average
	var volatilityPercent float64
	if stddev != nil && *stddev > 0 {
		volatilityPercent = (*stddev / avgPrice) * 100
	}

	// Check for recent spike
	hasSpikeSignal := false
	if max != nil && item.CurrentPrice > 0 {
		spikePercent := ((*max - item.CurrentPrice) / item.CurrentPrice) * 100
		if spikePercent > 20 {
			hasSpikeSignal = true
		}
	}

	// Determine signal and probability
	if volatilityPercent > 15 {
		result.Signal = "high_volatility"
		result.DropProbability = h.deterministicProbability(item.ID, 55, 75)
		result.ExpectedDrop = 12
		result.Reasoning = fmt.Sprintf("Volatilité élevée détectée (%.1f%% d'écart type). Le prix fluctue beaucoup.", volatilityPercent)
	} else if hasSpikeSignal {
		result.Signal = "recent_spike"
		result.DropProbability = h.deterministicProbability(item.ID, 65, 80)
		result.ExpectedDrop = 18
		result.Reasoning = "Le prix a récemment atteint un sommet. Une correction est probable."
	} else {
		result.Signal = "stable"
		result.DropProbability = 20
		result.ExpectedDrop = 3
		result.Reasoning = "Prix stable avec peu de fluctuations récentes."
	}

	return result
}

// deterministicProbability generates a deterministic probability based on item ID hash
func (h *Handler) deterministicProbability(itemID string, minProb, maxProb float64) float64 {
	hasher := fnv.New32a()
	hasher.Write([]byte(itemID))
	hash := hasher.Sum32()

	// Normalize hash to [0, 1]
	normalized := float64(hash%100) / 100.0

	// Scale to [minProb, maxProb]
	return minProb + (normalized * (maxProb - minProb))
}

// buildSummary generates a French summary based on items
func (h *Handler) buildSummary(items []WatchItem) string {
	if len(items) == 0 {
		return "Aucun article à surveiller."
	}

	highRiskCount := 0
	mediumRiskCount := 0

	for _, item := range items {
		if item.DropProbability >= 70 {
			highRiskCount++
		} else if item.DropProbability >= 40 {
			mediumRiskCount++
		}
	}

	if highRiskCount > 0 {
		return fmt.Sprintf(
			"%d article(s) avec forte probabilité de baisse, %d avec risque modéré. Temps d'agir.",
			highRiskCount, mediumRiskCount,
		)
	} else if mediumRiskCount > 0 {
		return fmt.Sprintf(
			"%d article(s) avec risque modéré de baisse. À surveiller de près.",
			mediumRiskCount,
		)
	}

	return fmt.Sprintf("Surveillance active de %d article(s). Aucune baisse majeure détectée.", len(items))
}
