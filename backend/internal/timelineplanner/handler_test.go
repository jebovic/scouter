package timelineplanner

import (
	"testing"
)

func TestBuildTimelineEmpty(t *testing.T) {
	tl := buildTimeline("m1", "Test Mission", nil)
	if tl == nil {
		t.Fatal("expected non-nil timeline")
	}
	if tl.TotalItems != 0 {
		t.Errorf("expected 0 items, got %d", tl.TotalItems)
	}
	if len(tl.Buckets) != 4 {
		t.Errorf("expected 4 buckets, got %d", len(tl.Buckets))
	}
}

func TestBuildTimelineBucketCount(t *testing.T) {
	items := []shoppingItem{
		{id: "i1", name: "Item A", status: "pending"},
		{id: "i2", name: "Item B", status: "researching"},
		{id: "i3", name: "Item C", status: "watching"},
		{id: "i4", name: "Item D", status: "purchased"},
	}
	tl := buildTimeline("m2", "Full Mission", items)
	if tl.TotalItems != 4 {
		t.Errorf("expected 4 total items, got %d", tl.TotalItems)
	}
	// Each bucket should have exactly 1 item.
	for i, b := range tl.Buckets {
		if b.ItemCount != 1 {
			t.Errorf("bucket[%d] expected 1 item, got %d", i, b.ItemCount)
		}
	}
}

func TestBuildTimelineWeekLabels(t *testing.T) {
	tl := buildTimeline("m3", "Mission", nil)
	for i, b := range tl.Buckets {
		if b.Week != i+1 {
			t.Errorf("bucket[%d] week = %d, want %d", i, b.Week, i+1)
		}
		if b.Label == "" {
			t.Errorf("bucket[%d] has empty label", i)
		}
		if b.Action == "" {
			t.Errorf("bucket[%d] has empty action", i)
		}
		if b.PromoHint == "" {
			t.Errorf("bucket[%d] has empty promoHint", i)
		}
	}
}

func TestStatusToWeek(t *testing.T) {
	tests := []struct {
		status string
		week   int
	}{
		{"pending", 0},
		{"", 0},
		{"researching", 1},
		{"watching", 2},
		{"shortlisted", 2},
		{"purchased", 3},
		{"done", 3},
	}
	for _, tt := range tests {
		if got := statusToWeek(tt.status); got != tt.week {
			t.Errorf("statusToWeek(%q) = %d, want %d", tt.status, got, tt.week)
		}
	}
}

func TestPromoHintsDeterministic(t *testing.T) {
	id := "some-mission-id"
	h1 := promoHints(id)
	h2 := promoHints(id)
	for i := 0; i < 4; i++ {
		if h1[i] != h2[i] {
			t.Errorf("promoHints not deterministic at index %d: %q vs %q", i, h1[i], h2[i])
		}
		if h1[i] == "" {
			t.Errorf("promoHints[%d] is empty", i)
		}
	}
}

func TestBuildTimelineSummary(t *testing.T) {
	tl := buildTimeline("m4", "My Mission", []shoppingItem{
		{id: "x", name: "X", status: "pending"},
	})
	if tl.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestBuildTimelineItemsInBuckets(t *testing.T) {
	items := []shoppingItem{
		{id: "a", name: "Alpha", status: "purchased"},
		{id: "b", name: "Beta", status: "purchased"},
	}
	tl := buildTimeline("m5", "Mission", items)
	// Both purchased → week 4 (index 3).
	if tl.Buckets[3].ItemCount != 2 {
		t.Errorf("expected 2 items in week 4 bucket, got %d", tl.Buckets[3].ItemCount)
	}
	for i := 0; i < 3; i++ {
		if tl.Buckets[i].ItemCount != 0 {
			t.Errorf("bucket[%d] should be empty, got %d items", i, tl.Buckets[i].ItemCount)
		}
	}
}
