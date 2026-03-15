package envelope

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repo is the storage interface consumed by Service.
// Both *Repository (real DB) and mockRepository (tests) satisfy it.
type Repo interface {
	FindAll(ctx context.Context) ([]Envelope, error)
	Create(ctx context.Context, e Envelope) (Envelope, error)
	Update(ctx context.Context, e Envelope) (Envelope, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error)
}

// Service is a thin validation layer over the Repo.
type Service struct {
	repo Repo
}

// NewService constructs a Service.
func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

// ListAll returns all envelopes without filtering.
func (s *Service) ListAll(ctx context.Context) ([]Envelope, error) {
	return s.repo.FindAll(ctx)
}

// CreateEnvelope validates the request and persists a new envelope.
func (s *Service) CreateEnvelope(ctx context.Context, req CreateRequest) (Envelope, error) {
	if req.Name == "" {
		return Envelope{}, ErrNameRequired
	}
	if req.MonthlyAmount <= 0 {
		return Envelope{}, ErrInvalidAmount
	}
	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}
	return s.repo.Create(ctx, Envelope{
		Name:          req.Name,
		MonthlyAmount: req.MonthlyAmount,
		Currency:      currency,
	})
}

// UpdateEnvelope applies a partial update to an existing envelope.
func (s *Service) UpdateEnvelope(ctx context.Context, id uuid.UUID, req UpdateRequest) (Envelope, error) {
	if req.Name != nil && *req.Name == "" {
		return Envelope{}, ErrNameRequired
	}
	if req.MonthlyAmount != nil && *req.MonthlyAmount <= 0 {
		return Envelope{}, ErrInvalidAmount
	}

	// Load current state so we can merge partial fields.
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return Envelope{}, err
	}
	var target Envelope
	found := false
	for _, e := range all {
		if e.ID == id {
			target = e
			found = true
			break
		}
	}
	if !found {
		return Envelope{}, pgx.ErrNoRows
	}

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.MonthlyAmount != nil {
		target.MonthlyAmount = *req.MonthlyAmount
	}
	if req.Currency != nil {
		target.Currency = *req.Currency
	}
	return s.repo.Update(ctx, target)
}

// GetSummary returns the envelope along with computed mission stats.
func (s *Service) GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error) {
	return s.repo.GetSummary(ctx, id)
}

// DeleteEnvelope removes an envelope by ID.
func (s *Service) DeleteEnvelope(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}
	return err
}
