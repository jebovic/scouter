package prediction_test

import (
	"math"
	"testing"

	"github.com/jibei/scouter/internal/prediction"
)

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ── LinearRegression ────────────────────────────────────────────────────────

func TestLinearRegression_PerfectLine(t *testing.T) {
	// y = 2x + 10 → slope=2, intercept=10
	xs := []float64{0, 1, 2, 3, 4}
	ys := []float64{10, 12, 14, 16, 18}

	slope, intercept := prediction.LinearRegression(xs, ys)

	if !almostEqual(slope, 2.0, 1e-9) {
		t.Errorf("slope: want 2.0, got %f", slope)
	}
	if !almostEqual(intercept, 10.0, 1e-9) {
		t.Errorf("intercept: want 10.0, got %f", intercept)
	}
}

func TestLinearRegression_FlatLine(t *testing.T) {
	xs := []float64{0, 1, 2, 3}
	ys := []float64{5, 5, 5, 5}

	slope, intercept := prediction.LinearRegression(xs, ys)

	if !almostEqual(slope, 0.0, 1e-9) {
		t.Errorf("slope: want 0.0, got %f", slope)
	}
	if !almostEqual(intercept, 5.0, 1e-9) {
		t.Errorf("intercept: want 5.0, got %f", intercept)
	}
}

func TestLinearRegression_NegativeSlope(t *testing.T) {
	// y = -3x + 30
	xs := []float64{0, 1, 2, 3, 4}
	ys := []float64{30, 27, 24, 21, 18}

	slope, intercept := prediction.LinearRegression(xs, ys)

	if !almostEqual(slope, -3.0, 1e-9) {
		t.Errorf("slope: want -3.0, got %f", slope)
	}
	if !almostEqual(intercept, 30.0, 1e-9) {
		t.Errorf("intercept: want 30.0, got %f", intercept)
	}
}

func TestLinearRegression_SinglePoint(t *testing.T) {
	// Single point — regression falls back gracefully (slope=0)
	xs := []float64{0}
	ys := []float64{42}

	slope, intercept := prediction.LinearRegression(xs, ys)

	// With a single point we cannot determine slope; expect slope=0, intercept=y
	if !almostEqual(slope, 0.0, 1e-9) {
		t.Errorf("slope: want 0.0, got %f", slope)
	}
	if !almostEqual(intercept, 42.0, 1e-9) {
		t.Errorf("intercept: want 42.0, got %f", intercept)
	}
}

// ── Predict ─────────────────────────────────────────────────────────────────

func TestPredict_PerfectRisingLine(t *testing.T) {
	// 5 points: y = 1x + 100, xs=[0..4], N=4
	// PredictedPrice = slope*(4+30) + intercept = 1*34 + 100 = 134
	// ConfidenceLevel = "high" (>=5 points)
	// Trend = "rising" (slope > 0.5% of mean)
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 100},
		{DayIndex: 1, Price: 101},
		{DayIndex: 2, Price: 102},
		{DayIndex: 3, Price: 103},
		{DayIndex: 4, Price: 104},
	}
	result := prediction.Predict("item-1", 104, points)

	if result.ItemID != "item-1" {
		t.Errorf("itemId: want item-1, got %s", result.ItemID)
	}
	if !almostEqual(result.PredictedPrice, 134.0, 1e-6) {
		t.Errorf("predictedPrice: want 134.0, got %f", result.PredictedPrice)
	}
	if result.ConfidenceLevel != "high" {
		t.Errorf("confidenceLevel: want high, got %s", result.ConfidenceLevel)
	}
	if result.Trend != "rising" {
		t.Errorf("trend: want rising, got %s", result.Trend)
	}
	if result.ForecastDays != 30 {
		t.Errorf("forecastDays: want 30, got %d", result.ForecastDays)
	}
	if result.DataPoints != 5 {
		t.Errorf("dataPoints: want 5, got %d", result.DataPoints)
	}
}

func TestPredict_FallingPrice(t *testing.T) {
	// y = -2x + 200, 5 points
	// PredictedPrice = -2*(4+30) + 200 = -68 + 200 = 132 → positive, no clamp needed
	// Trend = "falling"
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 200},
		{DayIndex: 1, Price: 198},
		{DayIndex: 2, Price: 196},
		{DayIndex: 3, Price: 194},
		{DayIndex: 4, Price: 192},
	}
	result := prediction.Predict("item-2", 192, points)

	if result.Trend != "falling" {
		t.Errorf("trend: want falling, got %s", result.Trend)
	}
	if !almostEqual(result.PredictedPrice, 132.0, 1e-6) {
		t.Errorf("predictedPrice: want 132.0, got %f", result.PredictedPrice)
	}
}

func TestPredict_StablePrice(t *testing.T) {
	// flat line
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 50},
		{DayIndex: 1, Price: 50},
		{DayIndex: 2, Price: 50},
		{DayIndex: 3, Price: 50},
		{DayIndex: 4, Price: 50},
	}
	result := prediction.Predict("item-3", 50, points)

	if result.Trend != "stable" {
		t.Errorf("trend: want stable, got %s", result.Trend)
	}
	if !almostEqual(result.PredictedPrice, 50.0, 1e-6) {
		t.Errorf("predictedPrice: want 50.0, got %f", result.PredictedPrice)
	}
}

// ── Confidence level thresholds ─────────────────────────────────────────────

func TestConfidenceLevel_High(t *testing.T) {
	points := make([]prediction.PricePoint, 5)
	for i := range points {
		points[i] = prediction.PricePoint{DayIndex: float64(i), Price: 100}
	}
	result := prediction.Predict("x", 100, points)
	if result.ConfidenceLevel != "high" {
		t.Errorf("want high, got %s", result.ConfidenceLevel)
	}
}

func TestConfidenceLevel_Medium_ThreePoints(t *testing.T) {
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 100},
		{DayIndex: 1, Price: 101},
		{DayIndex: 2, Price: 102},
	}
	result := prediction.Predict("x", 102, points)
	if result.ConfidenceLevel != "medium" {
		t.Errorf("want medium, got %s", result.ConfidenceLevel)
	}
}

func TestConfidenceLevel_Medium_FourPoints(t *testing.T) {
	points := make([]prediction.PricePoint, 4)
	for i := range points {
		points[i] = prediction.PricePoint{DayIndex: float64(i), Price: 100}
	}
	result := prediction.Predict("x", 100, points)
	if result.ConfidenceLevel != "medium" {
		t.Errorf("want medium, got %s", result.ConfidenceLevel)
	}
}

func TestConfidenceLevel_Low_TwoPoints(t *testing.T) {
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 100},
		{DayIndex: 1, Price: 101},
	}
	result := prediction.Predict("x", 101, points)
	if result.ConfidenceLevel != "low" {
		t.Errorf("want low, got %s", result.ConfidenceLevel)
	}
}

func TestConfidenceLevel_Low_OnePoint(t *testing.T) {
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 100},
	}
	result := prediction.Predict("x", 100, points)
	if result.ConfidenceLevel != "low" {
		t.Errorf("want low, got %s", result.ConfidenceLevel)
	}
}

// ── TrendPct clamping ───────────────────────────────────────────────────────

func TestTrendPct_RisingClamped(t *testing.T) {
	// first=1, last=1000 → real pct = 99900% → clamp to 99
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 1},
		{DayIndex: 1, Price: 1000},
	}
	result := prediction.Predict("x", 1000, points)
	if result.TrendPct > 99 {
		t.Errorf("trendPct should be clamped to 99, got %f", result.TrendPct)
	}
}

func TestTrendPct_FallingClamped(t *testing.T) {
	// first=1000, last=1 → real pct = -99.9% → clamp to -99
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 1000},
		{DayIndex: 1, Price: 1},
	}
	result := prediction.Predict("x", 1, points)
	if result.TrendPct < -99 {
		t.Errorf("trendPct should be clamped to -99, got %f", result.TrendPct)
	}
}

func TestTrendPct_Normal(t *testing.T) {
	// first=100, last=110 → pct = 10%
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 100},
		{DayIndex: 1, Price: 110},
	}
	result := prediction.Predict("x", 110, points)
	if !almostEqual(result.TrendPct, 10.0, 1e-6) {
		t.Errorf("trendPct: want 10.0, got %f", result.TrendPct)
	}
}

// ── PredictedPrice clamped to > 0 ───────────────────────────────────────────

func TestPredictedPrice_ClampedAboveZero(t *testing.T) {
	// Very steep negative slope that would produce negative predicted price
	// slope ~ -100, intercept ~ 0
	// PredictedPrice = slope*(N+30)+intercept would be very negative → clamp to small positive
	points := []prediction.PricePoint{
		{DayIndex: 0, Price: 1000},
		{DayIndex: 1, Price: 900},
		{DayIndex: 2, Price: 800},
		{DayIndex: 3, Price: 700},
		{DayIndex: 4, Price: 600},
	}
	result := prediction.Predict("x", 600, points)
	if result.PredictedPrice <= 0 {
		t.Errorf("predictedPrice must be > 0, got %f", result.PredictedPrice)
	}
}
