package mission

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// ErrNotFound is returned when a mission cannot be located by slug or ID.
var ErrNotFound = errors.New("not found")

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

func (s *Service) ListPaged(ctx context.Context, cursor *time.Time, limit int, includeArchived bool) ([]Mission, error) {
	missions, err := s.repo.ListPaged(ctx, cursor, limit, includeArchived)
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

// GenerateShareToken creates a shareable token for a mission. Idempotent — returns
// the existing token if the mission already has one.
func (s *Service) GenerateShareToken(ctx context.Context, id uuid.UUID) (*Mission, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	if m.ShareToken != nil {
		return m, nil
	}
	token, err := GenerateShareToken()
	if err != nil {
		return nil, err
	}
	return s.repo.SetShareToken(ctx, id, token)
}

// RevokeShareToken removes the share token from a mission.
func (s *Service) RevokeShareToken(ctx context.Context, id uuid.UUID) error {
	return s.repo.ClearShareToken(ctx, id)
}

// GetByShareToken returns the mission with the given share token, or nil if not found.
func (s *Service) GetByShareToken(ctx context.Context, token string) (*Mission, error) {
	return s.repo.GetByShareToken(ctx, token)
}

// Archive marks a mission as archived (hidden from dashboard by default).
func (s *Service) Archive(ctx context.Context, id uuid.UUID) (*Mission, error) {
	return s.repo.Archive(ctx, id)
}

// Unarchive removes the archived status from a mission.
func (s *Service) Unarchive(ctx context.Context, id uuid.UUID) (*Mission, error) {
	return s.repo.Unarchive(ctx, id)
}

// Duplicate creates a copy of the mission identified by slug. The copy gets a
// " (copie)" name suffix, phase reset to "researching", and lessons/shareToken/
// archivedAt cleared. All other fields are copied from the original.
func (s *Service) Duplicate(ctx context.Context, slug string) (*Mission, error) {
	orig, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("duplicate mission: get original: %w", err)
	}
	if orig == nil {
		return nil, fmt.Errorf("mission %q: %w", slug, ErrNotFound)
	}

	newName := orig.Name + " (copie)"
	newSlug, err := s.uniqueSlug(ctx, newName)
	if err != nil {
		return nil, fmt.Errorf("duplicate mission: generate slug: %w", err)
	}

	copy := Mission{
		Slug:           newSlug,
		Name:           newName,
		Icon:           orig.Icon,
		Category:       orig.Category,
		Budget:         orig.Budget,
		Currency:       orig.Currency,
		Locale:         orig.Locale,
		Phase:          "researching",
		Constraints:    orig.Constraints,
		CostCategories: orig.CostCategories,
		Timeline:       orig.Timeline,
		WeightProfile:  orig.WeightProfile,
		EnvelopeID:     orig.EnvelopeID,
		// Lessons, ShareToken, ArchivedAt intentionally omitted (nil)
	}
	if copy.Constraints == nil {
		copy.Constraints = []Constraint{}
	}
	if copy.CostCategories == nil {
		copy.CostCategories = []string{}
	}

	return s.repo.Create(ctx, copy)
}

// CloneMission creates a full deep copy of a mission (options + shopping items)
// identified by slug. The repository handles the transaction and slug generation.
func (s *Service) CloneMission(ctx context.Context, slug string) (Mission, error) {
	return s.repo.CloneMission(ctx, slug)
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
