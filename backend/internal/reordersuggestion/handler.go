package reordersuggestion

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 10 * time.Minute

// frenchLoyaltyMerchants is the set of French merchants with well-known
// loyalty programmes (case-insensitive prefix match).
var frenchLoyaltyMerchants = []string{
	"fnac", "darty", "leclerc", "carrefour", "auchan",
	"intermarché", "intermarche", "boulanger", "cdiscount",
	"la redoute", "amazon.fr", "monoprix",
}

type cacheEntry struct {
	resp      Response
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/reorder-suggestions.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

const itemsQuery = `
	SELECT id::text, name, price, target_price, status, cost_category, merchant
	FROM shopping_items
	WHERE mission_id = $1
	AND   status IN ('buy','flash-sale','crisis','preorder')
`

// GetSuggestions handles GET /api/missions/{missionId}/reorder-suggestions.
func (h *Handler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId is required")
		return
	}

	// Serve from cache.
	h.mu.Lock()
	if entry, ok := h.cache[missionID]; ok && time.Now().Before(entry.expiresAt) {
		resp := entry.resp
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, resp)
		return
	}
	h.mu.Unlock()

	resp, err := h.compute(r, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.mu.Lock()
	h.cache[missionID] = cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) compute(r *http.Request, missionID string) (Response, error) {
	rows, err := h.pool.Query(r.Context(), itemsQuery, missionID)
	if err != nil {
		return Response{}, err
	}
	defer rows.Close()

	var items []rawItem
	for rows.Next() {
		var it rawItem
		if err := rows.Scan(
			&it.id, &it.name, &it.price, &it.targetPrice,
			&it.status, &it.costCategory, &it.merchant,
		); err != nil {
			return Response{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return Response{}, err
	}

	suggestions := scoreItems(items)

	// Sort descending by score; stable ties by name.
	sort.SliceStable(suggestions, func(i, j int) bool {
		if suggestions[i].BuyScore != suggestions[j].BuyScore {
			return suggestions[i].BuyScore > suggestions[j].BuyScore
		}
		return suggestions[i].Name < suggestions[j].Name
	})

	batches := buildMerchantBatches(suggestions)

	total := 0.0
	for _, s := range suggestions {
		total += s.Price
	}

	if suggestions == nil {
		suggestions = []Suggestion{}
	}
	if batches == nil {
		batches = []MerchantBatch{}
	}

	return Response{
		MissionID:       missionID,
		Suggestions:     suggestions,
		MerchantBatches: batches,
		TotalToBuy:      total,
	}, nil
}

// scoreItems computes a Suggestion for each item with its buy-now score.
func scoreItems(items []rawItem) []Suggestion {
	out := make([]Suggestion, 0, len(items))
	for _, it := range items {
		score, reasoning := computeScore(it)
		priority := scoreToPriority(score)
		out = append(out, Suggestion{
			ItemID:    it.id,
			Name:      it.name,
			Price:     it.price,
			Status:    it.status,
			BuyScore:  score,
			Priority:  priority,
			Reasoning: reasoning,
			Merchant:  it.merchant,
		})
	}
	return out
}

// computeScore returns the buy-now score and primary reasoning string for an item.
func computeScore(it rawItem) (int, string) {
	score := 0
	var reasons []string

	switch it.status {
	case "flash-sale":
		score += 100
		reasons = append(reasons, "Vente flash — achetez maintenant")
	case "crisis":
		score += 90
		reasons = append(reasons, "Situation urgente")
	}

	if it.targetPrice != nil {
		tp := *it.targetPrice
		if it.price <= tp {
			score += 30
			reasons = append(reasons, "Prix sous objectif")
		} else if it.price <= tp*1.05 {
			score += 15
			reasons = append(reasons, "Prix proche de l'objectif (-5%)")
		}
	}

	if it.price > 1000 {
		score += 10
		reasons = append(reasons, "Article haute valeur — bloquez le prix")
	}

	if hasFrenchLoyalty(it.merchant) {
		score += 5
		reasons = append(reasons, "Programme fidélité disponible")
	}

	// "buy" status base score when not already scored via flash-sale/crisis.
	if it.status == "buy" && score == 0 {
		score += 20
		reasons = append(reasons, "Prêt à acheter")
	}
	if it.status == "preorder" && score == 0 {
		score += 10
		reasons = append(reasons, "Précommande disponible")
	}

	reasoning := strings.Join(reasons, " · ")
	if reasoning == "" {
		reasoning = "Aucune urgence particulière"
	}
	return score, reasoning
}

// scoreToPriority maps a numeric score to a French-language priority label.
func scoreToPriority(score int) string {
	switch {
	case score >= 80:
		return "immédiat"
	case score >= 50:
		return "cette semaine"
	case score >= 20:
		return "ce mois"
	default:
		return "flexible"
	}
}

// hasFrenchLoyalty returns true when the merchant name matches a known French
// loyalty programme (case-insensitive prefix or full match).
func hasFrenchLoyalty(merchant string) bool {
	lower := strings.ToLower(strings.TrimSpace(merchant))
	for _, m := range frenchLoyaltyMerchants {
		if strings.HasPrefix(lower, m) || lower == m {
			return true
		}
	}
	return false
}

// buildMerchantBatches groups suggestions by merchant and returns batches for
// merchants with 2+ items so the buyer can consolidate orders.
func buildMerchantBatches(suggestions []Suggestion) []MerchantBatch {
	// Accumulate item names per merchant (preserving insertion order).
	order := []string{}
	groups := map[string][]string{}
	for _, s := range suggestions {
		m := s.Merchant
		if m == "" {
			continue
		}
		if _, seen := groups[m]; !seen {
			order = append(order, m)
		}
		groups[m] = append(groups[m], s.Name)
	}

	var batches []MerchantBatch
	for _, m := range order {
		names := groups[m]
		if len(names) < 2 {
			continue
		}
		batches = append(batches, MerchantBatch{
			Merchant:  m,
			Items:     names,
			BatchNote: "Regroupez vos achats chez " + m + " pour économiser sur la livraison",
		})
	}
	return batches
}
