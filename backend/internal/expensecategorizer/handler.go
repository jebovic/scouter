package expensecategorizer

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 15 * time.Minute

type cacheEntry struct {
	resp      ExpenseSummary
	expiresAt time.Time
}

// Handler handles expense categorizer endpoints.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]cacheEntry
}

// NewHandler creates a new Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

var categoryMap = map[string]struct {
	label string
	icon  string
}{
	"electronique": {label: "📱 Électronique", icon: "📱"},
	"alimentaire":  {label: "🛒 Alimentaire", icon: "🛒"},
	"mode":         {label: "👗 Mode", icon: "👗"},
	"sport":        {label: "⚽ Sport", icon: "⚽"},
	"voyage":       {label: "✈️ Voyage", icon: "✈️"},
	"maison":       {label: "🏠 Maison", icon: "🏠"},
	"beaute":       {label: "💄 Beauté", icon: "💄"},
	"loisirs":      {label: "🎮 Loisirs", icon: "🎮"},
	"autre":        {label: "📦 Autre", icon: "📦"},
}

// categorizeCostCategory maps a cost_category string to one of the 9 expense categories.
func categorizeCostCategory(costCategory string) string {
	lc := strings.ToLower(costCategory)

	// Exact match first (fastest)
	if map[string]bool{
		"electronique": true, "alimentaire": true, "mode": true,
		"sport": true, "voyage": true, "maison": true,
		"beaute": true, "loisirs": true, "autre": true,
	}[lc] {
		return lc
	}

	// Keyword-based categorization
	// Check electronics
	electronicsKeywords := []string{"électro", "tech", "info", "phone", "tablet", "ordinat", "computer"}
	for _, kw := range electronicsKeywords {
		if strings.Contains(lc, kw) {
			return "electronique"
		}
	}

	// Check food
	foodKeywords := []string{"alim", "food", "épicerie", "resto", "course", "grocery"}
	for _, kw := range foodKeywords {
		if strings.Contains(lc, kw) {
			return "alimentaire"
		}
	}

	// Check sport (before mode to avoid confusion)
	sportKeywords := []string{"sport", "fitness", "gym", "athlet"}
	for _, kw := range sportKeywords {
		if strings.Contains(lc, kw) {
			return "sport"
		}
	}

	// Check fashion/mode (after sport)
	fashionKeywords := []string{"vêt", "mode", "habit", "chaussure", "vêtement", "clothes", "fashion"}
	for _, kw := range fashionKeywords {
		if strings.Contains(lc, kw) {
			return "mode"
		}
	}

	// Check travel
	travelKeywords := []string{"voyage", "hôtel", "hotel", "avion", "train", "transport", "trip"}
	for _, kw := range travelKeywords {
		if strings.Contains(lc, kw) {
			return "voyage"
		}
	}

	// Check home
	homeKeywords := []string{"maison", "jardin", "déco", "meuble", "cuisine", "home", "furniture"}
	for _, kw := range homeKeywords {
		if strings.Contains(lc, kw) {
			return "maison"
		}
	}

	// Check beauty
	beautyKeywords := []string{"beauté", "cosmétique", "soin", "parfum", "beauty", "cosmetic"}
	for _, kw := range beautyKeywords {
		if strings.Contains(lc, kw) {
			return "beaute"
		}
	}

	// Check leisure/entertainment
	leisureKeywords := []string{"loisir", "jeu", "livre", "musique", "ciné", "movie", "game", "entertainment"}
	for _, kw := range leisureKeywords {
		if strings.Contains(lc, kw) {
			return "loisirs"
		}
	}

	// Default to autre
	return "autre"
}

// GetSummary handles GET /api/missions/{missionId}/expense-summary.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	missionIDStr := chi.URLParam(r, "missionId")
	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	// Check cache
	if cached, ok := h.cache.Load(missionIDStr); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(missionIDStr)
	}

	// Query shopping items excluding 'defer' status
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, name, cost_category, price, status
		FROM shopping_items
		WHERE mission_id = $1 AND status != 'defer'
		ORDER BY name ASC
	`, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	// Accumulate category data
	breakdownMap := make(map[string]CategoryBreakdown)
	totalSpent := 0.0
	totalItems := 0

	for rows.Next() {
		var (
			id           uuid.UUID
			name         string
			costCategory string
			price        float64
			status       string
		)
		if err := rows.Scan(&id, &name, &costCategory, &price, &status); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		category := categorizeCostCategory(costCategory)
		totalSpent += price
		totalItems++

		bd := breakdownMap[category]
		bd.Category = category
		bd.Label = categoryMap[category].label
		bd.Icon = categoryMap[category].icon
		bd.TotalSpent += price
		bd.ItemCount++
		breakdownMap[category] = bd
	}

	if err = rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Compute percentages and build sorted breakdown list
	breakdowns := make([]CategoryBreakdown, 0, len(breakdownMap))
	for _, bd := range breakdownMap {
		if totalSpent > 0 {
			bd.Percentage = (bd.TotalSpent / totalSpent) * 100
		}
		breakdowns = append(breakdowns, bd)
	}

	// Sort by totalSpent descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].TotalSpent > breakdowns[j].TotalSpent
	})

	// Determine top category and build insight
	topCategory := "autre"
	insight := "Aucune dépense enregistrée pour cette mission"

	if len(breakdowns) > 0 {
		topCategory = breakdowns[0].Category
		if totalItems > 0 {
			topPct := int(breakdowns[0].Percentage)
			insight = fmt.Sprintf("Votre principale dépense est %s avec %d%% du budget", categoryMap[topCategory].label, topPct)
		}
	}

	summary := ExpenseSummary{
		MissionID:   missionID.String(),
		TotalItems:  totalItems,
		TotalSpent:  totalSpent,
		Breakdowns:  breakdowns,
		TopCategory: topCategory,
		Insight:     insight,
	}

	// Cache the response
	h.cache.Store(missionIDStr, cacheEntry{
		resp:      summary,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, summary)
}
