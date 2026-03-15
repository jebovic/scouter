package priceannotation

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 2 * time.Hour

type cacheEntry struct {
	result    PriceAnnotations
	expiresAt time.Time
}

// Handler exposes the price-annotations endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetAnnotations handles GET /api/missions/{missionId}/items/{itemId}/price-annotations.
func (h *Handler) GetAnnotations(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	// Serve from cache when still fresh.
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

// pricePoint holds a single price history record used for analysis.
type pricePoint struct {
	price      float64
	recordedAt time.Time
}

// compute fetches price history and derives annotation events.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (PriceAnnotations, error) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT price, recorded_at
		FROM price_history
		WHERE item_id = $1
		ORDER BY recorded_at ASC`, itemID)
	if err != nil {
		return PriceAnnotations{}, err
	}
	defer rows.Close()

	var points []pricePoint
	for rows.Next() {
		var pp pricePoint
		if err := rows.Scan(&pp.price, &pp.recordedAt); err != nil {
			return PriceAnnotations{}, err
		}
		points = append(points, pp)
	}
	if err := rows.Err(); err != nil {
		return PriceAnnotations{}, err
	}

	annotations := analyze(points)

	return PriceAnnotations{
		ItemID:      itemID.String(),
		Annotations: annotations,
	}, nil
}

// analyze computes annotation events from an ordered price series.
// The series must already be ordered by recorded_at ASC.
func analyze(points []pricePoint) []Annotation {
	if len(points) == 0 {
		return []Annotation{}
	}

	var annotations []Annotation
	allTimeLow := points[0].price

	// stableStart tracks the first index of the current stable window.
	// A window is stable when all prices are within ±2% of the window's first price.
	stableStart := 0
	stableCount := 1

	for i := 1; i < len(points); i++ {
		prev := points[i-1].price
		curr := points[i].price

		if prev == 0 {
			// Avoid division by zero; reset stable window and continue.
			stableStart = i
			stableCount = 1
			continue
		}

		changePct := (curr - prev) / prev * 100

		// --- New all-time low ---
		if curr < allTimeLow {
			allTimeLow = curr
			annotations = append(annotations, Annotation{
				RecordedAt:  points[i].recordedAt.UTC(),
				Price:       curr,
				Type:        TypeNewLow,
				Label:       "Nouveau bas",
				Description: "Prix le plus bas observé",
			})
		}

		// --- Rebound after low: price rises ≥5% after a low ---
		if changePct >= 5 {
			// Check whether the previous point was at or near the all-time low (within 2%).
			prevRelLow := (prev - allTimeLow) / allTimeLow * 100
			if prevRelLow <= 2 {
				annotations = append(annotations, Annotation{
					RecordedAt:  points[i].recordedAt.UTC(),
					Price:       curr,
					Type:        TypeRebound,
					Label:       "Remontée après creux",
					Description: "Le prix remonte d'au moins 5% après un creux",
				})
			}
		}

		// --- Large single-step spike up ≥10% ---
		if changePct >= 10 {
			annotations = append(annotations, Annotation{
				RecordedAt:  points[i].recordedAt.UTC(),
				Price:       curr,
				Type:        TypeSpikeUp,
				Label:       "Forte hausse",
				Description: "Hausse de prix d'au moins 10% en une étape",
			})
		}

		// --- Large single-step spike down ≥10% ---
		if changePct <= -10 {
			annotations = append(annotations, Annotation{
				RecordedAt:  points[i].recordedAt.UTC(),
				Price:       curr,
				Type:        TypeSpikeDown,
				Label:       "Forte baisse",
				Description: "Baisse de prix d'au moins 10% en une étape",
			})
		}

		// --- Stabilisation: 5+ consecutive points within ±2% of the window anchor ---
		anchor := points[stableStart].price
		relDiff := (curr - anchor) / anchor * 100
		if relDiff >= -2 && relDiff <= 2 {
			stableCount++
			if stableCount == 5 {
				// Emit annotation at the 5th point of the stable window.
				annotations = append(annotations, Annotation{
					RecordedAt:  points[i].recordedAt.UTC(),
					Price:       curr,
					Type:        TypeStable,
					Label:       "Stabilisation",
					Description: "Le prix s'est stabilisé sur au moins 5 observations consécutives",
				})
			}
		} else {
			// Price moved outside ±2% — reset window to current point.
			stableStart = i
			stableCount = 1
		}
	}

	return annotations
}
