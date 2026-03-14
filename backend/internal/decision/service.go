package decision

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// Service orchestrates scoring + LLM summarization and persists the result.
type Service struct {
	repo       Repository
	decAgent   *Agent
	missionRepo mission.Repository
	optionRepo  option.Repository
}

// NewService creates a new decision Service.
func NewService(
	repo Repository,
	decAgent *Agent,
	missionRepo mission.Repository,
	optionRepo option.Repository,
) *Service {
	return &Service{
		repo:        repo,
		decAgent:    decAgent,
		missionRepo: missionRepo,
		optionRepo:  optionRepo,
	}
}

// Run fetches mission + options, scores them, calls the LLM, and persists the Decision.
func (s *Service) Run(ctx context.Context, missionID uuid.UUID) (*Decision, error) {
	m, err := s.missionRepo.GetByID(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("get mission: %w", err)
	}
	if m == nil {
		return nil, ErrMissionNotFound
	}

	opts, err := s.optionRepo.ListByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	if len(opts) == 0 {
		return nil, ErrNoOptions
	}

	scores := Score(*m, opts)

	summary, err := s.decAgent.Summarize(ctx, *m, scores)
	if err != nil {
		return nil, fmt.Errorf("summarize: %w", err)
	}

	d := Decision{
		MissionID: missionID,
		Scores:    scores,
		Summary:   summary,
	}

	created, err := s.repo.Create(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("persist decision: %w", err)
	}

	return created, nil
}

// GetLatest returns the most recent decision for a mission, or nil if none exists.
func (s *Service) GetLatest(ctx context.Context, missionID uuid.UUID) (*Decision, error) {
	d, err := s.repo.GetLatestByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("get latest decision: %w", err)
	}
	return d, nil
}
