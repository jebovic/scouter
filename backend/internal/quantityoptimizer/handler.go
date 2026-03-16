package quantityoptimizer

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

const cacheTTL = 10 * time.Minute

var defaultQuantities = []int{1, 2, 3, 5, 10}

// QuantityTier represents a single quantity option with pricing.
type QuantityTier struct {
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unitPrice"`
	TotalPrice    float64 `json:"totalPrice"`
	Discount      float64 `json:"discountPct"`
	BudgetFit     bool    `json:"budgetFit"`
	Recommendation string `json:"recommendation"`
}

// OptimizerReport is the full response.
type OptimizerReport struct {
	ItemID       string         `json:"itemId"`
	ItemName     string         `json:"itemName"`
	BasePrice    float64        `json:"basePrice"`
	Budget       *float64       `json:"budget"`
	OptimalQty   int            `json:"optimalQty"`
	Tiers        []QuantityTier `json:"tiers"`
	Summary      string         `json:"summary"`
}

type cacheEntry struct {
	data      *OptimizerReport
	expiresAt time.Time
}

// Handler serves the quantity optimizer endpoint.
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

// GetReport handles GET /api/missions/{missionId}/items/{itemId}/quantity-optimizer.
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	itemID := chi.URLParam(r, "itemId")
	if missionID == "" || itemID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId and itemId required")
		return
	}

	cacheKey := missionID + ":" + itemID

	h.mu.Lock()
	if entry, ok := h.cache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		result := entry.data
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, result)
		return
	}
	h.mu.Unlock()

	result, err := h.load(r.Context(), missionID, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.mu.Lock()
	h.cache[cacheKey] = cacheEntry{data: result, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, result)
}

type itemRow struct {
	id           string
	name         string
	currentPrice *float64
}

func (h *Handler) load(ctx context.Context, missionID, itemID string) (*OptimizerReport, error) {
	var item itemRow
	err := h.pool.QueryRow(ctx,
		`SELECT id::text, name, price FROM shopping_items
		 WHERE id::text = $1
		   AND mission_id = (SELECT id FROM missions WHERE id::text = $2 OR slug = $2)`,
		itemID, missionID,
	).Scan(&item.id, &item.name, &item.currentPrice)
	if err != nil {
		return nil, err
	}

	// Fetch mission budget constraint (optional).
	var budget *float64
	_ = h.pool.QueryRow(ctx,
		`SELECT (constraints->>'budget')::float FROM missions WHERE (id::text = $1 OR slug = $1) AND constraints->>'budget' IS NOT NULL`,
		missionID,
	).Scan(&budget)

	basePrice := 0.0
	if item.currentPrice != nil {
		basePrice = *item.currentPrice
	}

	return buildReport(item.id, item.name, basePrice, budget), nil
}

// buildReport is a pure function for testability.
func buildReport(itemID, itemName string, basePrice float64, budget *float64) *OptimizerReport {
	if basePrice <= 0 {
		basePrice = computeFallbackPrice(itemID)
	}

	tiers := make([]QuantityTier, 0, len(defaultQuantities))
	bestQty := 1
	bestScore := -1.0

	for _, qty := range defaultQuantities {
		discount := computeDiscount(itemID, qty)
		unitPrice := basePrice * (1 - discount/100)
		totalPrice := unitPrice * float64(qty)
		budgetFit := budget == nil || totalPrice <= *budget

		rec := recommendationFor(qty, discount, budgetFit)

		// Score: maximize discount while fitting budget.
		score := discount
		if !budgetFit {
			score = -1
		}
		if score > bestScore {
			bestScore = score
			bestQty = qty
		}

		tiers = append(tiers, QuantityTier{
			Quantity:      qty,
			UnitPrice:     roundTo2(unitPrice),
			TotalPrice:    roundTo2(totalPrice),
			Discount:      roundTo2(discount),
			BudgetFit:     budgetFit,
			Recommendation: rec,
		})
	}

	summary := fmt.Sprintf("Quantité optimale : %d unité(s) — meilleur rapport qualité/prix pour \"%s\".", bestQty, itemName)

	return &OptimizerReport{
		ItemID:     itemID,
		ItemName:   itemName,
		BasePrice:  roundTo2(basePrice),
		Budget:     budget,
		OptimalQty: bestQty,
		Tiers:      tiers,
		Summary:    summary,
	}
}

// computeDiscount uses FNV-32a to produce a deterministic discount curve.
// Quantity 1 → base discount; higher quantities get progressively more (up to 25%).
func computeDiscount(itemID string, qty int) float64 {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%s:%d", itemID, qty)))
	base := float64(h.Sum32()%11) // 0-10% base per tier
	// Add qty bonus: each additional unit adds ~2% discount.
	bonus := float64(qty-1) * 2.0
	total := base + bonus
	if total > 25 {
		total = 25
	}
	return total
}

// computeFallbackPrice produces a deterministic price when current_price is NULL.
func computeFallbackPrice(itemID string) float64 {
	h := fnv.New32a()
	h.Write([]byte(itemID + "price"))
	return float64(h.Sum32()%951) + 50.0 // 50-1000
}

func recommendationFor(qty int, discount float64, budgetFit bool) string {
	if !budgetFit {
		return "Hors budget"
	}
	switch {
	case qty == 1:
		return "Achat standard"
	case discount >= 15:
		return "Excellent rapport qualité/prix"
	case discount >= 8:
		return "Bonne économie"
	default:
		return "Économie légère"
	}
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
