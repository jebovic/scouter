package dealintel

// minSnapshotsForScore is the minimum number of snapshots needed to compute
// a meaningful deal score. With fewer than this, we return nil.
const minSnapshotsForScore = 3

// DealScore holds computed deal intelligence for a shopping item.
type DealScore struct {
	// PctBelowAvg is (historicalAvg - currentPrice) / historicalAvg * 100.
	// Positive means the current price is below average (good deal).
	PctBelowAvg float64 `json:"pctBelowAvg"`

	// PctBelowTarget is (targetPrice - currentPrice) / targetPrice * 100.
	// Positive means the current price is below the target (target hit).
	// Zero when no target is set.
	PctBelowTarget float64 `json:"pctBelowTarget"`

	// HistoricalAvg is the mean of the snapshot prices.
	HistoricalAvg float64 `json:"historicalAvg"`

	// Trend is the computed price direction.
	Trend Trend `json:"trend"`

	// SnapshotCount is the number of price history points used.
	SnapshotCount int `json:"snapshotCount"`
}

// ComputeDealScore computes deal intelligence for an item given its current price,
// optional target price, and historical price snapshots.
// Returns nil when there are fewer than minSnapshotsForScore data points.
// Snapshots must be ordered oldest-first by RecordedAt (as returned by ListPriceHistory).
func ComputeDealScore(currentPrice float64, targetPrice *float64, snapshots []PricePoint) *DealScore {
	if len(snapshots) < minSnapshotsForScore {
		return nil
	}

	// Compute historical average from all snapshots.
	var sum float64
	for _, s := range snapshots {
		sum += s.Price
	}
	avg := sum / float64(len(snapshots))

	pctBelowAvg := 0.0
	if avg != 0 {
		pctBelowAvg = (avg - currentPrice) / avg * 100
	}

	pctBelowTarget := 0.0
	if targetPrice != nil && *targetPrice != 0 {
		pctBelowTarget = (*targetPrice - currentPrice) / *targetPrice * 100
	}

	return &DealScore{
		PctBelowAvg:    pctBelowAvg,
		PctBelowTarget: pctBelowTarget,
		HistoricalAvg:  avg,
		Trend:          ComputeTrend(snapshots),
		SnapshotCount:  len(snapshots),
	}
}
