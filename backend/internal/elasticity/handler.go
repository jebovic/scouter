package elasticity

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = time.Hour

type cacheEntry struct {
	result    ElasticityResult
	expiresAt time.Time
}

// Handler exposes the elasticity endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetElasticity handles GET /api/missions/{missionId}/items/{itemId}/elasticity.
func (h *Handler) GetElasticity(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.result)
			return
		}
	}

	result, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, result)
}

type priceRecord struct {
	price      float64
	recordedAt time.Time
}

func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (ElasticityResult, error) {
	ctx := r.Context()

	rows, err := h.pool.Query(ctx, `
		SELECT price, recorded_at
		FROM price_history
		WHERE item_id = $1
		ORDER BY recorded_at ASC`, itemID)
	if err != nil {
		return ElasticityResult{}, err
	}
	defer rows.Close()

	var records []priceRecord
	for rows.Next() {
		var rec priceRecord
		if err := rows.Scan(&rec.price, &rec.recordedAt); err != nil {
			return ElasticityResult{}, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return ElasticityResult{}, err
	}

	insufficient := ElasticityResult{
		ItemID:      itemID.String(),
		Elasticity:  nil,
		Insight:     "Données insuffisantes",
		PricePoints: len(records),
	}

	if len(records) < 5 {
		return insufficient, nil
	}

	// Compute price elasticity using consecutive readings.
	type elasticityPoint struct {
		deltaP float64
		deltaQ float64
	}

	var points []elasticityPoint
	for i := 1; i < len(records); i++ {
		p1 := records[i-1].price
		p2 := records[i].price
		if p1 == 0 {
			continue
		}
		deltaP := (p2 - p1) / p1

		// Demand proxy: price drop >2% → demand up; price rise >2% → demand down; stable → 1.0
		var demandChange float64
		if deltaP < -0.02 {
			demandChange = math.Abs(deltaP) * 2.0
		} else if deltaP > 0.02 {
			demandChange = -math.Abs(deltaP) * 0.8
		}
		// deltaQ = demandChange relative to baseline demand of 1.0
		deltaQ := demandChange

		if deltaP == 0 {
			continue
		}
		points = append(points, elasticityPoint{deltaP: deltaP, deltaQ: deltaQ})
	}

	if len(points) == 0 {
		return insufficient, nil
	}

	// Average elasticity = mean(deltaQ / deltaP).
	sum := 0.0
	for _, pt := range points {
		sum += pt.deltaQ / pt.deltaP
	}
	avgElasticity := sum / float64(len(points))

	// Normalize to -3 to 3 range.
	avgElasticity = math.Max(-3.0, math.Min(3.0, avgElasticity))
	avgElasticity = math.Round(avgElasticity*100) / 100

	// Classify.
	var classification, insight string
	switch {
	case avgElasticity > 1.5:
		classification = "Très élastique"
		insight = "Très élastique — les promotions ont un fort impact"
	case avgElasticity >= 0.5:
		classification = "Élastique"
		insight = "Élastique — sensible aux variations de prix"
	case avgElasticity >= -0.5:
		classification = "Inélastique"
		insight = "Inélastique — le prix a peu d'effet sur la demande"
	default:
		classification = "Fortement inélastique"
		insight = "Fortement inélastique — achetez au meilleur prix sans attendre"
	}

	// Optimal buy threshold: last price * (1 - min(0.2, abs(elasticity) * 0.05)).
	lastPrice := records[len(records)-1].price
	discount := math.Min(0.2, math.Abs(avgElasticity)*0.05)
	threshold := math.Round(lastPrice*(1-discount)*100) / 100

	e := avgElasticity
	t := threshold

	return ElasticityResult{
		ItemID:              itemID.String(),
		Elasticity:          &e,
		Classification:      classification,
		Insight:             insight,
		OptimalBuyThreshold: &t,
		PricePoints:         len(records),
	}, nil
}
