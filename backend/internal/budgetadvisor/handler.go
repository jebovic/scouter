package budgetadvisor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

type cacheEntry struct {
	advice    *BudgetAdvice
	timestamp time.Time
}

type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // string -> *cacheEntry
}

const cacheTTL = 10 * time.Minute

// NewHandler creates a new budget advisor handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
	}
}

// GetAdvice handles GET /api/missions/{missionId}/budget-advice
func (h *Handler) GetAdvice(w http.ResponseWriter, r *http.Request) {
	missionParam := chi.URLParam(r, "missionId")
	if missionParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId required")
		return
	}

	// Check cache first
	if cached, ok := h.cache.Load(missionParam); ok {
		entry := cached.(*cacheEntry)
		if time.Since(entry.timestamp) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.advice)
			return
		}
	}

	advice, err := h.computeAdvice(r.Context(), missionParam)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Store in cache
	h.cache.Store(missionParam, &cacheEntry{
		advice:    advice,
		timestamp: time.Now(),
	})

	httputil.WriteJSON(w, http.StatusOK, advice)
}

func (h *Handler) computeAdvice(ctx context.Context, missionParam string) (*BudgetAdvice, error) {
	// 1. Query mission budget, resolve by UUID or slug
	var missionUUID string
	var totalBudget float64
	err := h.pool.QueryRow(ctx, "SELECT id::text, budget FROM missions WHERE id::text = $1 OR slug = $1", missionParam).Scan(&missionUUID, &totalBudget)
	if err != nil {
		return nil, err
	}

	// 2. Query items excluding 'defer' status
	rows, err := h.pool.Query(ctx,
		`SELECT id::text, name, price, status, target_price
		 FROM shopping_items
		 WHERE mission_id::text = $1 AND status != 'defer'
		 ORDER BY created_at`,
		missionUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type item struct {
		id          string
		name        string
		price       float64
		status      string
		targetPrice *float64
	}

	var items []item
	for rows.Next() {
		var it item
		err := rows.Scan(&it.id, &it.name, &it.price, &it.status, &it.targetPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	// 3. Compute spent (sum of 'buy' status items)
	var spent float64
	for _, it := range items {
		if it.status == "buy" {
			spent += it.price
		}
	}

	remaining := totalBudget - spent

	// 4. For each non-'buy' item, determine allocation
	type allocItem struct {
		itemID          string
		name            string
		price           float64
		allocatedBudget float64
		status          string
		priority        int
	}

	var allocations []allocItem
	for _, it := range items {
		if it.status == "buy" {
			continue
		}

		allocBudget := it.price
		if it.targetPrice != nil {
			allocBudget = *it.targetPrice
		}

		// Determine priority based on status
		priority := 5
		switch it.status {
		case "flash-sale":
			priority = 1
		case "watch":
			priority = 2
		case "preorder":
			priority = 3
		case "crisis":
			priority = 4
		}

		allocations = append(allocations, allocItem{
			itemID:          it.id,
			name:            it.name,
			price:           it.price,
			allocatedBudget: allocBudget,
			status:          it.status,
			priority:        priority,
		})
	}

	// 5. Sort by priority then price desc
	sort.Slice(allocations, func(i, j int) bool {
		if allocations[i].priority != allocations[j].priority {
			return allocations[i].priority < allocations[j].priority
		}
		return allocations[i].allocatedBudget > allocations[j].allocatedBudget
	})

	// 6. Assign priority 1..N and compute feasibility
	var allocationItems []AllocationItem
	runningTotal := 0.0
	for idx, alloc := range allocations {
		feasible := (runningTotal + alloc.allocatedBudget) <= remaining

		reason := h.buildReason(alloc.status, alloc.allocatedBudget, feasible)

		allocationItems = append(allocationItems, AllocationItem{
			ItemID:          alloc.itemID,
			Name:            alloc.name,
			Price:           alloc.price,
			AllocatedBudget: alloc.allocatedBudget,
			Priority:        idx + 1,
			Reason:          reason,
			Feasible:        feasible,
		})

		if feasible {
			runningTotal += alloc.allocatedBudget
		}
	}

	totalAllocated := runningTotal
	surplus := remaining - totalAllocated

	// 7. Build French advice
	advice := h.buildAdvice(surplus, totalAllocated, remaining)

	return &BudgetAdvice{
		MissionID:      missionUUID,
		TotalBudget:    totalBudget,
		Spent:          spent,
		Remaining:      remaining,
		Allocations:    allocationItems,
		TotalAllocated: totalAllocated,
		Surplus:        surplus,
		Advice:         advice,
	}, nil
}

func (h *Handler) buildReason(status string, allocatedBudget float64, feasible bool) string {
	var base string
	switch status {
	case "flash-sale":
		base = "🔥 Excellent affaire - Ne pas manquer"
	case "watch":
		base = "👀 À surveiller de près"
	case "preorder":
		base = "📅 Réservation recommandée"
	case "crisis":
		base = "🚨 Priorité urgente"
	default:
		base = "Allocation standard"
	}

	if !feasible {
		return fmt.Sprintf("%s (budget insuffisant pour cet article)", base)
	}
	return base
}

func (h *Handler) buildAdvice(surplus, totalAllocated, remaining float64) string {
	if surplus > 0 {
		return fmt.Sprintf(
			"💰 Budget bien géré! Il vous reste %.2f€ après allocation de %.2f€.",
			surplus,
			totalAllocated,
		)
	}
	if surplus < 0 {
		return fmt.Sprintf(
			"⚠️ Dépassement prévu de %.2f€. Réduisez les allocations ou augmentez le budget.",
			-surplus,
		)
	}
	return fmt.Sprintf(
		"✅ Budget équilibré! Vous avez alloué exactement les %.2f€ disponibles.",
		totalAllocated,
	)
}
