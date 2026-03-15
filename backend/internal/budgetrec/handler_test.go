package budgetrec

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/shopping"
)

func ptr(v float64) *float64 { return &v }

func TestAnalyze_EmptyItems(t *testing.T) {
	got := analyze("mission-1", nil)
	if got.HealthScore != 100 {
		t.Errorf("expected HealthScore 100 for empty list, got %d", got.HealthScore)
	}
	if got.TotalBudget != 0 || got.TotalSpent != 0 {
		t.Errorf("expected zero totals for empty list")
	}
	if len(got.Recommendations) != 0 {
		t.Errorf("expected no recommendations for empty list")
	}
}

func TestAnalyze_OnBudget(t *testing.T) {
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "tech", Price: 100, OriginalEstimate: ptr(100)},
		{ID: uuid.New(), CostCategory: "tech", Price: 95, OriginalEstimate: ptr(100)},
	}
	got := analyze("m", items)
	if got.HealthScore != 100 {
		t.Errorf("expected HealthScore 100 when on budget, got %d", got.HealthScore)
	}
	if len(got.Recommendations) != 0 {
		t.Errorf("expected no recommendations when on budget, got %d", len(got.Recommendations))
	}
}

func TestAnalyze_Overspent_Medium(t *testing.T) {
	// 120 spent vs 100 budget → 20 % overrun → medium priority
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "food", Price: 120, OriginalEstimate: ptr(100)},
	}
	got := analyze("m", items)
	if len(got.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(got.Recommendations))
	}
	rec := got.Recommendations[0]
	if rec.Priority != PriorityMedium {
		t.Errorf("expected medium priority, got %s", rec.Priority)
	}
	if rec.SaveAmount != 10 { // (120-100)*0.5
		t.Errorf("expected SaveAmount 10, got %f", rec.SaveAmount)
	}
	if got.HealthScore >= 100 {
		t.Errorf("expected HealthScore < 100 when overspent, got %d", got.HealthScore)
	}
}

func TestAnalyze_Overspent_High(t *testing.T) {
	// 130 spent vs 100 budget → 30 % overrun → high priority
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "food", Price: 130, OriginalEstimate: ptr(100)},
	}
	got := analyze("m", items)
	if len(got.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(got.Recommendations))
	}
	if got.Recommendations[0].Priority != PriorityHigh {
		t.Errorf("expected high priority, got %s", got.Recommendations[0].Priority)
	}
}

func TestAnalyze_Underspent(t *testing.T) {
	// 50 spent vs 100 budget → 50 % used → underspent
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "transport", Price: 50, OriginalEstimate: ptr(100)},
	}
	got := analyze("m", items)
	if len(got.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation for underspent category, got %d", len(got.Recommendations))
	}
	rec := got.Recommendations[0]
	if rec.Priority != PriorityLow {
		t.Errorf("expected low priority, got %s", rec.Priority)
	}
	if got.HealthScore != 100 {
		t.Errorf("expected HealthScore 100 when underspent, got %d", got.HealthScore)
	}
}

func TestAnalyze_NoOriginalEstimate_UsesPrice(t *testing.T) {
	// Without OriginalEstimate both budget and spent equal Price → on budget.
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "misc", Price: 200},
	}
	got := analyze("m", items)
	if got.TotalBudget != 200 || got.TotalSpent != 200 {
		t.Errorf("expected budget=200 spent=200, got budget=%f spent=%f", got.TotalBudget, got.TotalSpent)
	}
	if len(got.Recommendations) != 0 {
		t.Errorf("expected no recommendations when budget == spent")
	}
}

func TestAnalyze_HealthScoreCappedAt20(t *testing.T) {
	// 300 spent vs 100 budget → 200 % overrun → penalty capped at 80 → score 20
	items := []shopping.Item{
		{ID: uuid.New(), CostCategory: "luxury", Price: 300, OriginalEstimate: ptr(100)},
	}
	got := analyze("m", items)
	if got.HealthScore != 20 {
		t.Errorf("expected HealthScore 20 (penalty capped at 80), got %d", got.HealthScore)
	}
}

func TestAnalyze_MissionIDPassedThrough(t *testing.T) {
	got := analyze("test-mission-id", nil)
	if got.MissionID != "test-mission-id" {
		t.Errorf("expected MissionID to be passed through, got %s", got.MissionID)
	}
}
