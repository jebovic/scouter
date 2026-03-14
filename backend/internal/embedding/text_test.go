package embedding

import (
	"strings"
	"testing"

	"github.com/jibei/scouter/internal/option"
)

func TestBuildEmbedText_BasicFields(t *testing.T) {
	opt := option.Option{
		Name:     "Sony WH-1000XM5",
		Category: "Headphones",
		Notes:    "Best-in-class ANC.",
	}
	got := BuildEmbedText(opt)
	if !strings.Contains(got, "Sony WH-1000XM5") {
		t.Errorf("missing name: %q", got)
	}
	if !strings.Contains(got, "Headphones") {
		t.Errorf("missing category: %q", got)
	}
	if !strings.Contains(got, "Best-in-class ANC") {
		t.Errorf("missing notes: %q", got)
	}
}

func TestBuildEmbedText_Attributes(t *testing.T) {
	opt := option.Option{
		Name:     "Sony WH-1000XM5",
		Category: "Headphones",
		Attributes: []option.Attribute{
			{Key: "battery", Label: "Battery Life", Value: "30h", Type: "text"},
			{Key: "weight", Label: "Weight", Value: 250, Type: "text"},
		},
	}
	got := BuildEmbedText(opt)
	if !strings.Contains(got, "Battery Life") {
		t.Errorf("missing attribute label: %q", got)
	}
	if !strings.Contains(got, "30h") {
		t.Errorf("missing attribute value: %q", got)
	}
}

func TestBuildEmbedText_NoNotes(t *testing.T) {
	opt := option.Option{Name: "X", Category: "Y"}
	got := BuildEmbedText(opt)
	if strings.HasSuffix(got, " ") {
		t.Errorf("trailing space: %q", got)
	}
}

func TestFormatVector(t *testing.T) {
	tests := []struct {
		vec  []float64
		want string
	}{
		{[]float64{1, 2, 3}, "[1,2,3]"},
		{[]float64{0.1, 0.2}, "[0.1,0.2]"},
		{[]float64{}, "[]"},
	}
	for _, tc := range tests {
		got := formatVector(tc.vec)
		if got != tc.want {
			t.Errorf("formatVector(%v) = %q, want %q", tc.vec, got, tc.want)
		}
	}
}
