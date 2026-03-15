package targetsuggestion

import (
	"encoding/json"
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

// Strategy represents a single target-price recommendation strategy.
type Strategy struct {
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Label     string  `json:"label"`
	Rationale string  `json:"rationale"`
}

// SuggestionResponse is the JSON envelope returned by GetSuggestion.
type SuggestionResponse struct {
	ItemID     string     `json:"itemId"`
	Strategies []Strategy `json:"strategies"`
}

type cacheEntry struct {
	resp      SuggestionResponse
	expiresAt time.Time
}

// Handler exposes target-suggestion endpoints.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetSuggestion handles GET /api/missions/{missionId}/items/{itemId}/target-suggestion.
func (h *Handler) GetSuggestion(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 30-min in-memory cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(rawItemID)
	}

	prices, err := h.fetchPrices(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if len(prices) < 3 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "insufficient data")
		return
	}

	resp := computeSuggestions(rawItemID, prices)

	h.cache.Store(rawItemID, &cacheEntry{
		resp:      resp,
		expiresAt: time.Now().Add(30 * time.Minute),
	})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// ApplySuggestion handles POST /api/missions/{missionId}/items/{itemId}/target-suggestion/apply.
// Body: {"target": "conservative"|"moderate"|"aggressive"}
func (h *Handler) ApplySuggestion(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Target != "conservative" && req.Target != "moderate" && req.Target != "aggressive" {
		httputil.WriteError(w, http.StatusBadRequest, "target must be one of: conservative, moderate, aggressive")
		return
	}

	// Resolve (or recompute) the suggestion to get the price for the chosen tier.
	var resp SuggestionResponse
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			resp = entry.resp
		}
	}
	if resp.ItemID == "" {
		prices, err := h.fetchPrices(r, itemID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if len(prices) < 3 {
			httputil.WriteError(w, http.StatusUnprocessableEntity, "insufficient data")
			return
		}
		resp = computeSuggestions(rawItemID, prices)
		h.cache.Store(rawItemID, &cacheEntry{
			resp:      resp,
			expiresAt: time.Now().Add(30 * time.Minute),
		})
	}

	var targetPrice float64
	for _, s := range resp.Strategies {
		if s.Type == req.Target {
			targetPrice = s.Price
			break
		}
	}

	// Update target_price in DB and return the updated item.
	row := h.pool.QueryRow(r.Context(), `
		UPDATE shopping_items
		SET target_price = $2
		WHERE id = $1
		RETURNING id, mission_id, name, merchant, cost_category, price,
		          original_estimate, target_price, status, note, url, pinned, created_at`,
		itemID, targetPrice)

	type Item struct {
		ID               uuid.UUID `json:"id"`
		MissionID        uuid.UUID `json:"missionId"`
		Name             string    `json:"name"`
		Merchant         string    `json:"merchant"`
		CostCategory     string    `json:"costCategory"`
		Price            float64   `json:"price"`
		OriginalEstimate *float64  `json:"originalEstimate,omitempty"`
		TargetPrice      *float64  `json:"targetPrice,omitempty"`
		Status           string    `json:"status"`
		Note             string    `json:"note,omitempty"`
		URL              string    `json:"url,omitempty"`
		Pinned           bool      `json:"pinned"`
		CreatedAt        time.Time `json:"createdAt"`
	}

	var item Item
	if err := row.Scan(
		&item.ID, &item.MissionID, &item.Name, &item.Merchant, &item.CostCategory,
		&item.Price, &item.OriginalEstimate, &item.TargetPrice, &item.Status,
		&item.Note, &item.URL, &item.Pinned, &item.CreatedAt,
	); err != nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	// Invalidate cache so next GET recalculates with the new target context.
	h.cache.Delete(rawItemID)

	httputil.WriteJSON(w, http.StatusOK, item)
}

// fetchPrices loads all recorded prices for itemID from price_history.
func (h *Handler) fetchPrices(r *http.Request, itemID uuid.UUID) ([]float64, error) {
	rows, err := h.pool.Query(r.Context(),
		`SELECT price FROM price_history WHERE item_id = $1 ORDER BY recorded_at`,
		itemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// computeSuggestions derives the three strategy prices from the price history.
// Requires len(prices) >= 3.
func computeSuggestions(itemID string, prices []float64) SuggestionResponse {
	minP := prices[0]
	sum := 0.0
	for _, p := range prices {
		sum += p
		if p < minP {
			minP = p
		}
	}
	avg := sum / float64(len(prices))

	conservative := round2(minP)
	moderate := round2(minP + (avg-minP)*0.3)
	aggressive := round2(avg * 0.85)

	return SuggestionResponse{
		ItemID: itemID,
		Strategies: []Strategy{
			{
				Type:  "conservative",
				Price: conservative,
				Label: "Prix plancher historique",
				Rationale: fmt.Sprintf(
					"Correspond au prix le plus bas observé (%.2f€). Stratégie prudente : attendez que le prix descende à son niveau minimal historique.",
					conservative,
				),
			},
			{
				Type:  "moderate",
				Price: moderate,
				Label: "Opportunité modérée",
				Rationale: fmt.Sprintf(
					"30%% au-dessus du plancher, en dessous de la moyenne (%.2f€). Bon équilibre entre attente et opportunité.",
					avg,
				),
			},
			{
				Type:  "aggressive",
				Price: aggressive,
				Label: "Réduction agressive",
				Rationale: fmt.Sprintf(
					"15%% en dessous de la moyenne historique (%.2f€). Idéal si vous souhaitez capturer une vraie promotion.",
					avg,
				),
			},
		},
	}
}

// round2 rounds a float to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
