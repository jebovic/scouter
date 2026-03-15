package cashbacktracker

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

// merchantCashbackMap is a hardcoded registry of French cashback platforms and their rates.
// Keys are lowercased merchant names, values are [platform, cashbackPct].
var merchantCashbackMap = map[string][2]interface{}{
	"fnac":       {"iGraal", 4.5},
	"darty":      {"iGraal", 3.0},
	"boulanger":  {"Poulpeo", 2.5},
	"cdiscount":  {"iGraal", 5.0},
	"amazon":     {"Poulpeo", 1.5},
	"leclerc":    {"Cashback World", 2.0},
	"auchan":     {"Poulpeo", 3.5},
	"ldlc":       {"iGraal", 4.0},
}

// platformTips is a map of French tips for each platform.
var platformTips = map[string]string{
	"iGraal":         "💡 iGraal offre les meilleurs taux de cashback parmi les marchands français. Pensez à consulter leurs offres flash hebdomadaires!",
	"Poulpeo":        "💡 Poulpeo propose des bonus de bienvenue avantageux pour les nouveaux utilisateurs. Débutez par une inscription pour cumuler les gains!",
	"Cashback World": "💡 Cashback World est idéal pour les achats à l'étranger. Ses taux compétitifs couvrent de nombreuses enseignes internationales et françaises.",
}

// cacheEntry holds a cached response with its expiration time.
type cacheEntry struct {
	data      *CashbackSummary
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/cashback.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]cacheEntry
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// merchantRow represents a single aggregated merchant from the shopping_items query.
type merchantRow struct {
	merchant  string
	totalPrice float64
	itemCount int
}

// GetCashbackSummary handles GET /api/missions/{missionId}/cashback.
func (h *Handler) GetCashbackSummary(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid missionId")
		return
	}

	// Check cache first.
	cacheKey := missionID.String()
	if cached, ok := h.cache.Load(cacheKey); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		// Expired, will refetch below.
	}

	ctx := r.Context()

	// Fetch merchant aggregations from shopping_items.
	rows, err := h.pool.Query(ctx, `
		SELECT COALESCE(merchant, 'Inconnu') AS merchant, SUM(price) AS total_price, COUNT(*) AS item_count
		FROM shopping_items
		WHERE mission_id = $1
		GROUP BY merchant
		ORDER BY total_price DESC`,
		missionID,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query shopping items")
		return
	}
	defer rows.Close()

	var merchants []merchantRow
	for rows.Next() {
		var mr merchantRow
		if scanErr := rows.Scan(&mr.merchant, &mr.totalPrice, &mr.itemCount); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan merchant row")
			return
		}
		merchants = append(merchants, mr)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate merchant rows")
		return
	}

	// Build response.
	resp := buildCashbackResponse(missionID.String(), merchants)

	// Cache for 1 hour.
	h.cache.Store(cacheKey, cacheEntry{
		data:      resp,
		expiresAt: time.Now().Add(time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildCashbackResponse computes the CashbackSummary from raw merchant data. Pure function, testable without DB.
func buildCashbackResponse(missionID string, merchants []merchantRow) *CashbackSummary {
	// Compute cashback for each merchant.
	var cashbacks []MerchantCashback
	platformTotals := make(map[string]float64)

	for _, m := range merchants {
		merchantLower := strings.ToLower(m.merchant)

		// Look up cashback rate.
		var platform string
		var pct float64
		if rates, ok := merchantCashbackMap[merchantLower]; ok {
			platform = rates[0].(string)
			pct = rates[1].(float64)
		} else {
			// Default fallback.
			platform = "iGraal"
			pct = 2.0
		}

		estimatedAmount := m.totalPrice * (pct / 100)
		cashbacks = append(cashbacks, MerchantCashback{
			Merchant:        m.merchant,
			Platform:        platform,
			CashbackPct:     pct,
			EstimatedAmount: estimatedAmount,
			ItemCount:       m.itemCount,
		})

		// Accumulate by platform.
		platformTotals[platform] += estimatedAmount
	}

	// Compute total estimated and find best platform.
	var totalEstimated float64
	var bestPlatform string
	var bestAmount float64
	for platform, amount := range platformTotals {
		totalEstimated += amount
		if amount > bestAmount {
			bestAmount = amount
			bestPlatform = platform
		}
	}

	// Generate tip.
	tip := platformTips[bestPlatform]
	if tip == "" {
		tip = "💡 Utilisez un service de cashback pour maximiser vos économies sur vos achats."
	}

	return &CashbackSummary{
		MissionID:        missionID,
		MerchantCashback: cashbacks,
		TotalEstimated:   totalEstimated,
		BestPlatform:     bestPlatform,
		Tip:              tip,
	}
}
