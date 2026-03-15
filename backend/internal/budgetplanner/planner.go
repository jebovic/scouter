package budgetplanner

import (
	"fmt"
	"sort"
	"time"

	"github.com/jibei/scouter/internal/shopping"
)

// statusScore returns the base priority score for an item status.
func statusScore(status string) int {
	switch status {
	case "flash-sale":
		return 50
	case "buy":
		return 30
	case "watch":
		return 10
	case "defer", "crisis":
		return -10
	default:
		return 0
	}
}

// frenchReason returns a human-readable French explanation for why an item
// was placed in the sequence or deferred.
func frenchReason(item shopping.Item, priority int, inBudget bool) string {
	if !inBudget {
		switch item.Status {
		case "defer":
			return "Article reporté — hors budget disponible"
		case "crisis":
			return "Achat en crise — reporté faute de budget"
		default:
			return fmt.Sprintf("Budget insuffisant pour inclure cet article (%.0f €)", item.Price)
		}
	}

	switch item.Status {
	case "flash-sale":
		return "Vente flash — achetez immédiatement avant que l'offre expire"
	case "buy":
		if item.TargetPrice != nil && item.Price < *item.TargetPrice {
			return fmt.Sprintf("Prix cible atteint (%.0f € < cible %.0f €) — achat recommandé", item.Price, *item.TargetPrice)
		}
		return "Prêt à l'achat — priorité confirmée"
	case "watch":
		if item.TargetPrice != nil && item.Price < *item.TargetPrice {
			return fmt.Sprintf("Prix inférieur à la cible (%.0f € < %.0f €) — saisir l'opportunité", item.Price, *item.TargetPrice)
		}
		return "Surveillance active — inclus dans le budget disponible"
	case "crisis":
		return "Achat urgent malgré le statut crise — reste dans le budget"
	case "defer":
		return "Reporté mais inclus — budget disponible"
	default:
		return "Inclus dans le plan d'achat"
	}
}

// computePriority calculates a numeric priority score for a shopping item.
// Higher values are purchased first.
func computePriority(item shopping.Item) int {
	score := statusScore(item.Status)
	if item.TargetPrice != nil && item.Price < *item.TargetPrice {
		score += 20
	}
	return score
}

// buildBudgetStatus classifies the overall budget health.
func buildBudgetStatus(budget, estimate float64) string {
	if budget <= 0 {
		return "over"
	}
	ratio := estimate / budget
	switch {
	case ratio <= 0.9:
		return "under"
	case ratio <= 1.05:
		return "tight"
	default:
		return "over"
	}
}

// Plan takes a mission budget and its shopping items and returns a sequenced
// PurchasePlan using a greedy fit strategy sorted by priority score then price.
func Plan(missionID string, budget float64, items []shopping.Item) PurchasePlan {
	type scored struct {
		item     shopping.Item
		priority int
	}

	scored_items := make([]scored, len(items))
	for i, it := range items {
		scored_items[i] = scored{item: it, priority: computePriority(it)}
	}

	// Sort by priority descending, then price ascending as tie-breaker.
	sort.Slice(scored_items, func(i, j int) bool {
		if scored_items[i].priority != scored_items[j].priority {
			return scored_items[i].priority > scored_items[j].priority
		}
		return scored_items[i].item.Price < scored_items[j].item.Price
	})

	var sequence []PlannedItem
	var deferred []PlannedItem
	var totalEstimate float64
	var savings float64
	remaining := budget

	for _, si := range scored_items {
		it := si.item
		totalEstimate += it.Price

		if it.TargetPrice != nil && it.Price < *it.TargetPrice {
			savings += *it.TargetPrice - it.Price
		}

		inBudget := remaining >= it.Price
		if inBudget {
			remaining -= it.Price
			cumulative := budget - remaining
			sequence = append(sequence, PlannedItem{
				ItemID:         it.ID.String(),
				ItemName:       it.Name,
				Price:          it.Price,
				Status:         it.Status,
				Priority:       si.priority,
				Reason:         frenchReason(it, si.priority, true),
				CumulativeCost: cumulative,
			})
		} else {
			deferred = append(deferred, PlannedItem{
				ItemID:         it.ID.String(),
				ItemName:       it.Name,
				Price:          it.Price,
				Status:         it.Status,
				Priority:       si.priority,
				Reason:         frenchReason(it, si.priority, false),
				CumulativeCost: 0,
			})
		}
	}

	// Ensure nil slices serialize as empty JSON arrays.
	if sequence == nil {
		sequence = []PlannedItem{}
	}
	if deferred == nil {
		deferred = []PlannedItem{}
	}

	return PurchasePlan{
		MissionID:     missionID,
		TotalBudget:   budget,
		TotalEstimate: totalEstimate,
		BudgetStatus:  buildBudgetStatus(budget, totalEstimate),
		Sequence:      sequence,
		Deferred:      deferred,
		Savings:       savings,
		GeneratedAt:   time.Now().UTC(),
	}
}
