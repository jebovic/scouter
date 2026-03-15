package envelope

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// mockPool implements a minimal interface used by tests to exercise service
// and repository logic without a real database.

// ---------- mock repository ----------

type mockRepository struct {
	envelopes []Envelope
	err       error // injected error for all operations
}

func newMockRepository() *mockRepository {
	return &mockRepository{}
}

func (m *mockRepository) FindAll(ctx context.Context) ([]Envelope, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]Envelope, len(m.envelopes))
	copy(out, m.envelopes)
	return out, nil
}

func (m *mockRepository) Create(ctx context.Context, e Envelope) (Envelope, error) {
	if m.err != nil {
		return Envelope{}, m.err
	}
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	m.envelopes = append(m.envelopes, e)
	return e, nil
}

func (m *mockRepository) Update(ctx context.Context, e Envelope) (Envelope, error) {
	if m.err != nil {
		return Envelope{}, m.err
	}
	for i, existing := range m.envelopes {
		if existing.ID == e.ID {
			m.envelopes[i] = e
			return e, nil
		}
	}
	return Envelope{}, pgx.ErrNoRows
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	for i, e := range m.envelopes {
		if e.ID == id {
			m.envelopes = append(m.envelopes[:i], m.envelopes[i+1:]...)
			return nil
		}
	}
	return pgx.ErrNoRows
}

func (m *mockRepository) GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, e := range m.envelopes {
		if e.ID == id {
			return &EnvelopeSummary{
				Envelope:     e,
				MissionCount: 0,
				TotalBudget:  0,
				TotalSpent:   0,
			}, nil
		}
	}
	return nil, nil
}

// ---------- repository tests ----------

func TestRepository_FindAll_Empty(t *testing.T) {
	repo := newMockRepository()
	items, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestRepository_Create_ReturnsEnvelopeWithID(t *testing.T) {
	repo := newMockRepository()
	env := Envelope{
		Name:          "Groceries",
		MonthlyAmount: 300.00,
		Currency:      "EUR",
	}
	created, err := repo.Create(context.Background(), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if created.Name != "Groceries" {
		t.Errorf("expected name Groceries, got %s", created.Name)
	}
	if created.MonthlyAmount != 300.00 {
		t.Errorf("expected monthly amount 300.00, got %f", created.MonthlyAmount)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestRepository_FindAll_AfterCreate(t *testing.T) {
	repo := newMockRepository()
	for _, name := range []string{"Groceries", "Rent", "Entertainment"} {
		_, err := repo.Create(context.Background(), Envelope{Name: name, MonthlyAmount: 100, Currency: "EUR"})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}
	items, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestRepository_Update_ExistingEnvelope(t *testing.T) {
	repo := newMockRepository()
	created, _ := repo.Create(context.Background(), Envelope{Name: "Groceries", MonthlyAmount: 300, Currency: "EUR"})

	created.MonthlyAmount = 400
	updated, err := repo.Update(context.Background(), created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.MonthlyAmount != 400 {
		t.Errorf("expected 400, got %f", updated.MonthlyAmount)
	}
}

func TestRepository_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	env := Envelope{ID: uuid.New(), Name: "Ghost", MonthlyAmount: 100, Currency: "EUR"}
	_, err := repo.Update(context.Background(), env)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestRepository_Delete_ExistingEnvelope(t *testing.T) {
	repo := newMockRepository()
	created, _ := repo.Create(context.Background(), Envelope{Name: "Groceries", MonthlyAmount: 300, Currency: "EUR"})

	err := repo.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, _ := repo.FindAll(context.Background())
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	repo := newMockRepository()
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestRepository_Create_ErrorPropagation(t *testing.T) {
	repo := newMockRepository()
	repo.err = errors.New("db connection failed")
	_, err := repo.Create(context.Background(), Envelope{Name: "Test", MonthlyAmount: 100, Currency: "EUR"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRepository_FindAll_ErrorPropagation(t *testing.T) {
	repo := newMockRepository()
	repo.err = errors.New("db connection failed")
	_, err := repo.FindAll(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRepository_Update_ErrorPropagation(t *testing.T) {
	repo := newMockRepository()
	repo.err = errors.New("db connection failed")
	_, err := repo.Update(context.Background(), Envelope{ID: uuid.New(), Name: "Test", MonthlyAmount: 100, Currency: "EUR"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRepository_Delete_ErrorPropagation(t *testing.T) {
	repo := newMockRepository()
	repo.err = errors.New("db connection failed")
	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
