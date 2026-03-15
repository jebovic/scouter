package purchase

import (
	"encoding/json"
	"testing"
)

func TestMonthlyStatPoint_JSONMarshaling(t *testing.T) {
	p := MonthlyStatPoint{
		Month:         "2026-01",
		TotalSpent:    1234.56,
		TotalBudget:   2000.00,
		PurchaseCount: 3,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got MonthlyStatPoint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Month != p.Month {
		t.Errorf("Month: got %q, want %q", got.Month, p.Month)
	}
	if got.TotalSpent != p.TotalSpent {
		t.Errorf("TotalSpent: got %f, want %f", got.TotalSpent, p.TotalSpent)
	}
	if got.TotalBudget != p.TotalBudget {
		t.Errorf("TotalBudget: got %f, want %f", got.TotalBudget, p.TotalBudget)
	}
	if got.PurchaseCount != p.PurchaseCount {
		t.Errorf("PurchaseCount: got %d, want %d", got.PurchaseCount, p.PurchaseCount)
	}
}

func TestMonthlyStatPoint_EmptySliceJSON(t *testing.T) {
	points := []MonthlyStatPoint{}
	data, err := json.Marshal(points)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// An empty slice must marshal to "[]", not "null".
	if string(data) != "[]" {
		t.Errorf("empty slice: got %q, want %q", string(data), "[]")
	}
}

func TestMonthlyStatPoint_JSONFieldNames(t *testing.T) {
	p := MonthlyStatPoint{
		Month:         "2025-12",
		TotalSpent:    99.99,
		TotalBudget:   500.00,
		PurchaseCount: 1,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal to map failed: %v", err)
	}

	for _, key := range []string{"month", "totalSpent", "totalBudget", "purchaseCount"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}
