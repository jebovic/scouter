package pricefloor

import "time"

// PriceFloor holds the price floor prediction for a shopping item.
type PriceFloor struct {
	ItemID          string    `json:"itemId"`
	CurrentPrice    float64   `json:"currentPrice"`
	PredictedFloor  float64   `json:"predictedFloor"`
	FloorConfidence string    `json:"floorConfidence"` // "low", "medium", "high"
	HistoricalMin   float64   `json:"historicalMin"`
	DistanceToFloor float64   `json:"distanceToFloor"` // currentPrice - predictedFloor
	DistancePct     float64   `json:"distancePct"`     // % above floor
	ShouldWait      bool      `json:"shouldWait"`
	WaitReason      string    `json:"waitReason"`
	BuyNowReason    string    `json:"buyNowReason"`
	GeneratedAt     time.Time `json:"generatedAt"`
}
