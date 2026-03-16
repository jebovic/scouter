package regretanalyzer

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// resolveMissionID resolves a slug-or-UUID string to a UUID by querying the missions table.
func resolveMissionID(ctx context.Context, pool *pgxpool.Pool, param string) (uuid.UUID, error) {
	var idStr string
	err := pool.QueryRow(ctx,
		`SELECT id::text FROM missions WHERE id::text = $1 OR slug = $1`,
		param,
	).Scan(&idStr)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(idStr)
}

// Handler analyzes purchase regret for a mission based on pricing signals.
type Handler struct {
	pool  *pgxpool.Pool
	cache map[string]*cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	resp      *Response
	timestamp time.Time
}

// NewHandler creates a new regret analyzer handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]*cacheEntry),
	}
}

// GetAnalysis handles GET /api/missions/{missionId}/regret-analysis
func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	missionParam := chi.URLParam(r, "missionId")
	missionID, err := resolveMissionID(r.Context(), h.pool, missionParam)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Check cache
	h.mu.RLock()
	if entry, ok := h.cache[missionID.String()]; ok && time.Since(entry.timestamp) < 1*time.Hour {
		h.mu.RUnlock()
		httputil.WriteJSON(w, http.StatusOK, entry.resp)
		return
	}
	h.mu.RUnlock()

	// Fetch shopping items for mission
	items, err := h.fetchShoppingItems(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// If no items, return empty analysis
	if len(items) == 0 {
		resp := &Response{
			MissionID:   missionID,
			TotalSpent:  0,
			Signals:     []RegretSignal{},
			RegretScore: 0,
			Summary:     "Aucun achat pour cette mission.",
		}
		h.cacheResponse(missionID.String(), resp)
		httputil.WriteJSON(w, http.StatusOK, resp)
		return
	}

	// Fetch price history for each item
	itemPrices := make(map[uuid.UUID]*ItemPrices)
	totalSpent := 0.0
	for _, item := range items {
		totalSpent += item.Price
		hist, err := h.fetchPriceHistory(r.Context(), item.ID)
		if err != nil {
			// Silently skip if price history fails for this item
			itemPrices[item.ID] = &ItemPrices{
				MinPrice:     item.Price,
				AvgPrice:     item.Price,
				Count:        1,
				Name:         item.Name,
				ItemID:       item.ID,
				CurrentPrice: item.Price,
				Status:       item.Status,
				TargetPrice:  item.TargetPrice,
			}
			continue
		}
		if hist == nil {
			hist = &ItemPrices{
				MinPrice:     item.Price,
				AvgPrice:     item.Price,
				Count:        1,
				Name:         item.Name,
				ItemID:       item.ID,
				CurrentPrice: item.Price,
				Status:       item.Status,
				TargetPrice:  item.TargetPrice,
			}
		}
		hist.CurrentPrice = item.Price
		hist.Status = item.Status
		hist.TargetPrice = item.TargetPrice
		itemPrices[item.ID] = hist
	}

	// Generate signals
	signals, score := h.generateSignals(items, itemPrices)

	// Generate summary
	summary := h.generateSummary(score)

	resp := &Response{
		MissionID:   missionID,
		TotalSpent:  totalSpent,
		Signals:     signals,
		RegretScore: score,
		Summary:     summary,
	}

	h.cacheResponse(missionID.String(), resp)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

type ShoppingItem struct {
	ID               uuid.UUID
	Name             string
	Price            float64
	Status           string
	TargetPrice      *float64
	OriginalEstimate *float64
}

type ItemPrices struct {
	ItemID       uuid.UUID
	Name         string
	MinPrice     float64
	AvgPrice     float64
	Count        int
	CurrentPrice float64
	Status       string
	TargetPrice  *float64
}

func (h *Handler) fetchShoppingItems(ctx context.Context, missionID uuid.UUID) ([]ShoppingItem, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, price, status, target_price, original_estimate
		FROM shopping_items
		WHERE mission_id::text = $1
		ORDER BY created_at ASC
	`, missionID.String())
	if err != nil {
		return nil, fmt.Errorf("fetch shopping items: %w", err)
	}
	defer rows.Close()

	var items []ShoppingItem
	for rows.Next() {
		var item ShoppingItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Status, &item.TargetPrice, &item.OriginalEstimate); err != nil {
			return nil, fmt.Errorf("scan shopping item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (h *Handler) fetchPriceHistory(ctx context.Context, itemID uuid.UUID) (*ItemPrices, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT price FROM shopping_price_snapshots
		WHERE item_id = $1
		ORDER BY recorded_at ASC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("fetch price history: %w", err)
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var price float64
		if err := rows.Scan(&price); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, price)
	}

	if len(prices) == 0 {
		return nil, nil
	}

	minPrice := prices[0]
	sumPrice := 0.0
	for _, p := range prices {
		if p < minPrice {
			minPrice = p
		}
		sumPrice += p
	}

	return &ItemPrices{
		MinPrice: minPrice,
		AvgPrice: sumPrice / float64(len(prices)),
		Count:    len(prices),
	}, nil
}

func (h *Handler) generateSignals(items []ShoppingItem, prices map[uuid.UUID]*ItemPrices) ([]RegretSignal, int) {
	var signals []RegretSignal
	regretScore := 0

	for _, item := range items {
		itemPrices, ok := prices[item.ID]
		if !ok {
			continue
		}

		// Check for overpaid signal
		if itemPrices.Count > 1 && itemPrices.AvgPrice > 0 {
			pctAboveAvg := (item.Price - itemPrices.AvgPrice) / itemPrices.AvgPrice
			if pctAboveAvg > 0.1 {
				// Paid 10%+ above average
				severity := "medium"
				if pctAboveAvg > 0.2 {
					severity = "high"
					regretScore += 30
				} else {
					regretScore += 15
				}

				potentialSaving := item.Price - itemPrices.AvgPrice
				signals = append(signals, RegretSignal{
					Type:            "overpaid",
					Severity:        severity,
					Description:     fmt.Sprintf("Payé %.0f%% au-dessus de la moyenne observée (%.2f€ vs %.2f€)", pctAboveAvg*100, item.Price, itemPrices.AvgPrice),
					PotentialSaving: potentialSaving,
				})
			}
		}

		// Check for timing signal
		if itemPrices.Count > 1 && itemPrices.MinPrice > 0 {
			pctAboveMin := (item.Price - itemPrices.MinPrice) / itemPrices.MinPrice
			if pctAboveMin > 0.15 {
				// Paid 15%+ above minimum observed
				potentialSaving := item.Price - itemPrices.MinPrice
				signals = append(signals, RegretSignal{
					Type:            "timing",
					Severity:        "low",
					Description:     fmt.Sprintf("Acheté quand le prix était %.0f%% au-dessus du minimum (%.2f€ vs %.2f€)", pctAboveMin*100, item.Price, itemPrices.MinPrice),
					PotentialSaving: potentialSaving,
				})
				regretScore += 5
			}
		}

		// Check for impulse signal
		if item.Status == "flash-sale" && item.TargetPrice != nil && item.Price > *item.TargetPrice {
			potentialSaving := item.Price - *item.TargetPrice
			signals = append(signals, RegretSignal{
				Type:            "impulse",
				Severity:        "medium",
				Description:     fmt.Sprintf("Achat impulsif en flash-sale au-dessus du prix cible (%.2f€ vs cible %.2f€)", item.Price, *item.TargetPrice),
				PotentialSaving: potentialSaving,
			})
			regretScore += 15
		}
	}

	// Cap regret score at 100
	regretScore = int(math.Min(100, float64(regretScore)))

	return signals, regretScore
}

func (h *Handler) generateSummary(score int) string {
	switch {
	case score <= 20:
		return "Excellentes décisions d'achat. Vous avez bien géré votre budget."
	case score <= 50:
		return "Quelques opportunités manquées. Considérez une meilleure planification pour les achats futurs."
	case score <= 80:
		return "Plusieurs achats auraient pu être optimisés. Revoyez votre timing et vos prix cibles."
	default:
		return "Révision stratégique recommandée. Analysez les tendances de prix avant d'acheter."
	}
}

func (h *Handler) cacheResponse(key string, resp *Response) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache[key] = &cacheEntry{
		resp:      resp,
		timestamp: time.Now(),
	}
}
