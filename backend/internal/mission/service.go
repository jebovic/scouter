package mission

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// Service handles mission business logic.
type Service struct {
	repo Repository
}

// NewService creates a new mission service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Mission, error) {
	missions, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if missions == nil {
		missions = []Mission{}
	}
	return missions, nil
}

func (s *Service) ListPaged(ctx context.Context, cursor *time.Time, limit int) ([]Mission, error) {
	missions, err := s.repo.ListPaged(ctx, cursor, limit)
	if err != nil {
		return nil, err
	}
	if missions == nil {
		missions = []Mission{}
	}
	return missions, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Mission, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Mission, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Mission, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Budget <= 0 {
		return nil, fmt.Errorf("budget must be positive")
	}

	slug, err := s.uniqueSlug(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	icon := req.Icon
	if icon == "" {
		icon = "target"
	}
	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}
	locale := req.Locale
	if locale == "" {
		locale = "fr-FR"
	}

	m := Mission{
		Slug:           slug,
		Name:           req.Name,
		Icon:           icon,
		Category:       req.Category,
		Budget:         req.Budget,
		Currency:       currency,
		Locale:         locale,
		Phase:          "researching",
		Constraints:    req.Constraints,
		CostCategories: req.CostCategories,
	}
	if m.Constraints == nil {
		m.Constraints = []Constraint{}
	}
	if m.CostCategories == nil {
		m.CostCategories = []string{}
	}

	return s.repo.Create(ctx, m)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Mission, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// uniqueSlug generates a URL-safe slug from the name, appending a counter if needed.
// It caps retries at 100 to prevent unbounded loops.
func (s *Service) uniqueSlug(ctx context.Context, name string) (string, error) {
	base := nonAlphanumeric.ReplaceAllString(strings.ToLower(name), "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "mission"
	}

	slug := base
	for i := 2; i <= 101; i++ {
		existing, err := s.repo.GetBySlug(ctx, slug)
		if err != nil {
			return "", fmt.Errorf("check slug uniqueness: %w", err)
		}
		if existing == nil {
			return slug, nil
		}
		if i > 100 {
			return "", fmt.Errorf("could not generate unique slug for %q after 100 attempts", name)
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	// unreachable, but satisfies the compiler
	return "", fmt.Errorf("could not generate unique slug for %q", name)
}
