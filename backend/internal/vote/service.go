package vote

import (
	"context"
	"fmt"
)

// Service orchestrates vote business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new vote service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// RecordVote validates and persists a collaborator's vote on an option.
// vote must be -1 (downvote), 0 (neutral/retract), or 1 (upvote).
func (s *Service) RecordVote(ctx context.Context, optionID, collaboratorID string, v int) error {
	if v < -1 || v > 1 {
		return fmt.Errorf("vote: value must be -1, 0, or 1, got %d", v)
	}
	return s.repo.Record(ctx, optionID, collaboratorID, v)
}

// GetAggregated returns aggregated vote counts for an option.
func (s *Service) GetAggregated(ctx context.Context, optionID string) (*AggregatedVotes, error) {
	return s.repo.GetAggregated(ctx, optionID)
}
