package comment

import "time"

// Comment represents a single comment on a mission (optionally attached to an option).
type Comment struct {
	ID        string    `json:"id"`
	MissionID string    `json:"missionId"`
	OptionID  *string   `json:"optionId,omitempty"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateCommentRequest is the payload for creating a new comment.
type CreateCommentRequest struct {
	OptionID *string `json:"optionId,omitempty" validate:"omitempty"`
	Author   string  `json:"author"             validate:"required,min=1,max=50"`
	Body     string  `json:"body"               validate:"required,min=1,max=1000"`
}
