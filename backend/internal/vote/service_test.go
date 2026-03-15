package vote_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jibei/scouter/internal/vote"
)

func TestService_RecordVote_InvalidValue(t *testing.T) {
	svc := vote.NewService(nil) // repo won't be called on validation error
	err := svc.RecordVote(nil, "opt-1", "collab-1", 2) //nolint:staticcheck
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be -1, 0, or 1")
}

func TestService_RecordVote_InvalidNegative(t *testing.T) {
	svc := vote.NewService(nil)
	err := svc.RecordVote(nil, "opt-1", "collab-1", -2) //nolint:staticcheck
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be -1, 0, or 1")
}
