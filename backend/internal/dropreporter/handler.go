package dropreporter

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

type cacheEntry struct {
	prediction DropPrediction
	expiresAt  time.Time
}

// Handler exposes the drop-prediction endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetPrediction handles GET /api/missions/{missionId}/items/{itemId}/drop-prediction.
func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()

	// Check 1h cache.
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.prediction)
			return
		}
	}

	prediction, err := h.compute(r, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(key, cacheEntry{
		prediction: prediction,
		expiresAt:  time.Now().Add(time.Hour),
	})

	httputil.WriteJSON(w, http.StatusOK, prediction)
}

// priceRecord holds a single price history entry.
type priceRecord struct {
	price      float64
	recordedAt time.Time
}

// compute fetches price history and derives a DropPrediction.
func (h *Handler) compute(r *http.Request, itemID uuid.UUID) (DropPrediction, error) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT price, recorded_at
		FROM price_history
		WHERE item_id = $1
		ORDER BY recorded_at DESC
		LIMIT 30`, itemID)
	if err != nil {
		return DropPrediction{}, err
	}
	defer rows.Close()

	var records []priceRecord
	for rows.Next() {
		var rec priceRecord
		if err := rows.Scan(&rec.price, &rec.recordedAt); err != nil {
			return DropPrediction{}, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return DropPrediction{}, err
	}

	return calculate(itemID.String(), records), nil
}

// calculate derives a DropPrediction from raw records using pure math — no DB deps.
func calculate(itemID string, records []priceRecord) DropPrediction {
	noData := DropPrediction{
		ItemID:           itemID,
		Probability:      0,
		PredictedDropPct: 0,
		Horizon:          30,
		Confidence:       "faible",
		Signals:          []string{"Pas assez de données"},
		GeneratedAt:      time.Now().UTC(),
	}

	n := len(records)
	if n == 0 {
		return noData
	}

	// Records come DESC from DB; reverse to chronological order for regression.
	chron := make([]priceRecord, n)
	for i, rec := range records {
		chron[n-1-i] = rec
	}

	// --- Linear regression (least-squares slope in price/day) ---
	// Use ordinal time index to avoid float precision issues with Unix timestamps.
	// x[i] = days elapsed from first record.
	t0 := chron[0].recordedAt
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, rec := range chron {
		xs[i] = rec.recordedAt.Sub(t0).Hours() / 24.0
		ys[i] = rec.price
	}

	slope, _ := linearRegression(xs, ys)

	// --- Historical amplitude (max - min) ---
	minPrice, maxPrice := ys[0], ys[0]
	for _, p := range ys {
		if p < minPrice {
			minPrice = p
		}
		if p > maxPrice {
			maxPrice = p
		}
	}
	amplitude := maxPrice - minPrice
	avgPrice := mean(ys)

	// --- Days since last drop ---
	daysSinceLastDrop := -1
	for i := 1; i < n; i++ {
		if chron[i].price < chron[i-1].price {
			daysSinceLastDrop = int(time.Since(chron[i].recordedAt).Hours() / 24)
			break
		}
	}

	// --- Velocity: absolute slope per day ---
	velocity := math.Abs(slope)

	// --- Probability calculation ---
	// Three components, each 0–1, weighted and clamped.

	// 1. Slope negativity: negative slope → higher probability.
	//    Map slope to [-inf, 0] → [0,1]. Use tanh for smooth saturation.
	slopeComponent := 0.0
	if slope < 0 {
		// Normalise by 1% of avgPrice per day as reference rate.
		ref := avgPrice * 0.01
		if ref <= 0 {
			ref = 1.0
		}
		slopeComponent = math.Tanh(math.Abs(slope) / ref)
	}

	// 2. Recency of last drop: more recent → higher probability.
	recencyComponent := 0.0
	if daysSinceLastDrop >= 0 {
		// Decays to ~0 at 30 days.
		recencyComponent = math.Exp(-float64(daysSinceLastDrop) / 10.0)
	}

	// 3. Volatility: higher amplitude relative to avg → more likely to drop.
	volatilityComponent := 0.0
	if avgPrice > 0 {
		volatilityComponent = math.Min(1.0, (amplitude/avgPrice)*3.0)
	}

	// Weighted sum.
	prob := 0.5*slopeComponent + 0.3*recencyComponent + 0.2*volatilityComponent
	prob = math.Max(0.0, math.Min(1.0, prob))

	// Low-confidence floor: not enough data.
	confidence := "faible"
	horizon := 30
	if n >= 15 {
		confidence = "élevé"
		horizon = 7
	} else if n >= 5 {
		confidence = "moyen"
		horizon = 14
	}

	// --- Predicted drop percent: historical amplitude / 2 as % of avgPrice ---
	predictedDropPct := 0.0
	if avgPrice > 0 {
		predictedDropPct = math.Round((amplitude/2/avgPrice)*1000) / 10 // round to 1 decimal
	}

	// --- Signals (French) ---
	signals := buildSignals(slope, daysSinceLastDrop, velocity, avgPrice, n, confidence)

	return DropPrediction{
		ItemID:           itemID,
		Probability:      math.Round(prob*1000) / 1000, // 3 decimal places
		PredictedDropPct: predictedDropPct,
		Horizon:          horizon,
		Confidence:       confidence,
		Signals:          signals,
		GeneratedAt:      time.Now().UTC(),
	}
}

// linearRegression computes the least-squares slope and intercept for (x, y) pairs.
func linearRegression(xs, ys []float64) (slope, intercept float64) {
	n := float64(len(xs))
	if n < 2 {
		return 0, 0
	}
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return slope, intercept
}

// mean returns the arithmetic mean of a slice.
func mean(vs []float64) float64 {
	if len(vs) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range vs {
		s += v
	}
	return s / float64(len(vs))
}

// buildSignals returns French-language explanation strings for the prediction.
func buildSignals(slope float64, daysSinceLastDrop int, velocity, avgPrice float64, n int, confidence string) []string {
	var signals []string

	if n < 5 {
		signals = append(signals, "Données insuffisantes pour une prédiction fiable")
		return signals
	}

	if slope < 0 {
		signals = append(signals, "Tendance baissière détectée")
	} else if slope > 0 {
		signals = append(signals, "Tendance haussière — baisse peu probable à court terme")
	} else {
		signals = append(signals, "Prix stable, pas de tendance marquée")
	}

	if daysSinceLastDrop >= 0 {
		if daysSinceLastDrop == 0 {
			signals = append(signals, "Dernière baisse aujourd'hui")
		} else if daysSinceLastDrop == 1 {
			signals = append(signals, "Dernière baisse il y a 1 jour")
		} else {
			signals = append(signals, "Dernière baisse il y a "+itoa(daysSinceLastDrop)+" jours")
		}
	} else {
		signals = append(signals, "Aucune baisse de prix observée dans l'historique")
	}

	// Velocity signal.
	if avgPrice > 0 {
		velocityPct := velocity / avgPrice * 100
		switch {
		case velocityPct >= 1.0:
			signals = append(signals, "Forte vélocité de variation des prix")
		case velocityPct >= 0.3:
			signals = append(signals, "Vélocité de variation modérée")
		default:
			signals = append(signals, "Faible vélocité de variation des prix")
		}
	}

	// Confidence signal.
	switch confidence {
	case "élevé":
		signals = append(signals, "Historique riche ("+itoa(n)+" relevés) — haute confiance")
	case "moyen":
		signals = append(signals, "Historique partiel ("+itoa(n)+" relevés) — confiance moyenne")
	default:
		signals = append(signals, "Historique limité ("+itoa(n)+" relevés) — confiance faible")
	}

	return signals
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
