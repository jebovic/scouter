package watchlist

import (
	"time"

	"github.com/google/uuid"
)

type WatchEntry struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	ItemID        uuid.UUID  `json:"itemId" db:"item_id"`
	ItemName      string     `json:"itemName" db:"item_name"`
	BaselinePrice float64    `json:"baselinePrice" db:"baseline_price"`
	ThresholdPct  float64    `json:"thresholdPct" db:"threshold_pct"`
	CurrentPrice  float64    `json:"currentPrice" db:"current_price"`
	TriggeredAt   *time.Time `json:"triggeredAt,omitempty" db:"triggered_at"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
}

type CreateRequest struct {
	ItemID        string  `json:"itemId"`
	ItemName      string  `json:"itemName"`
	BaselinePrice float64 `json:"baselinePrice"`
	ThresholdPct  float64 `json:"thresholdPct"`
}

type CheckResult struct {
	Triggered bool    `json:"triggered"`
	DropPct   float64 `json:"dropPct"`
	Message   string  `json:"message"`
}
