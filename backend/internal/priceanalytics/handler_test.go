package priceanalytics

import (
	"math"
	"testing"
	"time"
)

func TestComputeEmpty(t *testing.T) {
	stats := compute("abc", []priceRow{})
	if stats.DataPoints != 0 {
		t.Fatalf("expected 0 data points, got %d", stats.DataPoints)
	}
}

func TestComputeSinglePoint(t *testing.T) {
	now := time.Now()
	stats := compute("id1", []priceRow{{price: 100.0, recordedAt: now}})
	if stats.Min != 100.0 {
		t.Errorf("min: want 100, got %f", stats.Min)
	}
	if stats.Max != 100.0 {
		t.Errorf("max: want 100, got %f", stats.Max)
	}
	if stats.Average != 100.0 {
		t.Errorf("avg: want 100, got %f", stats.Average)
	}
	if stats.Volatility != 0 {
		t.Errorf("volatility: want 0, got %f", stats.Volatility)
	}
	if stats.DataPoints != 1 {
		t.Errorf("data points: want 1, got %d", stats.DataPoints)
	}
}

func TestComputeMultiPoint(t *testing.T) {
	base := time.Now().Add(-30 * 24 * time.Hour)
	pts := []priceRow{
		{price: 100.0, recordedAt: base},
		{price: 120.0, recordedAt: base.Add(10 * 24 * time.Hour)},
		{price: 80.0, recordedAt: base.Add(20 * 24 * time.Hour)},
	}
	stats := compute("id2", pts)

	if stats.Min != 80.0 {
		t.Errorf("min: want 80, got %f", stats.Min)
	}
	if stats.Max != 120.0 {
		t.Errorf("max: want 120, got %f", stats.Max)
	}
	wantAvg := 100.0
	if math.Abs(stats.Average-wantAvg) > 0.001 {
		t.Errorf("avg: want %f, got %f", wantAvg, stats.Average)
	}
	if stats.DataPoints != 3 {
		t.Errorf("data points: want 3, got %d", stats.DataPoints)
	}
	// Volatility should be positive
	if stats.Volatility <= 0 {
		t.Errorf("volatility: want > 0, got %f", stats.Volatility)
	}
}

func TestComputeTrendSlope(t *testing.T) {
	base := time.Now().Add(-10 * 24 * time.Hour)
	// Strictly rising prices
	pts := []priceRow{
		{price: 100.0, recordedAt: base},
		{price: 110.0, recordedAt: base.Add(5 * 24 * time.Hour)},
		{price: 120.0, recordedAt: base.Add(10 * 24 * time.Hour)},
	}
	stats := compute("id3", pts)
	if stats.TrendSlope <= 0 {
		t.Errorf("expected positive slope for rising prices, got %f", stats.TrendSlope)
	}

	// Strictly falling prices
	pts2 := []priceRow{
		{price: 120.0, recordedAt: base},
		{price: 110.0, recordedAt: base.Add(5 * 24 * time.Hour)},
		{price: 100.0, recordedAt: base.Add(10 * 24 * time.Hour)},
	}
	stats2 := compute("id4", pts2)
	if stats2.TrendSlope >= 0 {
		t.Errorf("expected negative slope for falling prices, got %f", stats2.TrendSlope)
	}
}

func TestComputePriceDrop7d(t *testing.T) {
	now := time.Now().UTC()
	// Baseline: 10 days ago at 200, current: today at 150 → drop of -50
	pts := []priceRow{
		{price: 200.0, recordedAt: now.Add(-10 * 24 * time.Hour)},
		{price: 150.0, recordedAt: now},
	}
	stats := compute("id5", pts)
	wantDrop := 150.0 - 200.0 // -50
	if math.Abs(stats.PriceDrop7d-wantDrop) > 0.001 {
		t.Errorf("priceDrop7d: want %f, got %f", wantDrop, stats.PriceDrop7d)
	}
}

func TestComputePriceDrop7dNoBaseline(t *testing.T) {
	now := time.Now().UTC()
	// All points within last 7 days → no baseline → drop7d = 0
	pts := []priceRow{
		{price: 100.0, recordedAt: now.Add(-3 * 24 * time.Hour)},
		{price: 90.0, recordedAt: now.Add(-1 * 24 * time.Hour)},
	}
	stats := compute("id6", pts)
	if stats.PriceDrop7d != 0 {
		t.Errorf("expected 0 drop when no baseline, got %f", stats.PriceDrop7d)
	}
}
