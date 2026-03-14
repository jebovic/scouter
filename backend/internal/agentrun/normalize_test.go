package agentrun

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MacBook Pro™ 14", "macbook pro 14"},
		{"MacBook Pro® (2024)", "macbook pro"},
		{"  Dell XPS  15  ", "dell xps 15"},
		{"Sony WH-1000XM5©", "sony wh-1000xm5"},
		{"ThinkPad X1 Carbon (Gen 11)", "thinkpad x1 carbon"},
		{"macbook pro 14", "macbook pro 14"},
	}
	for _, tt := range tests {
		got := normalizeName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeName(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestComputeDiff_AllNew(t *testing.T) {
	prev := []SnapshotItem{}
	curr := []SnapshotItem{
		{ID: "1", Name: "MacBook Pro"},
		{ID: "2", Name: "Dell XPS"},
	}
	d := ComputeDiff(prev, curr)
	if len(d.Added) != 2 {
		t.Errorf("Added = %d; want 2", len(d.Added))
	}
	if len(d.Removed) != 0 {
		t.Errorf("Removed = %d; want 0", len(d.Removed))
	}
}

func TestComputeDiff_AllRemoved(t *testing.T) {
	prev := []SnapshotItem{
		{ID: "1", Name: "MacBook Pro"},
	}
	curr := []SnapshotItem{}
	d := ComputeDiff(prev, curr)
	if len(d.Removed) != 1 {
		t.Errorf("Removed = %d; want 1", len(d.Removed))
	}
	if len(d.Added) != 0 {
		t.Errorf("Added = %d; want 0", len(d.Added))
	}
}

func TestComputeDiff_Unchanged(t *testing.T) {
	items := []SnapshotItem{
		{ID: "1", Name: "MacBook Pro"},
		{ID: "2", Name: "Dell XPS"},
	}
	d := ComputeDiff(items, items)
	if len(d.Added) != 0 {
		t.Errorf("Added = %d; want 0", len(d.Added))
	}
	if len(d.Removed) != 0 {
		t.Errorf("Removed = %d; want 0", len(d.Removed))
	}
}

func TestComputeDiff_NormalizedNameMatch(t *testing.T) {
	// "MacBook Pro™ (2024)" should match "macbook pro" — both normalize to same
	prev := []SnapshotItem{{ID: "1", Name: "MacBook Pro™ (2024)"}}
	curr := []SnapshotItem{{ID: "2", Name: "macbook pro"}}
	d := ComputeDiff(prev, curr)
	if len(d.Added) != 0 {
		t.Errorf("Added = %d; want 0 (names should match via normalization)", len(d.Added))
	}
	if len(d.Removed) != 0 {
		t.Errorf("Removed = %d; want 0 (names should match via normalization)", len(d.Removed))
	}
}

func TestComputeDiff_MixedChanges(t *testing.T) {
	prev := []SnapshotItem{
		{ID: "1", Name: "MacBook Pro"},
		{ID: "2", Name: "Dell XPS"},
	}
	curr := []SnapshotItem{
		{ID: "1", Name: "MacBook Pro"},
		{ID: "3", Name: "ThinkPad X1"},
	}
	d := ComputeDiff(prev, curr)
	if len(d.Added) != 1 || d.Added[0].Name != "ThinkPad X1" {
		t.Errorf("Added = %v; want [{ThinkPad X1}]", d.Added)
	}
	if len(d.Removed) != 1 || d.Removed[0].Name != "Dell XPS" {
		t.Errorf("Removed = %v; want [{Dell XPS}]", d.Removed)
	}
}
