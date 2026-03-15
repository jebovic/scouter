package shoppingoptimizer

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// shoppingLister loads all items for a mission.
type shoppingLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler handles POST /api/missions/{missionID}/shopping/optimize.
type Handler struct {
	repo shoppingLister
}

// NewHandler creates a Handler backed by the given pgx pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{repo: shopping.NewRepository(pool)}
}

// Optimize scores and ranks shopping items by deal quality and urgency.
func (h *Handler) Optimize(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	items, err := h.repo.ListByMission(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if len(items) == 0 {
		httputil.WriteJSON(w, http.StatusOK, OptimizeResponse{
			Items:   []OptimizedItem{},
			Summary: "Aucun article dans la liste.",
		})
		return
	}

	resp := buildResponse(items)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildResponse scores, sorts, and summarises shopping items.
func buildResponse(items []shopping.Item) OptimizeResponse {
	optimized := make([]OptimizedItem, 0, len(items))
	var totalBudget float64

	for _, item := range items {
		optimized = append(optimized, scoreItem(item))
		totalBudget += item.Price
	}

	// Sort by PriorityScore descending.
	sort.Slice(optimized, func(i, j int) bool {
		return optimized[i].PriorityScore > optimized[j].PriorityScore
	})

	// Compute total optimized = sum of items with urgency "now".
	var totalOptimized float64
	nowCount, soonCount, waitCount := 0, 0, 0
	for _, oi := range optimized {
		if oi.UrgencyLevel == "now" {
			totalOptimized += oi.Price
			nowCount++
		} else if oi.UrgencyLevel == "soon" {
			soonCount++
		} else {
			waitCount++
		}
	}

	summary := buildSummary(nowCount, soonCount, waitCount, totalOptimized)

	return OptimizeResponse{
		Items:          optimized,
		TotalBudget:    totalBudget,
		TotalOptimized: totalOptimized,
		Summary:        summary,
	}
}

// scoreItem computes a priority score and urgency level for one shopping item.
func scoreItem(item shopping.Item) OptimizedItem {
	score := 0.0
	var reasons []string
	urgency := "wait"

	switch item.Status {
	case "flash-sale":
		score += 40
		urgency = "now"
		reasons = append(reasons, "Vente flash — achat immédiat recommandé")
	case "buy":
		score += 20
		urgency = "now"
		reasons = append(reasons, "Statut : prêt à acheter")
	case "watch":
		score += 10
		urgency = "soon"
		reasons = append(reasons, "À surveiller")
	case "crisis":
		score += 15
		urgency = "now"
		reasons = append(reasons, "Besoin urgent")
	case "recommended":
		score += 12
		urgency = "soon"
		reasons = append(reasons, "Article recommandé")
	default:
		// defer / preorder / unknown → urgency stays "wait"
		score += 0
	}

	// Target-price bonus.
	savingsPotential := 0.0
	if item.TargetPrice != nil {
		tp := *item.TargetPrice
		if item.Price <= tp {
			score += 30
			reasons = append(reasons, "Prix cible atteint")
		} else if item.Price > tp {
			savingsPotential = item.Price - tp
		}
	}

	// Base score: lower price = higher priority (capped at 0).
	base := 50.0 - (item.Price / 100.0)
	if base < 0 {
		base = 0
	}
	score += base

	return OptimizedItem{
		ItemID:           item.ID.String(),
		ItemName:         item.Name,
		Merchant:         item.Merchant,
		Price:            item.Price,
		PriorityScore:    score,
		Reasons:          reasons,
		UrgencyLevel:     urgency,
		SavingsPotential: savingsPotential,
	}
}

// buildSummary returns a French summary sentence.
func buildSummary(nowCount, soonCount, waitCount int, totalOptimized float64) string {
	switch {
	case nowCount == 0 && soonCount == 0:
		return "Aucun article urgent. Attendez de meilleures opportunités."
	case nowCount == 0:
		return fmt.Sprintf("%d article(s) à surveiller. Pas d'achat immédiat nécessaire.", soonCount)
	default:
		return fmt.Sprintf(
			"%d article(s) à acheter maintenant (%.0f €), %d à surveiller, %d en attente.",
			nowCount, totalOptimized, soonCount, waitCount,
		)
	}
}
