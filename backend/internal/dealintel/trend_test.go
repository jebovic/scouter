package dealintel_test

import (
	"testing"
	"time"

	"github.com/jibei/scouter/internal/dealintel"
)

func TestComputeTrend(t *testing.T) {
	now := time.Now()
	price := func(p float64, hoursAgo float64) dealintel.PricePoint {
		return dealintel.PricePoint{Price: p, RecordedAt: now.Add(-time.Duration(hoursAgo * float64(time.Hour)))}
	}

	tests := []struct {
		name   string
		points []dealintel.PricePoint
		want   dealintel.Trend
	}{
		{
			name:   "empty snapshots returns stable",
			points: nil,
			want:   dealintel.TrendStable,
		},
		{
			name:   "single point returns stable",
			points: []dealintel.PricePoint{price(100, 0)},
			want:   dealintel.TrendStable,
		},
		{
			name:   "two identical points returns stable",
			points: []dealintel.PricePoint{price(100, 24), price(100, 0)},
			want:   dealintel.TrendStable,
		},
		{
			name:   "dropping trend — 10% drop",
			points: []dealintel.PricePoint{price(100, 96), price(97, 72), price(94, 48), price(91, 24), price(90, 0)},
			want:   dealintel.TrendDropping,
		},
		{
			name:   "rising trend — 10% rise",
			points: []dealintel.PricePoint{price(90, 96), price(93, 72), price(96, 48), price(99, 24), price(100, 0)},
			want:   dealintel.TrendRising,
		},
		{
			name:   "flat with small variation returns stable",
			points: []dealintel.PricePoint{price(100, 96), price(101, 72), price(99, 48), price(100, 24), price(100, 0)},
			want:   dealintel.TrendStable,
		},
		{
			name:   "exactly at dropping threshold (3% drop)",
			points: []dealintel.PricePoint{price(100, 48), price(99, 24), price(97, 0)},
			want:   dealintel.TrendDropping,
		},
		{
			name:   "exactly at rising threshold (3% rise)",
			points: []dealintel.PricePoint{price(97, 48), price(99, 24), price(100, 0)},
			want:   dealintel.TrendRising,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dealintel.ComputeTrend(tc.points)
			if got != tc.want {
				t.Errorf("ComputeTrend() = %q, want %q", got, tc.want)
			}
		})
	}
}
