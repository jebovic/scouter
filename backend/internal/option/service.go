package option

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles option business logic.
type Service struct {
	repo Repository
}

// NewService creates a new option service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByMission(ctx context.Context, missionID uuid.UUID) ([]Option, error) {
	opts, err := s.repo.ListByMission(ctx, missionID)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = []Option{}
	}
	return opts, nil
}

func (s *Service) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Option, error) {
	opts, err := s.repo.ListByMissionPaged(ctx, missionID, cursor, limit)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = []Option{}
	}
	return opts, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Option, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, missionID uuid.UUID, req CreateRequest) (*Option, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	badge := req.Badge
	if badge == "" {
		badge = "watch"
	}

	o := Option{
		MissionID:  missionID,
		Name:       req.Name,
		Category:   req.Category,
		Badge:      badge,
		Attributes: req.Attributes,
		PriceRange: req.PriceRange,
		Notes:      req.Notes,
		Warnings:   req.Warnings,
		URL:        req.URL,
	}
	if o.Attributes == nil {
		o.Attributes = []Attribute{}
	}
	if o.Warnings == nil {
		o.Warnings = []string{}
	}

	return s.repo.Create(ctx, o)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Option, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Pin(ctx context.Context, id uuid.UUID) (*Option, error) {
	return s.repo.Pin(ctx, id)
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID, req RejectRequest) (*Option, error) {
	return s.repo.Reject(ctx, id, req)
}

func (s *Service) Unreject(ctx context.Context, id uuid.UUID) (*Option, error) {
	return s.repo.Unreject(ctx, id)
}

func (s *Service) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	return s.repo.DeletePinned(ctx, missionID)
}
