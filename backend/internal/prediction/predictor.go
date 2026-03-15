package prediction

import "math"

const (
	forecastHorizon = 30
	// slopePctThreshold is the fraction of mean price that slope must exceed
	// for a trend to be classified as rising or falling (0.5%).
	slopePctThreshold = 0.005
)

// LinearRegression computes the least-squares linear regression of ys onto xs,
// returning (slope, intercept). When fewer than 2 points are provided, slope is
// 0 and intercept is the mean of ys (or 0 if ys is empty).
func LinearRegression(xs, ys []float64) (slope, intercept float64) {
	n := len(xs)
	if n == 0 {
		return 0, 0
	}
	if n == 1 {
		return 0, ys[0]
	}

	// Compute means.
	var sumX, sumY float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Compute slope = Σ(xi-mx)(yi-my) / Σ(xi-mx)^2
	var num, den float64
	for i := range xs {
		dx := xs[i] - meanX
		num += dx * (ys[i] - meanY)
		den += dx * dx
	}

	if den == 0 {
		// All x values are the same — cannot determine slope.
		return 0, meanY
	}

	slope = num / den
	intercept = meanY - slope*meanX
	return slope, intercept
}

// Predict produces a PricePrediction from a slice of PricePoints.
// Points must be ordered by DayIndex (ascending). currentPrice is the item's
// live price stored in the DB (may differ from the last history snapshot).
func Predict(itemID string, currentPrice float64, points []PricePoint) PricePrediction {
	n := len(points)

	// Build xs/ys slices.
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i, p := range points {
		xs[i] = p.DayIndex
		ys[i] = p.Price
	}

	slope, intercept := LinearRegression(xs, ys)

	// Forecast: next x = last day index + 30 (or just 30 when n==0).
	var lastX float64
	if n > 0 {
		lastX = xs[n-1]
	}
	predicted := slope*(lastX+float64(forecastHorizon)) + intercept
	// Clamp to a small positive value so we never predict a non-positive price.
	if predicted <= 0 {
		predicted = 0.01
	}

	// Confidence level.
	var confidence string
	switch {
	case n >= 5:
		confidence = "high"
	case n >= 3:
		confidence = "medium"
	default:
		confidence = "low"
	}

	// Compute mean price for slope-based trend threshold.
	var meanPrice float64
	if n > 0 {
		var sum float64
		for _, y := range ys {
			sum += y
		}
		meanPrice = sum / float64(n)
	}

	// Trend classification based on slope relative to mean price.
	var trend string
	if meanPrice == 0 {
		trend = "stable"
	} else {
		threshold := slopePctThreshold * meanPrice
		switch {
		case slope > threshold:
			trend = "rising"
		case slope < -threshold:
			trend = "falling"
		default:
			trend = "stable"
		}
	}

	// TrendPct: (last - first) / first * 100, clamped to ±99.
	var trendPct float64
	if n >= 2 && ys[0] != 0 {
		raw := (ys[n-1] - ys[0]) / ys[0] * 100
		trendPct = math.Max(-99, math.Min(99, raw))
	}

	return PricePrediction{
		ItemID:          itemID,
		CurrentPrice:    currentPrice,
		PredictedPrice:  predicted,
		ConfidenceLevel: confidence,
		Trend:           trend,
		TrendPct:        trendPct,
		DataPoints:      n,
		ForecastDays:    forecastHorizon,
	}
}
