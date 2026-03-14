package dealintel_test

import (
	"math"
	"testing"
	"time"

	"github.com/jibei/scouter/internal/dealintel"
)

func float64Ptr(v float64) *float64 { return &v }

func TestComputeDealScore(t *testing.T) {
	now := time.Now()
	snap := func(p float64, hoursAgo float64) dealintel.PricePoint {
		return dealintel.PricePoint{Price: p, RecordedAt: now.Add(-time.Duration(hoursAgo * float64(time.Hour)))}
	}

	tests := []struct {
		name          string
		current       float64
		target        *float64
		points        []dealintel.PricePoint
		wantNil       bool   // expect nil DealScore (insufficient data)
		wantPositive  bool   // score > 0 (below average)
		wantNegative  bool   // score < 0 (above average)
		wantTargetHit bool   // pctBelowTarget >= 0
	}{
		{
			name:    "no snapshots returns nil",
			current: 100,
			target:  nil,
			points:  nil,
			wantNil: true,
		},
		{
			name:    "one snapshot returns nil",
			current: 100,
			target:  nil,
			points:  []dealintel.PricePoint{snap(100, 24)},
			wantNil: true,
		},
		{
			name:    "two snapshots returns nil",
			current: 100,
			target:  nil,
			points:  []dealintel.PricePoint{snap(100, 48), snap(100, 24)},
			wantNil: true,
		},
		{
			name:         "price below average gives positive score",
			current:      80,
			target:       nil,
			points:       []dealintel.PricePoint{snap(100, 72), snap(100, 48), snap(100, 24)},
			wantPositive: true,
		},
		{
			name:         "price above average gives negative score",
			current:      120,
			target:       nil,
			points:       []dealintel.PricePoint{snap(100, 72), snap(100, 48), snap(100, 24)},
			wantNegative: true,
		},
		{
			name:         "price at average gives ~0 score",
			current:      100,
			target:       nil,
			points:       []dealintel.PricePoint{snap(100, 72), snap(100, 48), snap(100, 24)},
			wantPositive: false,
			wantNegative: false,
		},
		{
			name:          "target hit when price below target",
			current:       90,
			target:        float64Ptr(100),
			points:        []dealintel.PricePoint{snap(100, 72), snap(100, 48), snap(100, 24)},
			wantPositive:  true,
			wantTargetHit: true,
		},
		{
			name:          "target not hit when price above target",
			current:       110,
			target:        float64Ptr(100),
			points:        []dealintel.PricePoint{snap(100, 72), snap(100, 48), snap(100, 24)},
			wantNegative:  true,
			wantTargetHit: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dealintel.ComputeDealScore(tc.current, tc.target, tc.points)

			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil DealScore, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("unexpected nil DealScore")
			}

			if tc.wantPositive && got.Score <= 0 {
				t.Errorf("expected positive score, got %.2f", got.Score)
			}
			if tc.wantNegative && got.Score >= 0 {
				t.Errorf("expected negative score, got %.2f", got.Score)
			}
			if !tc.wantPositive && !tc.wantNegative {
				if math.Abs(got.Score) > 1e-9 {
					t.Errorf("expected ~0 score, got %.2f", got.Score)
				}
			}
			if tc.wantTargetHit && got.PctBelowTarget < 0 {
				t.Errorf("expected target hit (pctBelowTarget >= 0), got %.2f", got.PctBelowTarget)
			}
			if !tc.wantTargetHit && tc.target != nil && got.PctBelowTarget >= 0 {
				t.Errorf("expected target not hit (pctBelowTarget < 0), got %.2f", got.PctBelowTarget)
			}
		})
	}
}
