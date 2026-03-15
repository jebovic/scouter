package negotiationscript

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = time.Hour

// knownFrenchMerchants is the set of well-known French merchants for contextual argument selection.
var knownFrenchMerchants = map[string]bool{
	"fnac":      true,
	"darty":     true,
	"cdiscount": true,
	"leclerc":   true,
	"carrefour": true,
	"auchan":    true,
	"boulanger": true,
	"ldlc":      true,
	"amazon":    true,
	"fnac.com":  true,
}

var staticTips = []string{
	"Mentionnez une offre concurrente précise pour plus de crédibilité",
	"Proposez de payer comptant si possible",
	"Restez poli et souriant — ça aide toujours",
}

type cacheEntry struct {
	resp      Response
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/items/{itemId}/negotiation-script.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetScript handles GET /api/missions/{missionId}/items/{itemId}/negotiation-script.
func (h *Handler) GetScript(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(key)
	}

	resp, found, err := h.build(r, missionID, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	h.cache.Store(key, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

type itemRow struct {
	id           uuid.UUID
	name         string
	price        float64
	targetPrice  *float64
	merchant     string
	costCategory string
	status       string
}

func (h *Handler) build(r *http.Request, missionID, itemID uuid.UUID) (Response, bool, error) {
	ctx := r.Context()

	// Fetch item details.
	var item itemRow
	var targetPriceVal *float64
	err := h.pool.QueryRow(ctx,
		`SELECT id, name, price, target_price, merchant, cost_category, status
		 FROM shopping_items
		 WHERE id = $1 AND mission_id = $2`,
		itemID, missionID,
	).Scan(&item.id, &item.name, &item.price, &targetPriceVal, &item.merchant, &item.costCategory, &item.status)
	if err != nil {
		// pgx returns ErrNoRows when no matching row exists — caller sees !found.
		if strings.Contains(err.Error(), "no rows") {
			return Response{}, false, nil
		}
		return Response{}, false, err
	}
	item.targetPrice = targetPriceVal

	// Fetch price history stats.
	var minPrice, avgPrice *float64
	_ = h.pool.QueryRow(ctx,
		`SELECT MIN(price), AVG(price) FROM price_history WHERE item_id = $1`,
		itemID,
	).Scan(&minPrice, &avgPrice)

	resp := buildResponse(item, minPrice, avgPrice)
	return resp, true, nil
}

func buildResponse(item itemRow, minPrice, avgPrice *float64) Response {
	script := buildScript(item, avgPrice)
	savings := computeSavings(item, minPrice)

	return Response{
		ItemID:           item.id.String(),
		Name:             item.name,
		Script:           script,
		Tips:             staticTips,
		EstimatedSavings: savings,
	}
}

func buildScript(item itemRow, avgPrice *float64) Script {
	return Script{
		Opening:      buildOpening(item, avgPrice),
		Argument:     buildArgument(item),
		CounterOffer: buildCounterOffer(item),
		Closing:      "Je suis prêt à effectuer l'achat aujourd'hui si nous trouvons un accord. Merci pour votre compréhension.",
	}
}

func buildOpening(item itemRow, avgPrice *float64) string {
	if avgPrice != nil && item.price < *avgPrice {
		return fmt.Sprintf(
			"Bonjour, je suis intéressé par %s à %.2f€. J'ai vu que ce produit était vendu en moyenne à %.2f€.",
			item.name, item.price, *avgPrice,
		)
	}
	return fmt.Sprintf(
		"Bonjour, je cherche %s. J'ai effectué une recherche approfondie des prix sur le marché.",
		item.name,
	)
}

func buildArgument(item itemRow) string {
	merchantLower := strings.ToLower(strings.TrimSpace(item.merchant))
	if knownFrenchMerchants[merchantLower] {
		return fmt.Sprintf(
			"En tant que client fidèle de %s, j'aimerais bénéficier d'un geste commercial.",
			item.merchant,
		)
	}
	target := item.price * 0.9
	if item.targetPrice != nil {
		target = *item.targetPrice
	}
	return fmt.Sprintf(
		"J'ai trouvé des offres concurrentes à %.2f€ chez d'autres marchands.",
		target,
	)
}

func buildCounterOffer(item itemRow) string {
	if item.targetPrice != nil {
		return fmt.Sprintf(
			"Seriez-vous en mesure de me proposer ce produit à %.2f€ ?",
			*item.targetPrice,
		)
	}
	return "Pourriez-vous faire un effort sur le prix ? Je peux décider d'acheter immédiatement."
}

func computeSavings(item itemRow, minPrice *float64) float64 {
	if item.targetPrice != nil {
		savings := item.price - *item.targetPrice
		if savings < 0 {
			savings = 0
		}
		return math.Round(savings*100) / 100
	}
	if minPrice != nil {
		savings := item.price - *minPrice
		if savings < 0 {
			savings = 0
		}
		return math.Round(savings*100) / 100
	}
	return 0
}
