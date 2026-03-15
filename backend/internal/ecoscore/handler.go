package ecoscore

import (
	"context"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// ItemLister fetches shopping items for a mission.
type ItemLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler serves GET /api/missions/{missionID}/eco-score.
type Handler struct {
	items ItemLister
}

// NewHandler creates a Handler wired to the given item lister.
func NewHandler(items ItemLister) *Handler {
	return &Handler{items: items}
}

// GetEcoScore handles GET /api/missions/{missionID}/eco-score.
func (h *Handler) GetEcoScore(w http.ResponseWriter, r *http.Request) {
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

	score := compute(missionIDStr, items)
	httputil.WriteJSON(w, http.StatusOK, score)
}

// compute is the pure calculation function, separated for testability.
func compute(missionID string, items []shopping.Item) EcoScore {
	itemScores := make([]ItemEcoScore, 0, len(items))
	var totalCO2 float64
	var sustainableShopping []string

	for _, it := range items {
		co2 := estimateCO2(it.CostCategory, it.Price)
		grade := gradeForCO2(co2)

		itemScores = append(itemScores, ItemEcoScore{
			ItemID:   it.ID.String(),
			ItemName: it.Name,
			Category: it.CostCategory,
			CO2kg:    roundTwo(co2),
			Grade:    grade,
		})
		totalCO2 += co2

		if co2 < 20 {
			sustainableShopping = append(sustainableShopping, it.Name)
		}
	}

	if sustainableShopping == nil {
		sustainableShopping = []string{}
	}

	totalCO2 = roundTwo(totalCO2)
	ecoGrade := gradeForCO2(totalCO2)
	gradeColor := colorForGrade(ecoGrade)
	trees := int(math.Ceil(totalCO2 / 25))
	if trees < 0 {
		trees = 0
	}
	tips := tipsForGrade(ecoGrade)

	return EcoScore{
		MissionID:               missionID,
		TotalCO2kg:              totalCO2,
		EcoGrade:                ecoGrade,
		GradeColor:              gradeColor,
		ItemScores:              itemScores,
		SustainableShopping:     sustainableShopping,
		EcoTips:                 tips,
		CompensationTreesNeeded: trees,
	}
}

// estimateCO2 returns a CO2 estimate in kg using category + price heuristics.
func estimateCO2(category string, price float64) float64 {
	cat := strings.ToLower(strings.TrimSpace(category))
	switch {
	case strings.Contains(cat, "electron") || strings.Contains(cat, "tech") || strings.Contains(cat, "informatique"):
		return 80 + price*0.5
	case strings.Contains(cat, "cloth") || strings.Contains(cat, "mode") || strings.Contains(cat, "v\u00eatement") || strings.Contains(cat, "textile"):
		return 15 + price*0.3
	case strings.Contains(cat, "food") || strings.Contains(cat, "aliment") || strings.Contains(cat, "cuisine") || strings.Contains(cat, "repas"):
		return 5
	case strings.Contains(cat, "transport") || strings.Contains(cat, "voyage") || strings.Contains(cat, "travel"):
		return 200
	case strings.Contains(cat, "sport") || strings.Contains(cat, "fitness"):
		return 20 + price*0.1
	default:
		return 30 + price*0.2
	}
}

// gradeForCO2 returns A–E grade based on total CO2 kg.
func gradeForCO2(co2 float64) string {
	switch {
	case co2 < 50:
		return "A"
	case co2 < 150:
		return "B"
	case co2 < 300:
		return "C"
	case co2 < 500:
		return "D"
	default:
		return "E"
	}
}

// colorForGrade maps grade letters to CSS-friendly color names.
func colorForGrade(grade string) string {
	switch grade {
	case "A":
		return "green"
	case "B":
		return "lime"
	case "C":
		return "amber"
	case "D":
		return "orange"
	default:
		return "red"
	}
}

// tipsForGrade returns three French eco-tips based on the mission grade.
func tipsForGrade(grade string) []string {
	switch grade {
	case "A":
		return []string{
			"Bravo ! Vos achats ont un faible impact environnemental.",
			"Pensez à l'économie circulaire pour maintenir votre score.",
			"Partagez vos bonnes pratiques avec votre entourage.",
		}
	case "B":
		return []string{
			"Privilégiez les produits reconditionnés pour réduire votre empreinte.",
			"Choisissez des marques éco-responsables et locales.",
			"Limitez les achats d'appareils électroniques neufs.",
		}
	case "C":
		return []string{
			"Réduisez vos achats électroniques ou optez pour le reconditionné.",
			"Compensez votre empreinte carbone en plantant des arbres locaux.",
			"Favorisez les transports en commun pour vos déplacements d'achat.",
		}
	case "D":
		return []string{
			"Votre empreinte carbone est élevée — réduisez les achats non essentiels.",
			"Optez pour la location ou le partage plutôt que l'achat neuf.",
			"Compensez vos émissions via un programme de reforestation certifié.",
		}
	default:
		return []string{
			"Impact très élevé — envisagez de reporter ou d'annuler certains achats.",
			"Renseignez-vous sur les alternatives durables disponibles.",
			"Une compensation carbone complète est recommandée pour cette mission.",
		}
	}
}

// roundTwo rounds a float64 to two decimal places.
func roundTwo(v float64) float64 {
	return math.Round(v*100) / 100
}
