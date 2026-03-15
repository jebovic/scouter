package comparison

import "time"

// Weight holds a per-mission criterion weight for the comparison matrix.
type Weight struct {
	ID        string    `json:"id"`
	MissionID string    `json:"missionId"`
	Criteria  string    `json:"criteria"`
	Weight    int       `json:"weight"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SetWeightRequest is the body payload for upsert.
type SetWeightRequest struct {
	Criteria string `json:"criteria"`
	Weight   int    `json:"weight"`
}

// OptionScore is a ranked option with its computed weighted score.
type OptionScore struct {
	OptionID   string  `json:"optionId"`
	OptionName string  `json:"optionName"`
	Score      float64 `json:"score"` // 0–100
	Rank       int     `json:"rank"`
}

// MatrixResponse is the full comparison matrix result.
type MatrixResponse struct {
	Weights []Weight      `json:"weights"`
	Scores  []OptionScore `json:"scores"`
}

// rawOption is an internal representation used during score computation.
type rawOption struct {
	id    string
	name  string
	attrs map[string]interface{}
}
