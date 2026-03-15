package mission

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handlerMockRepo extends mockRepo with duplicate-specific control.
type handlerMockRepo struct {
	mockRepo
	createErr error
}

func newHandlerMockRepo() *handlerMockRepo {
	return &handlerMockRepo{
		mockRepo: mockRepo{
			missions: make(map[string]*Mission),
			byID:     make(map[uuid.UUID]*Mission),
		},
	}
}

func (h *handlerMockRepo) Create(ctx context.Context, m Mission) (*Mission, error) {
	if h.createErr != nil {
		return nil, h.createErr
	}
	return h.mockRepo.Create(ctx, m)
}

// routerWithDuplicate wires POST /api/missions/{slug}/duplicate on a fresh chi router.
func routerWithDuplicate(svc *Service) chi.Router {
	r := chi.NewRouter()
	handler := NewHandler(svc)
	r.Post("/api/missions/{slug}/duplicate", handler.Duplicate)
	return r
}

func seedMission(repo *handlerMockRepo, name, slug, phase string) *Mission {
	id := uuid.New()
	m := &Mission{
		ID:             id,
		Slug:           slug,
		Name:           name,
		Icon:           "target",
		Category:       "electronics",
		Budget:         1000,
		Currency:       "EUR",
		Locale:         "fr-FR",
		Phase:          phase,
		Constraints:    []Constraint{{Key: "k", Label: "l", Value: "v", Type: "hard"}},
		CostCategories: []string{"cat1"},
		Timeline:       []TimelineEvent{{Date: "2026-01-01", Label: "ev", Urgent: false}},
		WeightProfile:  WeightProfile{Price: 1, Quality: 1, Feature: 1},
		Lessons:        ptr("some lesson"),
		ShareToken:     ptr("tok"),
		ArchivedAt:     ptrTime(time.Now()),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	repo.missions[slug] = m
	repo.byID[id] = m
	return m
}

func ptr(s string) *string { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// TestHandler_Duplicate_Success verifies 201 response with correct field resets.
func TestHandler_Duplicate_Success(t *testing.T) {
	repo := newHandlerMockRepo()
	svc := NewService(repo)
	seedMission(repo, "Gaming PC", "gaming-pc", "buying")

	r := routerWithDuplicate(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/duplicate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201\nbody: %s", w.Code, w.Body.String())
	}

	var m Mission
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Name must be "Gaming PC (copie)"
	wantName := "Gaming PC (copie)"
	if m.Name != wantName {
		t.Errorf("Name = %q; want %q", m.Name, wantName)
	}

	// Phase must be reset to "researching"
	if m.Phase != "researching" {
		t.Errorf("Phase = %q; want %q", m.Phase, "researching")
	}

	// Lessons, ShareToken, ArchivedAt must be nil
	if m.Lessons != nil {
		t.Errorf("Lessons = %v; want nil", *m.Lessons)
	}
	if m.ShareToken != nil {
		t.Errorf("ShareToken = %v; want nil", *m.ShareToken)
	}
	if m.ArchivedAt != nil {
		t.Errorf("ArchivedAt = %v; want nil", *m.ArchivedAt)
	}

	// Slug must differ from original
	if m.Slug == "gaming-pc" {
		t.Errorf("Slug = %q; want a unique slug different from original", m.Slug)
	}

	// Other fields copied from original
	if m.Budget != 1000 {
		t.Errorf("Budget = %v; want 1000", m.Budget)
	}
	if m.Category != "electronics" {
		t.Errorf("Category = %q; want electronics", m.Category)
	}
	if len(m.Constraints) != 1 {
		t.Errorf("Constraints len = %d; want 1", len(m.Constraints))
	}
}

// TestHandler_Duplicate_NotFound verifies 404 when source mission doesn't exist.
func TestHandler_Duplicate_NotFound(t *testing.T) {
	repo := newHandlerMockRepo()
	svc := NewService(repo)

	r := routerWithDuplicate(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/nonexistent/duplicate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

// TestHandler_Duplicate_SlugUniqueness verifies the duplicate gets a unique slug.
func TestHandler_Duplicate_SlugUniqueness(t *testing.T) {
	repo := newHandlerMockRepo()
	svc := NewService(repo)
	seedMission(repo, "Gaming PC", "gaming-pc", "researching")

	r := routerWithDuplicate(svc)

	// First duplicate → "gaming-pc-copie"
	req1 := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/duplicate", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first duplicate status = %d; want 201\nbody: %s", w1.Code, w1.Body.String())
	}

	var m1 Mission
	if err := json.Unmarshal(w1.Body.Bytes(), &m1); err != nil {
		t.Fatalf("decode m1: %v", err)
	}

	// Second duplicate → must not collide with the first
	req2 := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/duplicate", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("second duplicate status = %d; want 201\nbody: %s", w2.Code, w2.Body.String())
	}

	var m2 Mission
	if err := json.Unmarshal(w2.Body.Bytes(), &m2); err != nil {
		t.Fatalf("decode m2: %v", err)
	}

	if m1.Slug == m2.Slug {
		t.Errorf("duplicate slugs collide: both = %q", m1.Slug)
	}
}

// TestService_Duplicate_ResetsFields verifies service-level field resets.
func TestService_Duplicate_ResetsFields(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed a done mission with all optional fields set
	id := uuid.New()
	original := &Mission{
		ID:             id,
		Slug:           "laptop",
		Name:           "Laptop",
		Icon:           "laptop",
		Category:       "computing",
		Budget:         2000,
		Currency:       "USD",
		Locale:         "en-US",
		Phase:          "done",
		Constraints:    []Constraint{{Key: "k", Label: "label", Value: "v", Type: "soft"}},
		CostCategories: []string{"hardware"},
		WeightProfile:  WeightProfile{Price: 2, Quality: 3, Feature: 1},
		Lessons:        ptr("lessons text"),
		ShareToken:     ptr("sharetoken"),
		ArchivedAt:     ptrTime(time.Now()),
	}
	repo.missions["laptop"] = original
	repo.byID[id] = original

	dup, err := svc.Duplicate(ctx, "laptop")
	if err != nil {
		t.Fatalf("Duplicate() error: %v", err)
	}

	// Name suffix
	if !strings.HasSuffix(dup.Name, " (copie)") {
		t.Errorf("Name = %q; want suffix ' (copie)'", dup.Name)
	}
	// Phase reset
	if dup.Phase != "researching" {
		t.Errorf("Phase = %q; want 'researching'", dup.Phase)
	}
	// Sensitive fields cleared
	if dup.Lessons != nil {
		t.Errorf("Lessons should be nil")
	}
	if dup.ShareToken != nil {
		t.Errorf("ShareToken should be nil")
	}
	if dup.ArchivedAt != nil {
		t.Errorf("ArchivedAt should be nil")
	}
	// Other fields preserved
	if dup.Budget != original.Budget {
		t.Errorf("Budget = %v; want %v", dup.Budget, original.Budget)
	}
	if dup.Currency != original.Currency {
		t.Errorf("Currency = %q; want %q", dup.Currency, original.Currency)
	}
	if dup.Icon != original.Icon {
		t.Errorf("Icon = %q; want %q", dup.Icon, original.Icon)
	}
	if dup.Category != original.Category {
		t.Errorf("Category = %q; want %q", dup.Category, original.Category)
	}
	// Slug uniqueness (different from original)
	if dup.Slug == original.Slug {
		t.Errorf("Slug = %q; must differ from original", dup.Slug)
	}
	// New ID
	if dup.ID == original.ID {
		t.Error("ID must differ from original")
	}
}

// TestService_Duplicate_NotFound verifies nil is returned when slug missing.
func TestService_Duplicate_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Duplicate(context.Background(), "nope")
	if err == nil {
		t.Error("Duplicate(nonexistent) = nil; want error")
	}
}
