// Package dealintel provides pure-Go price trend and deal score computation.
// It has no I/O dependencies and is designed for easy unit testing.
package dealintel

import "time"

// Trend indicates the price direction of an item over its recent history.
type Trend string

const (
	TrendDropping Trend = "dropping"
	TrendStable   Trend = "stable"
	TrendRising   Trend = "rising"
)

// PricePoint is a single price observation at a point in time.
type PricePoint struct {
	Price      float64
	RecordedAt time.Time
}

// trendThreshold is the fractional change from first to last snapshot
// required to classify a trend as dropping or rising.
// 3% change threshold.
const trendThreshold = 0.03

// ComputeTrend classifies the price direction from a series of snapshots.
// Returns TrendStable when fewer than 2 data points are available.
// Uses the change from the oldest to the newest snapshot.
// Snapshots must be ordered oldest-first by RecordedAt.
func ComputeTrend(snapshots []PricePoint) Trend {
	if len(snapshots) < 2 {
		return TrendStable
	}

	oldest := snapshots[0].Price
	newest := snapshots[len(snapshots)-1].Price

	if oldest == 0 {
		return TrendStable
	}

	change := (newest - oldest) / oldest

	switch {
	case change <= -trendThreshold:
		return TrendDropping
	case change >= trendThreshold:
		return TrendRising
	default:
		return TrendStable
	}
}
