package collaborator

import (
	"context"
	"fmt"
)

// Service orchestrates collaborator business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new collaborator service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateInvite validates the role and persists a new invite record.
func (s *Service) CreateInvite(ctx context.Context, missionID, email string, role Role) (*Collaborator, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("collaborator: invalid role: %s", role)
	}
	return s.repo.Create(ctx, missionID, email, role)
}

// GetInviteInfo returns the collaborator record associated with the token.
func (s *Service) GetInviteInfo(ctx context.Context, token string) (*Collaborator, error) {
	return s.repo.GetByInviteToken(ctx, token)
}

// JoinMission marks a collaborator as joined using the invite token.
// displayName is required; returns an error if empty.
func (s *Service) JoinMission(ctx context.Context, token, displayName string) (*Collaborator, error) {
	if displayName == "" {
		return nil, fmt.Errorf("collaborator: display name is required")
	}
	return s.repo.MarkJoined(ctx, token, displayName)
}

// ListByMission returns all collaborators for a mission.
func (s *Service) ListByMission(ctx context.Context, missionID string) ([]Collaborator, error) {
	return s.repo.ListByMission(ctx, missionID)
}

// RemoveCollaborator deletes a collaborator record by ID.
func (s *Service) RemoveCollaborator(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
