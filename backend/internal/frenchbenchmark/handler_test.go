package frenchbenchmark

import (
	"testing"
)

func TestBuildBenchmarkNoUserPrices(t *testing.T) {
	result := buildBenchmark("mission-abc", "Mission Test", nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.MissionID != "mission-abc" {
		t.Errorf("wrong missionId: %s", result.MissionID)
	}
	if result.MedianPrice <= 0 {
		t.Errorf("expected positive median, got %f", result.MedianPrice)
	}
	// No user prices: avgPrice should equal marketMedian
	if result.YourAvgPrice != result.MedianPrice {
		t.Errorf("expected avgPrice == medianPrice when no user prices, got %f vs %f",
			result.YourAvgPrice, result.MedianPrice)
	}
	if result.SampleSize < 10 || result.SampleSize > 100 {
		t.Errorf("sampleSize out of range: %d", result.SampleSize)
	}
}

func TestBuildBenchmarkGoodPrice(t *testing.T) {
	// Median is deterministic; set low prices to force bon_prix verdict.
	missionID := "test-mission-good"
	median := computeMarketMedian(missionID)
	lowPrice := median * 0.80 // 20% below median
	prices := []float64{lowPrice, lowPrice}

	result := buildBenchmark(missionID, "Good Deal Mission", prices)
	if result.VerdictCode != "bon_prix" {
		t.Errorf("expected bon_prix, got %s (avg=%f, median=%f)", result.VerdictCode, result.YourAvgPrice, result.MedianPrice)
	}
}

func TestBuildBenchmarkHighPrice(t *testing.T) {
	missionID := "test-mission-high"
	median := computeMarketMedian(missionID)
	highPrice := median * 1.30 // 30% above median
	prices := []float64{highPrice}

	result := buildBenchmark(missionID, "Expensive Mission", prices)
	if result.VerdictCode != "au_dessus_du_marche" {
		t.Errorf("expected au_dessus_du_marche, got %s (avg=%f, median=%f)", result.VerdictCode, result.YourAvgPrice, result.MedianPrice)
	}
}

func TestBuildBenchmarkAveragePrice(t *testing.T) {
	missionID := "test-mission-avg"
	median := computeMarketMedian(missionID)
	prices := []float64{median} // exactly at median

	result := buildBenchmark(missionID, "Average Mission", prices)
	if result.VerdictCode != "prix_moyen" {
		t.Errorf("expected prix_moyen, got %s", result.VerdictCode)
	}
}

func TestComputeMarketMedianDeterministic(t *testing.T) {
	id := "some-mission-id"
	m1 := computeMarketMedian(id)
	m2 := computeMarketMedian(id)
	if m1 != m2 {
		t.Errorf("market median not deterministic: %f vs %f", m1, m2)
	}
	if m1 < 50 || m1 > 1050 {
		t.Errorf("market median out of range: %f", m1)
	}
}

func TestPriceRange(t *testing.T) {
	result := buildBenchmark("range-test", "Range Test", nil)
	if result.PriceRange[0] >= result.PriceRange[1] {
		t.Errorf("expected priceRange[0] < priceRange[1], got %v", result.PriceRange)
	}
}
