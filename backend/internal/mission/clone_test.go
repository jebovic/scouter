package mission

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// cloneMockRepo extends mockRepo with per-mission options and shopping items
// so that CloneMission can deep-copy them in tests without hitting Postgres.
type cloneMockRepo struct {
	mockRepo
	// options keyed by missionID
	options map[uuid.UUID][]cloneOption
	// shopping items keyed by missionID
	shopping map[uuid.UUID][]cloneShoppingItem
	// cloneErr is returned by CloneMission when non-nil.
	cloneErr error
}

// cloneOption is a minimal stand-in for the option.Option type for testing.
type cloneOption struct {
	ID        uuid.UUID
	MissionID uuid.UUID
	Name      string
	Badge     string
}

// cloneShoppingItem is a minimal stand-in for shopping.Item for testing.
type cloneShoppingItem struct {
	ID        uuid.UUID
	MissionID uuid.UUID
	Name      string
	Price     float64
}

func newCloneMockRepo() *cloneMockRepo {
	return &cloneMockRepo{
		mockRepo: mockRepo{
			missions: make(map[string]*Mission),
			byID:     make(map[uuid.UUID]*Mission),
		},
		options:  make(map[uuid.UUID][]cloneOption),
		shopping: make(map[uuid.UUID][]cloneShoppingItem),
	}
}

// CloneMission implements Repository by performing an in-memory deep copy.
func (r *cloneMockRepo) CloneMission(ctx context.Context, slug string) (Mission, error) {
	if r.cloneErr != nil {
		return Mission{}, r.cloneErr
	}
	orig, ok := r.missions[slug]
	if !ok {
		return Mission{}, errNotFound(slug)
	}

	newID := uuid.New()
	newSlug := orig.Slug + "-copie"
	// Ensure uniqueness within the mock.
	if _, exists := r.missions[newSlug]; exists {
		newSlug = orig.Slug + "-copie-" + newID.String()[:4]
	}

	// Deep-copy constraints.
	consCopy := make([]Constraint, len(orig.Constraints))
	copy(consCopy, orig.Constraints)
	// Deep-copy cost categories.
	catsCopy := make([]string, len(orig.CostCategories))
	copy(catsCopy, orig.CostCategories)
	// Deep-copy timeline.
	tlCopy := make([]TimelineEvent, len(orig.Timeline))
	copy(tlCopy, orig.Timeline)

	cloned := Mission{
		ID:             newID,
		Slug:           newSlug,
		Name:           orig.Name + " (Copie)",
		Icon:           orig.Icon,
		Category:       orig.Category,
		Budget:         orig.Budget,
		Currency:       orig.Currency,
		Locale:         orig.Locale,
		Phase:          "researching",
		Constraints:    consCopy,
		CostCategories: catsCopy,
		Timeline:       tlCopy,
		WeightProfile:  orig.WeightProfile,
		EnvelopeID:     orig.EnvelopeID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	r.missions[newSlug] = &cloned
	r.byID[newID] = &cloned

	// Deep-copy options.
	if opts, ok := r.options[orig.ID]; ok {
		newOpts := make([]cloneOption, len(opts))
		for i, o := range opts {
			newOpts[i] = cloneOption{
				ID:        uuid.New(),
				MissionID: newID,
				Name:      o.Name,
				Badge:     o.Badge,
			}
		}
		r.options[newID] = newOpts
	}

	// Deep-copy shopping items.
	if items, ok := r.shopping[orig.ID]; ok {
		newItems := make([]cloneShoppingItem, len(items))
		for i, it := range items {
			newItems[i] = cloneShoppingItem{
				ID:        uuid.New(),
				MissionID: newID,
				Name:      it.Name,
				Price:     it.Price,
			}
		}
		r.shopping[newID] = newItems
	}

	return cloned, nil
}

// errNotFound returns a sentinel error matching isNotFound().
func errNotFound(slug string) error {
	return fmt.Errorf("mission %q: %w", slug, ErrNotFound)
}

// routerWithClone wires POST /api/missions/{slug}/clone on a fresh chi router.
func routerWithClone(svc *Service) chi.Router {
	r := chi.NewRouter()
	handler := NewHandler(svc)
	r.Post("/api/missions/{slug}/clone", handler.Clone)
	return r
}

func seedCloneMission(repo *cloneMockRepo, name, slug string) *Mission {
	id := uuid.New()
	m := &Mission{
		ID:             id,
		Slug:           slug,
		Name:           name,
		Icon:           "target",
		Category:       "electronics",
		Budget:         1500,
		Currency:       "EUR",
		Locale:         "fr-FR",
		Phase:          "buying",
		Constraints:    []Constraint{{Key: "k", Label: "l", Value: "v", Type: "hard"}},
		CostCategories: []string{"hardware"},
		Timeline:       []TimelineEvent{{Date: "2026-06-01", Label: "Deadline", Urgent: true}},
		WeightProfile:  WeightProfile{Price: 1, Quality: 2, Feature: 3},
		Lessons:        ptr("some lesson"),
		ShareToken:     ptr("sharetoken"),
		ArchivedAt:     ptrTime(time.Now()),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	repo.missions[slug] = m
	repo.byID[id] = m

	// Seed one option.
	repo.options[id] = []cloneOption{
		{ID: uuid.New(), MissionID: id, Name: "Option Alpha", Badge: "recommended"},
	}
	// Seed one shopping item.
	repo.shopping[id] = []cloneShoppingItem{
		{ID: uuid.New(), MissionID: id, Name: "Keyboard", Price: 120.0},
	}
	return m
}

// ── Handler tests ────────────────────────────────────────────────────────────

// TestHandler_Clone_Returns201 verifies the endpoint returns 201 with a new mission.
func TestHandler_Clone_Returns201(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	orig := seedCloneMission(repo, "Gaming PC", "gaming-pc")

	r := routerWithClone(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201\nbody: %s", w.Code, w.Body.String())
	}

	var m Mission
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Slug must differ from original.
	if m.Slug == orig.Slug {
		t.Errorf("Slug = %q; want a unique slug different from original %q", m.Slug, orig.Slug)
	}
	// Slug must contain "copie".
	if !strings.Contains(m.Slug, "copie") {
		t.Errorf("Slug = %q; want it to contain 'copie'", m.Slug)
	}
}

// TestHandler_Clone_NameContainsCopie verifies the cloned mission name has "(Copie)" suffix.
func TestHandler_Clone_NameContainsCopie(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	seedCloneMission(repo, "Gaming PC", "gaming-pc")

	r := routerWithClone(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201\nbody: %s", w.Code, w.Body.String())
	}

	var m Mission
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	wantName := "Gaming PC (Copie)"
	if m.Name != wantName {
		t.Errorf("Name = %q; want %q", m.Name, wantName)
	}
}

// TestHandler_Clone_PhaseReset verifies phase is reset to "researching".
func TestHandler_Clone_PhaseReset(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	seedCloneMission(repo, "Gaming PC", "gaming-pc")

	r := routerWithClone(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var m Mission
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if m.Phase != "researching" {
		t.Errorf("Phase = %q; want 'researching'", m.Phase)
	}
}

// TestHandler_Clone_SensitiveFieldsCleared verifies lessons, shareToken, archivedAt are nil.
func TestHandler_Clone_SensitiveFieldsCleared(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	seedCloneMission(repo, "Gaming PC", "gaming-pc")

	r := routerWithClone(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var m Mission
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if m.Lessons != nil {
		t.Errorf("Lessons = %v; want nil", *m.Lessons)
	}
	if m.ShareToken != nil {
		t.Errorf("ShareToken = %v; want nil", *m.ShareToken)
	}
	if m.ArchivedAt != nil {
		t.Errorf("ArchivedAt = %v; want nil", m.ArchivedAt)
	}
}

// TestHandler_Clone_NotFound verifies 404 for a non-existent slug.
func TestHandler_Clone_NotFound(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)

	r := routerWithClone(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/nonexistent/clone", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", w.Code)
	}
}

// TestHandler_Clone_SlugUniqueness verifies two clones of the same source get distinct slugs.
func TestHandler_Clone_SlugUniqueness(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	seedCloneMission(repo, "Gaming PC", "gaming-pc")

	r := routerWithClone(svc)

	var m1, m2 Mission

	req1 := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first clone status = %d; want 201\nbody: %s", w1.Code, w1.Body.String())
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &m1); err != nil {
		t.Fatalf("decode m1: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/missions/gaming-pc/clone", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("second clone status = %d; want 201\nbody: %s", w2.Code, w2.Body.String())
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &m2); err != nil {
		t.Fatalf("decode m2: %v", err)
	}

	if m1.Slug == m2.Slug {
		t.Errorf("duplicate slugs collide: both = %q", m1.Slug)
	}
}

// ── Service tests ────────────────────────────────────────────────────────────

// TestService_CloneMission_DeepCopiesOptions verifies options are deep-copied.
func TestService_CloneMission_DeepCopiesOptions(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	orig := seedCloneMission(repo, "Laptop Buy", "laptop-buy")

	cloned, err := svc.CloneMission(context.Background(), "laptop-buy")
	if err != nil {
		t.Fatalf("CloneMission() error: %v", err)
	}

	// Cloned options must exist and be linked to the new mission ID.
	clonedOpts, ok := repo.options[cloned.ID]
	if !ok || len(clonedOpts) == 0 {
		t.Fatalf("CloneMission: expected cloned options for new mission ID %s", cloned.ID)
	}
	for _, o := range clonedOpts {
		if o.MissionID != cloned.ID {
			t.Errorf("option MissionID = %s; want %s", o.MissionID, cloned.ID)
		}
		if o.ID == repo.options[orig.ID][0].ID {
			t.Errorf("option ID must differ from original")
		}
	}
}

// TestService_CloneMission_DeepCopiesShoppingItems verifies shopping items are deep-copied.
func TestService_CloneMission_DeepCopiesShoppingItems(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	orig := seedCloneMission(repo, "Laptop Buy", "laptop-buy")

	cloned, err := svc.CloneMission(context.Background(), "laptop-buy")
	if err != nil {
		t.Fatalf("CloneMission() error: %v", err)
	}

	clonedItems, ok := repo.shopping[cloned.ID]
	if !ok || len(clonedItems) == 0 {
		t.Fatalf("CloneMission: expected cloned shopping items for new mission ID %s", cloned.ID)
	}
	for _, it := range clonedItems {
		if it.MissionID != cloned.ID {
			t.Errorf("shopping item MissionID = %s; want %s", it.MissionID, cloned.ID)
		}
		if it.ID == repo.shopping[orig.ID][0].ID {
			t.Errorf("shopping item ID must differ from original")
		}
	}
}

// TestService_CloneMission_NotFound returns an error for unknown slugs.
func TestService_CloneMission_NotFound(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)

	_, err := svc.CloneMission(context.Background(), "does-not-exist")
	if err == nil {
		t.Error("CloneMission(nonexistent) = nil; want error")
	}
	if !isNotFound(err) {
		t.Errorf("error = %v; want a not-found error", err)
	}
}

// TestService_CloneMission_PreservesFields verifies non-sensitive fields are preserved.
func TestService_CloneMission_PreservesFields(t *testing.T) {
	repo := newCloneMockRepo()
	svc := NewService(repo)
	orig := seedCloneMission(repo, "Camera", "camera")

	cloned, err := svc.CloneMission(context.Background(), "camera")
	if err != nil {
		t.Fatalf("CloneMission() error: %v", err)
	}

	if cloned.Budget != orig.Budget {
		t.Errorf("Budget = %v; want %v", cloned.Budget, orig.Budget)
	}
	if cloned.Category != orig.Category {
		t.Errorf("Category = %q; want %q", cloned.Category, orig.Category)
	}
	if cloned.Icon != orig.Icon {
		t.Errorf("Icon = %q; want %q", cloned.Icon, orig.Icon)
	}
	if cloned.Currency != orig.Currency {
		t.Errorf("Currency = %q; want %q", cloned.Currency, orig.Currency)
	}
	if len(cloned.Constraints) != len(orig.Constraints) {
		t.Errorf("Constraints len = %d; want %d", len(cloned.Constraints), len(orig.Constraints))
	}
	// New ID.
	if cloned.ID == orig.ID {
		t.Error("ID must differ from original")
	}
}
