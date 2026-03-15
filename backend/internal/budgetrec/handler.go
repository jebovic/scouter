package budgetrec

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// ItemLister is the minimal interface required from the shopping service.
type ItemLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler serves the budget analysis endpoint.
type Handler struct {
	items ItemLister
}

// NewHandler creates a Handler wired to the given shopping item lister.
func NewHandler(items ItemLister) *Handler {
	return &Handler{items: items}
}

// GetAnalysis handles GET /api/missions/{missionID}/budget-analysis.
func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	missionIDStr := chi.URLParam(r, "missionID")
	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission ID")
		return
	}

	items, err := h.items.ListByMission(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list items")
		return
	}

	analysis := analyze(missionIDStr, items)
	httputil.WriteJSON(w, http.StatusOK, analysis)
}

// analyze is the pure computation function; it is package-level to facilitate
// direct unit testing without HTTP scaffolding.
func analyze(missionID string, items []shopping.Item) BudgetAnalysis {
	type catStats struct {
		budget float64
		spent  float64
	}

	cats := map[string]*catStats{}
	var totalBudget, totalSpent float64

	for _, it := range items {
		if _, ok := cats[it.CostCategory]; !ok {
			cats[it.CostCategory] = &catStats{}
		}
		// Reference budget: OriginalEstimate when available, else current price.
		budget := it.Price
		if it.OriginalEstimate != nil {
			budget = *it.OriginalEstimate
		}
		spent := it.Price

		cats[it.CostCategory].budget += budget
		cats[it.CostCategory].spent += spent
		totalBudget += budget
		totalSpent += spent
	}

	var recs []BudgetRecommendation
	var potentialSavings float64

	for cat, s := range cats {
		switch {
		case s.spent > s.budget*1.1: // overspent by >10 %
			save := s.spent - s.budget
			prio := PriorityMedium
			if s.spent > s.budget*1.25 {
				prio = PriorityHigh
			}
			recs = append(recs, BudgetRecommendation{
				Category:          cat,
				CurrentBudget:     s.budget,
				RecommendedBudget: s.budget * 1.05,
				Reason:            "Dépenses supérieures au budget initial — révisez les priorités",
				Priority:          prio,
				SaveAmount:        save * 0.5,
			})
			potentialSavings += save * 0.5

		case s.budget > 0 && s.spent < s.budget*0.6: // underspent — reallocate
			save := s.budget - s.spent*1.1
			recs = append(recs, BudgetRecommendation{
				Category:          cat,
				CurrentBudget:     s.budget,
				RecommendedBudget: s.spent * 1.1,
				Reason:            "Budget sous-utilisé — réaffectez vers les postes prioritaires",
				Priority:          PriorityLow,
				SaveAmount:        save,
			})
			potentialSavings += save
		}
	}

	// Health score: 100 when on budget, penalised for overruns.
	healthScore := 100
	if totalBudget > 0 {
		ratio := totalSpent / totalBudget
		if ratio > 1 {
			penalty := int((ratio - 1) * 100)
			if penalty > 80 {
				penalty = 80
			}
			healthScore = 100 - penalty
		}
	}

	return BudgetAnalysis{
		MissionID:        missionID,
		TotalBudget:      totalBudget,
		TotalSpent:       totalSpent,
		Recommendations:  recs,
		PotentialSavings: potentialSavings,
		HealthScore:      healthScore,
	}
}
