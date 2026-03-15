package quantityoptimizer

import (
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestBuildReportAllTiers(t *testing.T) {
	report := buildReport("item-abc", "Samsung TV", 1000.0, nil)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Tiers) != len(defaultQuantities) {
		t.Errorf("expected %d tiers, got %d", len(defaultQuantities), len(report.Tiers))
	}
	for i, tier := range report.Tiers {
		if tier.Quantity != defaultQuantities[i] {
			t.Errorf("tier[%d] quantity = %d, want %d", i, tier.Quantity, defaultQuantities[i])
		}
		if tier.UnitPrice <= 0 {
			t.Errorf("tier[%d] unitPrice <= 0", i)
		}
		if tier.TotalPrice <= 0 {
			t.Errorf("tier[%d] totalPrice <= 0", i)
		}
	}
}

func TestBuildReportBudgetFit(t *testing.T) {
	budget := 1200.0 // enough for 1 item at 1000, not for 2
	report := buildReport("item-budget", "Laptop", 1000.0, &budget)

	tier1 := report.Tiers[0] // qty=1
	if !tier1.BudgetFit {
		t.Errorf("qty=1 should fit budget of %.0f", budget)
	}
	// The higher quantity tiers with total > budget should not fit
	lastTier := report.Tiers[len(report.Tiers)-1] // qty=10
	if lastTier.BudgetFit {
		t.Errorf("qty=10 total=%.2f should not fit budget=%.0f", lastTier.TotalPrice, budget)
	}
}

func TestBuildReportOptimalQty(t *testing.T) {
	report := buildReport("opt-item", "Test Item", 100.0, nil)
	// Optimal qty should be in default quantities
	found := false
	for _, q := range defaultQuantities {
		if report.OptimalQty == q {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("optimalQty %d not in default quantities", report.OptimalQty)
	}
}

func TestBuildReportFallbackPrice(t *testing.T) {
	report := buildReport("no-price-item", "Unknown Item", 0, nil)
	if report.BasePrice <= 0 {
		t.Errorf("expected positive fallback price, got %f", report.BasePrice)
	}
}

func TestComputeDiscountDeterministic(t *testing.T) {
	id := "test-item-id"
	d1 := computeDiscount(id, 3)
	d2 := computeDiscount(id, 3)
	if d1 != d2 {
		t.Errorf("discount not deterministic: %f vs %f", d1, d2)
	}
}

func TestComputeDiscountIncreasing(t *testing.T) {
	id := "discount-test"
	// Discount for qty=10 should be >= discount for qty=1 (due to bonus).
	d1 := computeDiscount(id, 1)
	d10 := computeDiscount(id, 10)
	if d10 < d1 {
		t.Errorf("expected higher discount for qty=10 (%f) than qty=1 (%f)", d10, d1)
	}
}

func TestComputeDiscountCap(t *testing.T) {
	// Any discount should be capped at 25%.
	for _, qty := range defaultQuantities {
		d := computeDiscount("any-item", qty)
		if d > 25 {
			t.Errorf("discount exceeds 25%% for qty=%d: %f", qty, d)
		}
	}
}

func TestTierTotalPrice(t *testing.T) {
	report := buildReport("total-test", "Item", 100.0, nil)
	for _, tier := range report.Tiers {
		expected := tier.UnitPrice * float64(tier.Quantity)
		// Allow rounding tolerance of 0.02.
		diff := expected - tier.TotalPrice
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.02 {
			t.Errorf("qty=%d: totalPrice=%.2f, expected ~%.2f", tier.Quantity, tier.TotalPrice, expected)
		}
	}
}
