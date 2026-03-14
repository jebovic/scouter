package shopping

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles shopping item business logic.
type Service struct {
	repo Repository
}

// NewService creates a new shopping service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByMission(ctx context.Context, missionID uuid.UUID) ([]Item, error) {
	items, err := s.repo.ListByMission(ctx, missionID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []Item{}
	}
	return items, nil
}

func (s *Service) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Item, error) {
	items, err := s.repo.ListByMissionPaged(ctx, missionID, cursor, limit)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []Item{}
	}
	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, missionID uuid.UUID, req CreateRequest) (*Item, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	status := req.Status
	if status == "" {
		status = "watch"
	}

	item := Item{
		MissionID:        missionID,
		Name:             req.Name,
		Merchant:         req.Merchant,
		CostCategory:     req.CostCategory,
		Price:            req.Price,
		OriginalEstimate: req.OriginalEstimate,
		Status:           status,
		Note:             req.Note,
		URL:              req.URL,
	}

	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Item, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AddPriceSnapshot(ctx context.Context, itemID uuid.UUID, req PriceSnapshotRequest) (*PriceSnapshot, error) {
	if req.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}
	return s.repo.AddPriceSnapshot(ctx, itemID, req)
}

func (s *Service) ListPriceHistory(ctx context.Context, itemID uuid.UUID) ([]PriceSnapshot, error) {
	snapshots, err := s.repo.ListPriceHistory(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if snapshots == nil {
		snapshots = []PriceSnapshot{}
	}
	return snapshots, nil
}

func (s *Service) Pin(ctx context.Context, id uuid.UUID) (*Item, error) {
	return s.repo.Pin(ctx, id)
}

func (s *Service) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	return s.repo.DeletePinned(ctx, missionID)
}
