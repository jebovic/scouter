package conditionpricing

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// conditionSpec describes a refurbished grade/condition tier.
type conditionSpec struct {
	condition      string
	grade          string
	priceFactor    float64 // multiplier vs new price
	platforms      []string
	warrantyMonths int
}

var conditionSpecs = []conditionSpec{
	{
		condition:      "Comme neuf",
		grade:          "A+",
		priceFactor:    0.80,
		platforms:      []string{"Back Market"},
		warrantyMonths: 12,
	},
	{
		condition:      "Très bon état",
		grade:          "A",
		priceFactor:    0.70,
		platforms:      []string{"Back Market", "Recommerce"},
		warrantyMonths: 12,
	},
	{
		condition:      "Bon état",
		grade:          "B",
		priceFactor:    0.60,
		platforms:      []string{"Recommerce", "Fnac Occasion"},
		warrantyMonths: 6,
	},
	{
		condition:      "Acceptable",
		grade:          "C",
		priceFactor:    0.50,
		platforms:      []string{"Fnac Occasion"},
		warrantyMonths: 6,
	},
}

// searchURLTemplates maps platform names to their search URL templates.
var searchURLTemplates = map[string]string{
	"Back Market":   "https://www.backmarket.fr/fr-fr/search?q=%s",
	"Recommerce":    "https://www.recommerce.com/fr/recherche?q=%s",
	"Fnac Occasion": "https://www.fnac.com/SearchResult/ResultList.aspx?Search=%s&sa=0&ref=Search_d_Nav&esl-k=sem|nl|&bl=occasion",
}

// availabilityOptions lists French availability labels.
var availabilityOptions = []string{"Disponible", "Stock limité", "Indisponible"}

// itemGetter retrieves a shopping item by ID.
type itemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

type pgItemGetter struct {
	pool *pgxpool.Pool
}

func (g *pgItemGetter) GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error) {
	return shopping.NewRepository(g.pool).GetByID(ctx, id)
}

// cacheEntry wraps a ConditionReport with its expiry.
type cacheEntry struct {
	report    *ConditionReport
	expiresAt time.Time
}

// Handler exposes the condition-pricing endpoint.
type Handler struct {
	getter itemGetter
	cache  sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{getter: &pgItemGetter{pool: pool}}
}

// GetConditionPricing handles GET /api/missions/{missionId}/items/{itemId}/condition-pricing.
func (h *Handler) GetConditionPricing(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 2-hour cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.report)
			return
		}
		h.cache.Delete(rawItemID)
	}

	item, err := h.getter.GetByID(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	report := generate(item)
	h.cache.Store(rawItemID, &cacheEntry{
		report:    report,
		expiresAt: time.Now().Add(2 * time.Hour),
	})
	httputil.WriteJSON(w, http.StatusOK, report)
}

// generate builds a deterministic ConditionReport for item using FNV-32a hashing.
func generate(item *shopping.Item) *ConditionReport {
	newPrice := item.Price
	encodedName := url.QueryEscape(item.Name)

	conditions := make([]ConditionPrice, 0, len(conditionSpecs))
	var bestDeal *ConditionPrice
	bestValue := -1.0 // best score = highest discount (value for money)

	for _, spec := range conditionSpecs {
		// Pick platform deterministically using FNV hash of itemID + grade.
		h := fnv.New32a()
		h.Write([]byte(item.ID.String()))
		h.Write([]byte(spec.grade))
		seed := int(h.Sum32())

		// Select platform.
		platform := spec.platforms[seed%len(spec.platforms)]

		// Compute price (round to 2 decimal places).
		estPrice := math.Round(newPrice*spec.priceFactor*100) / 100

		// Discount percentage off vs new price.
		discount := 0.0
		if newPrice > 0 {
			discount = math.Round((1-spec.priceFactor)*10000) / 100
		}

		// Availability: use seed%3 to pick from the three options, but
		// weight toward "Disponible" (indices 0 and 1) since grade A/A+ are
		// usually available.
		availIdx := seed % 3
		if availIdx == 2 && (spec.grade == "A+" || spec.grade == "A") {
			// Demote "Indisponible" for top grades to "Stock limité".
			availIdx = 1
		}
		availability := availabilityOptions[availIdx]

		// Build search URL.
		tmpl := searchURLTemplates[platform]
		searchUrl := fmt.Sprintf(tmpl, encodedName)

		cp := ConditionPrice{
			Condition:      spec.condition,
			Grade:          spec.grade,
			EstPrice:       estPrice,
			Discount:       discount,
			Platform:       platform,
			Availability:   availability,
			WarrantyMonths: spec.warrantyMonths,
			SearchUrl:      searchUrl,
		}
		conditions = append(conditions, cp)

		// Best deal: highest discount among available grades.
		if availability != "Indisponible" && discount > bestValue {
			bestValue = discount
			clone := cp
			bestDeal = &clone
		}
	}

	recommendation := buildRecommendation(item.Name, newPrice, bestDeal)

	return &ConditionReport{
		ItemId:         item.ID.String(),
		ItemName:       item.Name,
		NewPrice:       newPrice,
		Conditions:     conditions,
		BestDeal:       bestDeal,
		Recommendation: recommendation,
		GeneratedAt:    time.Now().UTC(),
	}
}

// buildRecommendation returns a French recommendation string.
func buildRecommendation(itemName string, newPrice float64, best *ConditionPrice) string {
	if best == nil {
		return fmt.Sprintf("Aucune offre reconditionnée disponible pour %s pour le moment.", itemName)
	}
	savings := math.Round((newPrice-best.EstPrice)*100) / 100
	return fmt.Sprintf(
		"Pour %s, le meilleur rapport qualité/prix reconditionné est en grade %s (%s) sur %s à %.2f€ — économisez %.2f€ (%.0f%% de réduction) avec une garantie de %d mois.",
		itemName, best.Grade, best.Condition, best.Platform, best.EstPrice, savings, best.Discount, best.WarrantyMonths,
	)
}
