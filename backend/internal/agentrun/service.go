package agentrun

import (
	"context"

	"github.com/google/uuid"
)

// Service wraps Repository to maintain the Handler→Service→Repository pattern.
type Service struct {
	repo Repository
}

// NewService creates a new agent run service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListByMission returns agent runs for a mission, optionally filtered by agent type.
func (s *Service) ListByMission(ctx context.Context, missionID uuid.UUID, agentType *AgentType) ([]Run, error) {
	return s.repo.ListByMission(ctx, missionID, agentType)
}
