package vote

import "time"

// Vote records a single collaborator's vote on an option.
type Vote struct {
	ID             string    `json:"id"`
	OptionID       string    `json:"optionId"`
	CollaboratorID string    `json:"collaboratorId"`
	Vote           int       `json:"vote"` // -1, 0, or 1
	VotedAt        time.Time `json:"votedAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// RecordVoteRequest is the payload for submitting or updating a vote.
type RecordVoteRequest struct {
	CollaboratorID string `json:"collaboratorId"`
	Vote           int    `json:"vote"` // -1, 0, or 1
}

// AggregatedVotes summarises all votes for a single option.
type AggregatedVotes struct {
	OptionID       string         `json:"optionId"`
	Upvotes        int            `json:"upvotes"`
	Downvotes      int            `json:"downvotes"`
	Score          int            `json:"score"` // upvotes - downvotes
	ByCollaborator map[string]int `json:"byCollaborator"` // collaborator_id → vote value
}
