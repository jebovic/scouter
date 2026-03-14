// Package decision implements the Decision Engine: weighted scoring of options
// and a DecisionAgent that produces a natural-language recommendation summary.
package decision

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for service-level conditions the handler can map to HTTP status codes.
var (
	ErrMissionNotFound = errors.New("mission not found")
	ErrNoOptions       = errors.New("no options to score")
)

// ScoreResult holds the computed score and breakdown for a single option.
type ScoreResult struct {
	OptionID    uuid.UUID          `json:"optionId"`
	OptionName  string             `json:"optionName"`
	Score       float64            `json:"score"`      // 0-100 composite
	PriceScore  float64            `json:"priceScore"` // 0-100 price-fit component
	QualScore   float64            `json:"qualScore"`  // 0-100 attribute-quality component
	FeatScore   float64            `json:"featScore"`  // 0-100 feature-pass component
	Eliminated  bool               `json:"eliminated"` // true if any hard constraint is violated
	EliminReason string            `json:"eliminReason,omitempty"`
}

// Decision is the persisted result of a single run of the decision engine.
type Decision struct {
	ID        uuid.UUID     `json:"id"`
	MissionID uuid.UUID     `json:"missionId"`
	Scores    []ScoreResult `json:"scores"`
	Summary   string        `json:"summary"`
	CreatedAt time.Time     `json:"createdAt"`
}

// RunRequest carries mission + options into the Service.Run method.
// It is not persisted; it is constructed by the handler from existing data.
type RunRequest struct {
	MissionID uuid.UUID `json:"missionId"`
}
