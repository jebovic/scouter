package pricestreak

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

type cacheEntry struct {
	streak    PriceStreak
	expiresAt time.Time
}

// Handler exposes the price-streak endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetStreak handles GET /api/missions/{missionId}/items/{itemId}/price-streak.
func (h *Handler) GetStreak(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	// Check cache.
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.streak)
			return
		}
	}

	streak, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		streak:    streak,
		expiresAt: time.Now().Add(time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, streak)
}

// pricePoint holds a single price history record.
type pricePoint struct {
	price     float64
	createdAt time.Time
}

// compute queries price_history and derives a PriceStreak.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (PriceStreak, error) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT price, created_at
		FROM price_history
		WHERE item_id = $1
		ORDER BY created_at ASC`, itemID)
	if err != nil {
		return PriceStreak{}, err
	}
	defer rows.Close()

	var points []pricePoint
	for rows.Next() {
		var pp pricePoint
		if err := rows.Scan(&pp.price, &pp.createdAt); err != nil {
			return PriceStreak{}, err
		}
		points = append(points, pp)
	}
	if err := rows.Err(); err != nil {
		return PriceStreak{}, err
	}

	result := PriceStreak{
		ItemID:      itemID.String(),
		StreakType:  "stable",
		Urgency:     "low",
		GeneratedAt: time.Now().UTC(),
	}

	// Need at least 3 data points for meaningful analysis.
	if len(points) < 3 {
		result.Recommendation = "Pas assez de données pour détecter une tendance. Continuez à surveiller les prix."
		return result, nil
	}

	// Compute day-over-day changes from the end of the series.
	// Walk backwards to count consecutive drops or rises.
	n := len(points)

	// Determine direction of the last change to know what streak type to look for.
	lastIdx := n - 1
	streak := 0
	streakStartIdx := lastIdx

	// Check if we have a dropping streak (each price < previous).
	dropping := true
	rising := true

	for i := lastIdx; i >= 1; i-- {
		diff := points[i].price - points[i-1].price
		if diff >= 0 {
			// Not a drop — stop counting drops.
			dropping = false
		}
		if diff <= 0 {
			// Not a rise — stop counting rises.
			rising = false
		}
		if !dropping && !rising {
			break
		}
		streak++
		streakStartIdx = i - 1
	}

	// If there was no consistent direction for more than 0 steps.
	if streak == 0 || (!dropping && !rising) {
		result.Recommendation = "Les prix sont stables. Pas de tendance significative détectée."
		return result, nil
	}

	// Compute cumulative change over the streak.
	startPrice := points[streakStartIdx].price
	endPrice := points[lastIdx].price
	totalAmt := endPrice - startPrice
	totalPct := 0.0
	if startPrice > 0 {
		totalPct = (totalAmt / startPrice) * 100
	}

	result.CurrentStreak = streak
	result.StreakStart = points[streakStartIdx].createdAt.UTC()
	result.StreakStartPrice = startPrice
	result.TotalDropAmt = totalAmt
	result.TotalDropPct = totalPct

	switch {
	case dropping:
		result.StreakType = "dropping"
		result = applyDroppingRecommendation(result)
	case rising:
		result.StreakType = "rising"
		result = applyRisingRecommendation(result)
	}

	return result, nil
}

// applyDroppingRecommendation sets urgency and French recommendation for a dropping streak.
func applyDroppingRecommendation(s PriceStreak) PriceStreak {
	switch {
	case s.CurrentStreak >= 5:
		s.Urgency = "urgent"
		s.Recommendation = "Baisse prolongée depuis 5 jours ou plus — forte tendance à la baisse. Le moment idéal pour acheter est maintenant avant un possible rebond."
	case s.CurrentStreak >= 3:
		s.Urgency = "high"
		s.Recommendation = "Le prix baisse depuis plusieurs jours consécutifs. C'est probablement un bon moment pour acheter avant que la tendance ne s'inverse."
	case s.CurrentStreak >= 2:
		s.Urgency = "medium"
		s.Recommendation = "Légère tendance à la baisse détectée. Continuez à surveiller les prix avant de prendre une décision."
	default:
		s.Urgency = "low"
		s.Recommendation = "Une légère baisse de prix a été observée. Pas encore de tendance claire."
	}
	return s
}

// applyRisingRecommendation sets urgency and French recommendation for a rising streak.
func applyRisingRecommendation(s PriceStreak) PriceStreak {
	switch {
	case s.CurrentStreak >= 5:
		s.Urgency = "urgent"
		s.Recommendation = "Le prix monte depuis 5 jours ou plus. Achetez maintenant avant une nouvelle hausse si vous en avez besoin."
	case s.CurrentStreak >= 3:
		s.Urgency = "high"
		s.Recommendation = "Achetez maintenant avant une nouvelle hausse — le prix est en hausse constante depuis plusieurs jours."
	case s.CurrentStreak >= 2:
		s.Urgency = "medium"
		s.Recommendation = "Légère tendance à la hausse. Considérez d'acheter bientôt si ce produit vous intéresse."
	default:
		s.Urgency = "low"
		s.Recommendation = "Une légère hausse de prix a été observée. Restez vigilant."
	}
	return s
}
