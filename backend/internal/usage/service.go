package usage

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
)

// Service provides business-level usage operations.
type Service struct {
	repo Repository
}

// NewService creates a new usage Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Log persists one LLM call. agent is "research" or "pricing".
// missionID may be uuid.Nil if not applicable.
// wasFallback should be true when the secondary provider was used.
func (s *Service) Log(ctx context.Context, u llm.Usage, agent string, missionID uuid.UUID, wasFallback bool) {
	rec := Record{
		Provider:     u.Provider,
		Model:        u.Model,
		Agent:        agent,
		MissionID:    uuidPtr(missionID),
		InputTokens:  u.InputTokens,
		OutputTokens: u.OutputTokens,
		WasFallback:  wasFallback,
	}
	// Best-effort: logging failures must not affect the main call path.
	if err := s.repo.Insert(ctx, rec); err != nil {
		slog.WarnContext(ctx, "usage insert failed", "err", err, "agent", agent, "provider", u.Provider)
	}
}

// GetSummary returns aggregated token usage for the given period.
func (s *Service) GetSummary(ctx context.Context, period string) (Summary, error) {
	since := sinceFor(period)
	summary, err := s.repo.Summarize(ctx, since)
	if err != nil {
		return Summary{}, err
	}
	summary.Period = period
	return summary, nil
}

func sinceFor(period string) time.Time {
	now := time.Now().UTC()
	switch period {
	case "24h":
		return now.Add(-24 * time.Hour)
	case "30d":
		return now.AddDate(0, 0, -30)
	default: // "7d"
		return now.AddDate(0, 0, -7)
	}
}
