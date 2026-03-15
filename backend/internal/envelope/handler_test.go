package envelope

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// repoInterface mirrors the Repository methods used by the Service/Handler.
type repoInterface interface {
	FindAll(ctx context.Context) ([]Envelope, error)
	Create(ctx context.Context, e Envelope) (Envelope, error)
	Update(ctx context.Context, e Envelope) (Envelope, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error)
}

// testService wraps a mockRepository to satisfy Service-level logic in handler tests.
type testService struct {
	repo repoInterface
}

func newTestService(repo repoInterface) *testService {
	return &testService{repo: repo}
}

func (s *testService) ListAll(ctx context.Context) ([]Envelope, error) {
	return s.repo.FindAll(ctx)
}

func (s *testService) CreateEnvelope(ctx context.Context, req CreateRequest) (Envelope, error) {
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

func (s *testService) UpdateEnvelope(ctx context.Context, id uuid.UUID, req UpdateRequest) (Envelope, error) {
	if req.Name != nil && *req.Name == "" {
		return Envelope{}, ErrNameRequired
	}
	if req.MonthlyAmount != nil && *req.MonthlyAmount <= 0 {
		return Envelope{}, ErrInvalidAmount
	}
	existing, err := s.repo.FindAll(ctx)
	if err != nil {
		return Envelope{}, err
	}
	var target Envelope
	found := false
	for _, e := range existing {
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

func (s *testService) DeleteEnvelope(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *testService) GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error) {
	return s.repo.GetSummary(ctx, id)
}

// handlerFromMock builds a Handler backed by a mock repository.
func handlerFromMock(mock *mockRepository) *Handler {
	svc := NewService(mock)
	return NewHandler(svc)
}

func routerWithHandler(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Route("/api/envelopes", Routes(h))
	return r
}

// ---------- GET /api/envelopes ----------

func TestHandler_List_Empty(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var items []Envelope
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d", len(items))
	}
}

func TestHandler_List_WithItems(t *testing.T) {
	mock := newMockRepository()
	mock.Create(context.Background(), Envelope{Name: "Rent", MonthlyAmount: 1200, Currency: "EUR"})
	mock.Create(context.Background(), Envelope{Name: "Food", MonthlyAmount: 400, Currency: "EUR"})

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var items []Envelope
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestHandler_List_DBError(t *testing.T) {
	mock := newMockRepository()
	mock.err = errors.New("connection refused")

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ---------- POST /api/envelopes ----------

func TestHandler_Create_Success(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	body := `{"name":"Groceries","monthlyAmount":300,"currency":"EUR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if env.Name != "Groceries" {
		t.Errorf("expected name Groceries, got %s", env.Name)
	}
	if env.MonthlyAmount != 300 {
		t.Errorf("expected 300, got %f", env.MonthlyAmount)
	}
	if env.ID == uuid.Nil {
		t.Error("expected non-nil UUID in response")
	}
}

func TestHandler_Create_DefaultCurrency(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	body := `{"name":"Savings","monthlyAmount":500}`
	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var env Envelope
	json.NewDecoder(w.Body).Decode(&env)
	if env.Currency != "EUR" {
		t.Errorf("expected default EUR, got %s", env.Currency)
	}
}

func TestHandler_Create_MissingName(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	body := `{"monthlyAmount":300,"currency":"EUR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_ZeroAmount(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	body := `{"name":"Test","monthlyAmount":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_NegativeAmount(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	body := `{"name":"Test","monthlyAmount":-50}`
	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodPost, "/api/envelopes", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------- PATCH /api/envelopes/:id ----------

func TestHandler_Update_Success(t *testing.T) {
	mock := newMockRepository()
	created, _ := mock.Create(context.Background(), Envelope{Name: "Old", MonthlyAmount: 100, Currency: "EUR"})

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	newAmount := 200.0
	body, _ := json.Marshal(UpdateRequest{MonthlyAmount: &newAmount})
	req := httptest.NewRequest(http.MethodPatch, "/api/envelopes/"+created.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var env Envelope
	json.NewDecoder(w.Body).Decode(&env)
	if env.MonthlyAmount != 200 {
		t.Errorf("expected 200, got %f", env.MonthlyAmount)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	newAmount := 200.0
	body, _ := json.Marshal(UpdateRequest{MonthlyAmount: &newAmount})
	req := httptest.NewRequest(http.MethodPatch, "/api/envelopes/"+uuid.New().String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Update_InvalidUUID(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/envelopes/not-a-uuid", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodPatch, "/api/envelopes/"+uuid.New().String(), bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------- GET /api/envelopes/:id/summary ----------

func TestHandler_Summary_Success(t *testing.T) {
	mock := newMockRepository()
	created, _ := mock.Create(context.Background(), Envelope{Name: "Rent", MonthlyAmount: 1200, Currency: "EUR"})

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes/"+created.ID.String()+"/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var summary EnvelopeSummary
	if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if summary.Envelope.ID != created.ID {
		t.Errorf("expected envelope id %s, got %s", created.ID, summary.Envelope.ID)
	}
	if summary.Envelope.Name != "Rent" {
		t.Errorf("expected name Rent, got %s", summary.Envelope.Name)
	}
	// mock returns zeros for aggregates
	if summary.MissionCount != 0 {
		t.Errorf("expected missionCount 0, got %d", summary.MissionCount)
	}
	if summary.TotalBudget != 0 {
		t.Errorf("expected totalBudget 0, got %f", summary.TotalBudget)
	}
	if summary.TotalSpent != 0 {
		t.Errorf("expected totalSpent 0, got %f", summary.TotalSpent)
	}
}

func TestHandler_Summary_NotFound(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes/"+uuid.New().String()+"/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Summary_InvalidUUID(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes/not-a-uuid/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Summary_DBError(t *testing.T) {
	mock := newMockRepository()
	created, _ := mock.Create(context.Background(), Envelope{Name: "Food", MonthlyAmount: 400, Currency: "EUR"})
	mock.err = errors.New("db timeout")

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodGet, "/api/envelopes/"+created.ID.String()+"/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ---------- DELETE /api/envelopes/:id ----------

func TestHandler_Delete_Success(t *testing.T) {
	mock := newMockRepository()
	created, _ := mock.Create(context.Background(), Envelope{Name: "Rent", MonthlyAmount: 1200, Currency: "EUR"})

	h := handlerFromMock(mock)
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/envelopes/"+created.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/envelopes/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Delete_InvalidUUID(t *testing.T) {
	h := handlerFromMock(newMockRepository())
	r := routerWithHandler(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/envelopes/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
