package pricefloor

import (
	"fmt"
	"sort"
	"time"
)

// predict computes a PriceFloor from a slice of historical prices (ascending by time)
// and the current price of the item.
func predict(itemID string, currentPrice float64, prices []float64) PriceFloor {
	result := PriceFloor{
		ItemID:       itemID,
		CurrentPrice: currentPrice,
		GeneratedAt:  time.Now().UTC(),
	}

	n := len(prices)
	if n == 0 {
		result.FloorConfidence = "low"
		result.PredictedFloor = currentPrice
		result.WaitReason = "Pas assez de données historiques pour prédire un plancher. Continuez à surveiller les prix."
		return result
	}

	// Historical min across all recorded prices.
	historicalMin := prices[0]
	for _, p := range prices[1:] {
		if p < historicalMin {
			historicalMin = p
		}
	}
	result.HistoricalMin = historicalMin

	// Determine confidence and predicted floor.
	var predictedFloor float64
	switch {
	case n >= 15:
		result.FloorConfidence = "high"
		// Use the 5th-percentile of the sorted price distribution.
		predictedFloor = percentile(prices, 5)
	case n >= 5:
		result.FloorConfidence = "medium"
		// Use 5th percentile but tempered with historical min.
		predictedFloor = percentile(prices, 5)
	default:
		// < 5 data points — low confidence, use historical min as-is.
		result.FloorConfidence = "low"
		predictedFloor = historicalMin
	}

	// For medium/high confidence: assume floor could go 3% below 5th percentile.
	if result.FloorConfidence != "low" {
		predictedFloor = predictedFloor * 0.97
	}

	result.PredictedFloor = predictedFloor

	// Distance from current price to predicted floor.
	result.DistanceToFloor = currentPrice - predictedFloor
	if currentPrice > 0 {
		result.DistancePct = (result.DistanceToFloor / currentPrice) * 100
	}

	// Decision: wait if current price is more than 5% above floor.
	result.ShouldWait = result.DistancePct > 5.0

	// French explanations.
	result.WaitReason, result.BuyNowReason = buildExplanations(result)

	return result
}

// percentile returns the p-th percentile (0–100) of a slice of float64.
// The input slice is copied and sorted internally.
func percentile(data []float64, p int) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

// buildExplanations returns French wait/buy-now reasons based on the PriceFloor state.
func buildExplanations(pf PriceFloor) (waitReason, buyNowReason string) {
	confLabel := map[string]string{
		"low":    "faible",
		"medium": "moyenne",
		"high":   "haute",
	}[pf.FloorConfidence]

	if pf.ShouldWait {
		waitReason = fmt.Sprintf(
			"Le prix actuel (%.2f €) est %.1f %% au-dessus du plancher estimé (%.2f €). "+
				"Avec une confiance %s, nous recommandons d'attendre une baisse avant d'acheter.",
			pf.CurrentPrice, pf.DistancePct, pf.PredictedFloor, confLabel,
		)
		return waitReason, ""
	}

	// Near the floor.
	buyNowReason = fmt.Sprintf(
		"Le prix actuel (%.2f €) est proche du plancher estimé (%.2f €) — seulement %.1f %% au-dessus. "+
			"Avec une confiance %s, c'est probablement un bon moment pour acheter !",
		pf.CurrentPrice, pf.PredictedFloor, pf.DistancePct, confLabel,
	)
	return "", buyNowReason
}
