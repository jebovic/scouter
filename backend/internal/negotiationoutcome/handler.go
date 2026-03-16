package negotiationoutcome

import (
	"hash/fnv"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

var merchantTips = map[string]string{
	"fnac":      "Mentionnez votre carte Fnac+ pour un geste commercial supplémentaire",
	"darty":     "Demandez l'extension de garantie offerte en cas d'achat immédiat",
	"cdiscount": "Utilisez les codes promo CDav pour la négo",
	"boulanger": "Proposez de souscrire à la carte Boulanger",
	"leclerc":   "Faites jouer la carte fidélité E.Leclerc",
	"carrefour": "Montrez l'offre concurrente Amazon ou Fnac",
	"amazon":    "Contactez le service client pour un remboursement de différence de prix",
	"ldlc":      "LDLC est ouvert au marchandage sur les accessoires",
}

const defaultTip = "Restez poli et proposez un paiement immédiat pour obtenir un geste commercial"

type cacheEntry struct {
	summary   Summary
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/negotiation-outcomes.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: missionID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetOutcomes handles GET /api/missions/{missionId}/negotiation-outcomes.
func (h *Handler) GetOutcomes(w http.ResponseWriter, r *http.Request) {
	missionParam := chi.URLParam(r, "missionId")
	var missionUUID string
	err := h.pool.QueryRow(r.Context(),
		`SELECT id::text FROM missions WHERE id::text = $1 OR slug = $1`,
		missionParam,
	).Scan(&missionUUID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	key := missionUUID
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.summary)
			return
		}
		h.cache.Delete(key)
	}

	summary, err := h.build(r, missionUUID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{summary: summary, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, summary)
}

type itemRow struct {
	id       string
	name     string
	merchant string
	price    float64
}

func (h *Handler) build(r *http.Request, missionID string) (Summary, error) {
	ctx := r.Context()

	rows, err := h.pool.Query(ctx,
		`SELECT id, name, merchant, price
		 FROM shopping_items
		 WHERE mission_id::text = $1
		 ORDER BY name`,
		missionID,
	)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(&it.id, &it.name, &it.merchant, &it.price); err != nil {
			return Summary{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}

	outcomes := make([]Outcome, 0, len(items))
	for _, it := range items {
		var minPrice *float64
		_ = h.pool.QueryRow(ctx,
			`SELECT MIN(price) FROM price_history WHERE item_id = $1`,
			it.id,
		).Scan(&minPrice)

		outcomes = append(outcomes, buildOutcome(it, minPrice))
	}

	return buildSummary(outcomes), nil
}

func buildOutcome(it itemRow, minPrice *float64) Outcome {
	merchantKey := strings.ToLower(strings.TrimSpace(it.merchant))

	// NegotiatedPrice = MIN(price_history.price) if lower than current price.
	var negotiatedPrice *float64
	var saved float64
	if minPrice != nil && *minPrice < it.price {
		v := math.Round(*minPrice*100) / 100
		negotiatedPrice = &v
		saved = math.Round((it.price-v)*100) / 100
	}

	successRate := fnvSuccessRate(merchantKey)
	attempts := fnvAttempts(it.id)
	tip := merchantTips[merchantKey]
	if tip == "" {
		tip = defaultTip
	}
	verdict := computeVerdict(saved, it.price)

	return Outcome{
		ItemID:          it.id,
		ItemName:        it.name,
		Merchant:        it.merchant,
		AskingPrice:     math.Round(it.price*100) / 100,
		NegotiatedPrice: negotiatedPrice,
		Saved:           saved,
		SuccessRate:     successRate,
		TipForMerchant:  tip,
		Verdict:         verdict,
		Attempts:        attempts,
	}
}

// fnvSuccessRate derives a deterministic success rate (0.4–1.0) from the merchant name.
func fnvSuccessRate(merchant string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(merchant))
	raw := h.Sum32() % 60 // 0–59
	return math.Round((float64(raw)+40)/100*100) / 100
}

// fnvAttempts derives a deterministic attempt count (1–3) from the item ID.
func fnvAttempts(itemID string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(itemID))
	return int(h.Sum32()%3) + 1
}

func computeVerdict(saved, askingPrice float64) string {
	if saved <= 0 {
		return "pending"
	}
	if askingPrice > 0 && saved/askingPrice < 0.05 {
		return "partial"
	}
	return "success"
}

func buildSummary(outcomes []Outcome) Summary {
	var totalSaved float64
	var totalSuccess float64
	var totalAttempts int
	bestMerchant := ""
	bestRate := -1.0

	for _, o := range outcomes {
		totalSaved += o.Saved
		totalSuccess += o.SuccessRate
		totalAttempts += o.Attempts
		if o.SuccessRate > bestRate {
			bestRate = o.SuccessRate
			bestMerchant = o.Merchant
		}
	}

	var avgSuccess float64
	if len(outcomes) > 0 {
		avgSuccess = math.Round(totalSuccess/float64(len(outcomes))*100) / 100
	}

	return Summary{
		Outcomes:      outcomes,
		TotalSaved:    math.Round(totalSaved*100) / 100,
		AvgSuccess:    avgSuccess,
		BestMerchant:  bestMerchant,
		TotalAttempts: totalAttempts,
	}
}
