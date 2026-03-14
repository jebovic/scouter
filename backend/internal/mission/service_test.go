package mission

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepo is a stub Repository for unit testing the Service.
type mockRepo struct {
	missions map[string]*Mission // keyed by slug
	byID     map[uuid.UUID]*Mission
	nextErr  error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		missions: make(map[string]*Mission),
		byID:     make(map[uuid.UUID]*Mission),
	}
}

func (m *mockRepo) List(_ context.Context) ([]Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	out := make([]Mission, 0, len(m.missions))
	for _, v := range m.missions {
		out = append(out, *v)
	}
	return out, nil
}

func (m *mockRepo) ListActive(_ context.Context) ([]Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	out := make([]Mission, 0)
	for _, v := range m.missions {
		if v.Phase != "done" {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *mockRepo) ListPaged(_ context.Context, _ *time.Time, limit int) ([]Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	out := make([]Mission, 0, len(m.missions))
	for _, v := range m.missions {
		out = append(out, *v)
	}
	if len(out) > limit+1 {
		out = out[:limit+1]
	}
	return out, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	v, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockRepo) GetBySlug(_ context.Context, slug string) (*Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	v, ok := m.missions[slug]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockRepo) Create(_ context.Context, ms Mission) (*Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	ms.ID = uuid.New()
	m.missions[ms.Slug] = &ms
	m.byID[ms.ID] = &ms
	return &ms, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, req UpdateRequest) (*Mission, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	ms, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	if req.Name != nil {
		ms.Name = *req.Name
	}
	return ms, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	ms, ok := m.byID[id]
	if ok {
		delete(m.missions, ms.Slug)
		delete(m.byID, id)
	}
	return nil
}

func TestService_Create_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr bool
	}{
		{
			name:    "valid mission",
			req:     CreateRequest{Name: "Gaming PC", Budget: 1500},
			wantErr: false,
		},
		{
			name:    "empty name",
			req:     CreateRequest{Name: "", Budget: 1000},
			wantErr: true,
		},
		{
			name:    "whitespace name",
			req:     CreateRequest{Name: "   ", Budget: 1000},
			wantErr: true,
		},
		{
			name:    "zero budget",
			req:     CreateRequest{Name: "Laptop", Budget: 0},
			wantErr: true,
		},
		{
			name:    "negative budget",
			req:     CreateRequest{Name: "Laptop", Budget: -100},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(newMockRepo())
			_, err := svc.Create(context.Background(), tt.req)
			if tt.wantErr && err == nil {
				t.Errorf("Create(%q) = nil; want error", tt.req.Name)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Create(%q) = %v; want nil", tt.req.Name, err)
			}
		})
	}
}

func TestService_Create_Defaults(t *testing.T) {
	svc := NewService(newMockRepo())

	m, err := svc.Create(context.Background(), CreateRequest{
		Name:   "Trip to Japan",
		Budget: 3000,
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if m.Icon != "target" {
		t.Errorf("Icon = %q; want %q", m.Icon, "target")
	}
	if m.Currency != "EUR" {
		t.Errorf("Currency = %q; want %q", m.Currency, "EUR")
	}
	if m.Locale != "fr-FR" {
		t.Errorf("Locale = %q; want %q", m.Locale, "fr-FR")
	}
	if m.Phase != "researching" {
		t.Errorf("Phase = %q; want %q", m.Phase, "researching")
	}
	if m.Constraints == nil {
		t.Error("Constraints = nil; want empty slice")
	}
	if m.CostCategories == nil {
		t.Error("CostCategories = nil; want empty slice")
	}
}

func TestService_SlugGeneration(t *testing.T) {
	tests := []struct {
		name     string
		wantSlug string
	}{
		{"Gaming PC", "gaming-pc"},
		{"Trip to Japan!", "trip-to-japan"},
		{"  hello   world  ", "hello-world"},
		{"Rénovation 2026", "r-novation-2026"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(newMockRepo())
			m, err := svc.Create(context.Background(), CreateRequest{
				Name:   tt.name,
				Budget: 100,
			})
			if err != nil {
				t.Fatalf("Create() error: %v", err)
			}
			if m.Slug != tt.wantSlug {
				t.Errorf("Slug = %q; want %q", m.Slug, tt.wantSlug)
			}
		})
	}
}

func TestService_SlugUniqueness(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	m1, _ := svc.Create(ctx, CreateRequest{Name: "Gaming PC", Budget: 100})
	m2, _ := svc.Create(ctx, CreateRequest{Name: "Gaming PC", Budget: 200})
	m3, _ := svc.Create(ctx, CreateRequest{Name: "Gaming PC", Budget: 300})

	if m1.Slug != "gaming-pc" {
		t.Errorf("m1.Slug = %q; want %q", m1.Slug, "gaming-pc")
	}
	if m2.Slug != "gaming-pc-2" {
		t.Errorf("m2.Slug = %q; want %q", m2.Slug, "gaming-pc-2")
	}
	if m3.Slug != "gaming-pc-3" {
		t.Errorf("m3.Slug = %q; want %q", m3.Slug, "gaming-pc-3")
	}
}

func TestService_List_NilToEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	missions, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if missions == nil {
		t.Error("List() = nil; want empty slice")
	}
	if len(missions) != 0 {
		t.Errorf("List() len = %d; want 0", len(missions))
	}
}
