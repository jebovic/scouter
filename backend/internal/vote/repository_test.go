package vote_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jibei/scouter/internal/vote"
)

func TestVote_ValidRange(t *testing.T) {
	validVotes := []int{-1, 0, 1}
	for _, v := range validVotes {
		assert.True(t, v >= -1 && v <= 1, "vote %d should be valid", v)
	}
}

func TestVote_InvalidValues(t *testing.T) {
	invalidVotes := []int{-2, 2, 100, -100}
	for _, v := range invalidVotes {
		assert.False(t, v >= -1 && v <= 1, "vote %d should be invalid", v)
	}
}

func TestAggregatedVotes_Score(t *testing.T) {
	av := &vote.AggregatedVotes{
		OptionID:  "test-option-id",
		Upvotes:   3,
		Downvotes: 1,
		Score:     2,
	}
	assert.Equal(t, 2, av.Score)
	assert.Equal(t, 3, av.Upvotes)
	assert.Equal(t, 1, av.Downvotes)
	assert.Equal(t, "test-option-id", av.OptionID)
}

func TestAggregatedVotes_ByCollaborator(t *testing.T) {
	av := &vote.AggregatedVotes{
		OptionID:  "opt-1",
		Upvotes:   2,
		Downvotes: 0,
		Score:     2,
		ByCollaborator: map[string]int{
			"collab-a": 1,
			"collab-b": 1,
		},
	}
	assert.Equal(t, 1, av.ByCollaborator["collab-a"])
	assert.Equal(t, 1, av.ByCollaborator["collab-b"])
	assert.Equal(t, 2, len(av.ByCollaborator))
}

func TestRecordVoteRequest_Fields(t *testing.T) {
	req := vote.RecordVoteRequest{
		CollaboratorID: "collab-1",
		Vote:           1,
	}
	assert.Equal(t, "collab-1", req.CollaboratorID)
	assert.Equal(t, 1, req.Vote)
}
