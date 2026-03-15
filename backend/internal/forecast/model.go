// Package forecast implements the ForecastAgent, which uses LLM tool-use to
// generate a structured budget forecast for a mission. It owns prompt
// construction, tool schema definition, response parsing, and forecast
// persistence. The llm.Provider is transport-only.
package forecast

import (
	"errors"
	"time"
)

// ErrNotFound is returned by Repository.GetLatest when no forecast exists.
var ErrNotFound = errors.New("forecast not found")

// ForecastResult is the structured output of a budget forecast run.
type ForecastResult struct {
	ID              string    `json:"id"`
	MissionID       string    `json:"missionId"`
	EstimatedTotal  *float64  `json:"estimatedTotal"`
	Confidence      string    `json:"confidence"` // low | medium | high
	Recommendations []string  `json:"recommendations"`
	RiskFactors     []string  `json:"riskFactors"`
	MonthsToSave    *int      `json:"monthsToSave"`
	OptimalBuyTime  *string   `json:"optimalBuyTime"`
	CreatedAt       time.Time `json:"createdAt"`
}
