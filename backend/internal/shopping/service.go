package shopping

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/dealintel"
)

// ValidationError is returned by the service for business-rule violations.
// Handlers map this to HTTP 400; all other errors map to HTTP 500.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }

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
		return nil, ValidationError("name is required")
	}
	if req.Price < 0 {
		return nil, ValidationError("price must be non-negative")
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
		TargetPrice:      req.TargetPrice,
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
		return nil, ValidationError("price must be non-negative")
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

// GetPriceHistory returns the last `limit` price observations for an item as charting points.
// The result is always a non-nil slice.
func (s *Service) GetPriceHistory(ctx context.Context, itemID uuid.UUID, limit int) ([]PriceHistoryPoint, error) {
	pts, err := s.repo.GetPriceHistory(ctx, itemID, limit)
	if err != nil {
		return nil, err
	}
	if pts == nil {
		pts = []PriceHistoryPoint{}
	}
	return pts, nil
}

func (s *Service) Pin(ctx context.Context, id uuid.UUID) (*Item, error) {
	return s.repo.Pin(ctx, id)
}

func (s *Service) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	return s.repo.DeletePinned(ctx, missionID)
}

// GetDealScore computes deal intelligence for a shopping item.
// Returns (nil, false, nil) when the item is not found.
// Returns (nil, true, nil) when there are fewer than 3 price history snapshots.
func (s *Service) GetDealScore(ctx context.Context, itemID uuid.UUID) (*dealintel.DealScore, bool, error) {
	item, err := s.repo.GetByID(ctx, itemID)
	if err != nil {
		return nil, false, err
	}
	if item == nil {
		return nil, false, nil
	}

	snapshots, err := s.repo.ListPriceHistory(ctx, itemID)
	if err != nil {
		return nil, false, fmt.Errorf("list price history: %w", err)
	}

	points := make([]dealintel.PricePoint, len(snapshots))
	for i, snap := range snapshots {
		points[i] = dealintel.PricePoint{Price: snap.Price, RecordedAt: snap.RecordedAt}
	}
	return dealintel.ComputeDealScore(item.Price, item.TargetPrice, points), true, nil
}
