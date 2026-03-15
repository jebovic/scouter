package wishlistprioritizer

import (
	"testing"
	"time"
)

func TestBuildPrioritizedListEmpty(t *testing.T) {
	result := buildPrioritizedList(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.TopPick != nil {
		t.Errorf("expected nil topPick for empty list")
	}
}

func TestBuildPrioritizedListSingleItem(t *testing.T) {
	price := 80.0
	target := 100.0
	items := []rawItem{
		{
			id:               "abc123",
			name:             "Sony WH-1000XM5",
			lastCheckedPrice: &price,
			targetPrice:      &target,
			alertEnabled:     true,
			updatedAt:        time.Now(),
		},
	}
	result := buildPrioritizedList(items)
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", result.Items[0].Rank)
	}
	if result.TopPick == nil {
		t.Fatal("expected non-nil topPick")
	}
	if result.TopPick.ID != "abc123" {
		t.Errorf("expected topPick id abc123, got %s", result.TopPick.ID)
	}
	if result.Items[0].Score < 0 || result.Items[0].Score > 100 {
		t.Errorf("score out of range: %f", result.Items[0].Score)
	}
}

func TestBuildPrioritizedListMixedUrgency(t *testing.T) {
	highPrice := 200.0
	lowPrice := 50.0
	target := 100.0

	now := time.Now()
	old := now.Add(-40 * 24 * time.Hour)

	items := []rawItem{
		{
			id:               "low-urgency",
			name:             "Item Low",
			lastCheckedPrice: &highPrice,
			targetPrice:      &target,
			alertEnabled:     false,
			updatedAt:        old,
		},
		{
			id:               "high-urgency",
			name:             "Item High",
			lastCheckedPrice: &lowPrice,
			targetPrice:      &target,
			alertEnabled:     true,
			updatedAt:        now,
		},
	}

	result := buildPrioritizedList(items)
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}

	// Rank 1 should have higher score than rank 2
	if result.Items[0].Score < result.Items[1].Score {
		t.Errorf("items not sorted by score: %f < %f", result.Items[0].Score, result.Items[1].Score)
	}

	// Ranks assigned correctly
	if result.Items[0].Rank != 1 || result.Items[1].Rank != 2 {
		t.Errorf("ranks incorrect: %d, %d", result.Items[0].Rank, result.Items[1].Rank)
	}

	// Top pick is the highest scored item
	if result.TopPick == nil {
		t.Fatal("expected non-nil topPick")
	}
	if result.TopPick.ID != result.Items[0].ID {
		t.Errorf("topPick mismatch: %s vs %s", result.TopPick.ID, result.Items[0].ID)
	}
}

func TestComputeUrgencyAlertEnabled(t *testing.T) {
	it := rawItem{
		alertEnabled: true,
		updatedAt:    time.Now(),
	}
	score := computeUrgency(it)
	if score < 50 {
		t.Errorf("expected urgency >= 50 for alert-enabled, got %f", score)
	}
}

func TestComputeUrgencyDecay(t *testing.T) {
	old := time.Now().Add(-40 * 24 * time.Hour)
	it := rawItem{
		alertEnabled: false,
		updatedAt:    old,
	}
	score := computeUrgency(it)
	if score != 0 {
		t.Errorf("expected 0 urgency for stale non-alert item, got %f", score)
	}
}

func TestComputeTrendDeterministic(t *testing.T) {
	id := "test-uuid-123"
	s1 := computeTrend(id)
	s2 := computeTrend(id)
	if s1 != s2 {
		t.Errorf("trend not deterministic: %f vs %f", s1, s2)
	}
	if s1 < 0 || s1 > 100 {
		t.Errorf("trend out of range: %f", s1)
	}
}

func TestComputeBudgetFit(t *testing.T) {
	avg := 100.0
	low := 50.0
	high := 200.0

	fitLow := computeBudgetFit(&low, avg)
	fitHigh := computeBudgetFit(&high, avg)

	if fitLow <= fitHigh {
		t.Errorf("lower price should have higher budget fit: %f vs %f", fitLow, fitHigh)
	}

	// nil price
	fitNil := computeBudgetFit(nil, avg)
	if fitNil != 50 {
		t.Errorf("expected 50 for nil price, got %f", fitNil)
	}
}

func TestVerdictFor(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{80, "Priorité haute — acheter maintenant"},
		{60, "Priorité moyenne — surveiller"},
		{30, "Priorité basse — peut attendre"},
	}
	for _, tt := range tests {
		got := verdictFor(tt.score)
		if got != tt.expected {
			t.Errorf("verdictFor(%f) = %q, want %q", tt.score, got, tt.expected)
		}
	}
}
