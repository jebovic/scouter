package purchase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MissionPhaseUpdater is the subset of mission.Repository used by the purchase service.
type MissionPhaseUpdater interface {
	SetPhase(ctx context.Context, missionID uuid.UUID, phase string) error
}

// Service handles business logic for purchase records.
type Service interface {
	Record(ctx context.Context, missionID uuid.UUID, req CreateRequest) (*PurchaseRecord, error)
	Get(ctx context.Context, missionID uuid.UUID) (*PurchaseRecord, error)
	Update(ctx context.Context, missionID uuid.UUID, req UpdateRequest) (*PurchaseRecord, error)
}

type service struct {
	repo     Repository
	missions MissionPhaseUpdater
}

// NewService creates a purchase service.
func NewService(repo Repository, missions MissionPhaseUpdater) Service {
	return &service{repo: repo, missions: missions}
}

func (s *service) Record(ctx context.Context, missionID uuid.UUID, req CreateRequest) (*PurchaseRecord, error) {
	exists, err := s.repo.ExistsByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("checking existing purchase: %w", err)
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	purchasedAt, err := time.Parse("2006-01-02", req.PurchasedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid purchasedAt date: %w", err)
	}

	rec := PurchaseRecord{
		ID:           uuid.New(),
		MissionID:    missionID,
		OptionID:     req.OptionID,
		PurchasedAt:  purchasedAt,
		FinalPrice:   req.FinalPrice,
		Merchant:     req.Merchant,
		Satisfaction: req.Satisfaction,
		Review:       req.Review,
	}

	result, err := s.repo.Create(ctx, rec)
	if err != nil {
		return nil, fmt.Errorf("creating purchase record: %w", err)
	}

	// Advance mission phase to done — non-fatal if it fails.
	if phaseErr := s.missions.SetPhase(ctx, missionID, "done"); phaseErr != nil {
		_ = phaseErr
	}

	return result, nil
}

func (s *service) Get(ctx context.Context, missionID uuid.UUID) (*PurchaseRecord, error) {
	return s.repo.GetByMission(ctx, missionID)
}

func (s *service) Update(ctx context.Context, missionID uuid.UUID, req UpdateRequest) (*PurchaseRecord, error) {
	return s.repo.Update(ctx, missionID, req)
}
