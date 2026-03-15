package receipt_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/receipt"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeReceiptRepo struct {
	saved    []receipt.ReceiptAnalysis
	listErr  error
	saveErr  error
}

func (f *fakeReceiptRepo) Save(_ context.Context, a receipt.ReceiptAnalysis) (*receipt.ReceiptAnalysis, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	a.ID = uuid.New()
	a.AnalyzedAt = time.Now()
	f.saved = append(f.saved, a)
	return &a, nil
}

func (f *fakeReceiptRepo) ListByMission(_ context.Context, slug string) ([]receipt.ReceiptAnalysis, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var out []receipt.ReceiptAnalysis
	for _, a := range f.saved {
		if a.MissionSlug == slug {
			out = append(out, a)
		}
	}
	if out == nil {
		out = []receipt.ReceiptAnalysis{}
	}
	return out, nil
}

type fakeAgent struct {
	result *receipt.AnalysisResult
	err    error
	called int
}

func (f *fakeAgent) Analyze(_ context.Context, rawText string) (*receipt.AnalysisResult, error) {
	f.called++
	return f.result, f.err
}

// ---------------------------------------------------------------------------
// Router helper
// ---------------------------------------------------------------------------

func mountRouter(h *receipt.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/missions/{slug}/receipts", h.Analyze)
	r.Get("/api/missions/{slug}/receipts", h.List)
	return r
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Test 1: blank rawText returns 422.
func TestAnalyze_BlankTextReturns422(t *testing.T) {
	repo := &fakeReceiptRepo{}
	agent := &fakeAgent{result: &receipt.AnalysisResult{}}
	h := receipt.NewHandler(repo, agent)

	body := `{"rawText": "   ", "currency": "EUR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/missions/test-mission/receipts",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Error("agent must not be called when rawText is blank")
	}
}

// Test 2: valid receipt text returns 201 with analysis.
func TestAnalyze_Success(t *testing.T) {
	repo := &fakeReceiptRepo{}
	agentResult := &receipt.AnalysisResult{
		Items: []receipt.ReceiptItem{
			{Name: "Pain", Quantity: 2, UnitPrice: 1.50, Total: 3.00},
			{Name: "Lait", Quantity: 1, UnitPrice: 1.20, Total: 1.20},
		},
		Total:    4.20,
		Currency: "EUR",
	}
	agent := &fakeAgent{result: agentResult}
	h := receipt.NewHandler(repo, agent)

	body := `{"rawText": "Pain 2x1.50\nLait 1x1.20\nTotal: 4.20", "currency": "EUR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/missions/test-mission/receipts",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var got receipt.ReceiptAnalysis
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if len(got.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(got.Items))
	}
	if got.Total != 4.20 {
		t.Errorf("expected total 4.20, got %f", got.Total)
	}
	if got.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %s", got.Currency)
	}
	if got.MissionSlug != "test-mission" {
		t.Errorf("expected mission slug test-mission, got %s", got.MissionSlug)
	}
}

// Test 3: list empty returns 200 with empty array.
func TestList_Empty(t *testing.T) {
	repo := &fakeReceiptRepo{}
	agent := &fakeAgent{}
	h := receipt.NewHandler(repo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/test-mission/receipts", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got []receipt.ReceiptAnalysis
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d items", len(got))
	}
}

// Test 4: list with data returns 200 with analyses.
func TestList_WithData(t *testing.T) {
	repo := &fakeReceiptRepo{
		saved: []receipt.ReceiptAnalysis{
			{
				ID:          uuid.New(),
				MissionSlug: "test-mission",
				RawText:     "Pain 2x1.50",
				Items:       []receipt.ReceiptItem{{Name: "Pain", Quantity: 2, UnitPrice: 1.50, Total: 3.00}},
				Total:       3.00,
				Currency:    "EUR",
				AnalyzedAt:  time.Now(),
			},
		},
	}
	agent := &fakeAgent{}
	h := receipt.NewHandler(repo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/missions/test-mission/receipts", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got []receipt.ReceiptAnalysis
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 analysis, got %d", len(got))
	}
	if got[0].Total != 3.00 {
		t.Errorf("expected total 3.00, got %f", got[0].Total)
	}
}

// Test 5: agent error returns 500.
func TestAnalyze_AgentError(t *testing.T) {
	repo := &fakeReceiptRepo{}
	agent := &fakeAgent{err: errors.New("llm unavailable")}
	h := receipt.NewHandler(repo, agent)

	body := `{"rawText": "Pain 2x1.50\nTotal: 3.00", "currency": "EUR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/missions/test-mission/receipts",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// Test 6: currency defaults to EUR when not provided.
func TestAnalyze_CurrencyDefaultsToEUR(t *testing.T) {
	repo := &fakeReceiptRepo{}
	agent := &fakeAgent{result: &receipt.AnalysisResult{
		Items:    []receipt.ReceiptItem{{Name: "Pain", Quantity: 1, UnitPrice: 1.50, Total: 1.50}},
		Total:    1.50,
		Currency: "EUR",
	}}
	h := receipt.NewHandler(repo, agent)

	body := `{"rawText": "Pain 1x1.50"}`
	req := httptest.NewRequest(http.MethodPost, "/api/missions/test-mission/receipts",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var got receipt.ReceiptAnalysis
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Currency != "EUR" {
		t.Errorf("expected default currency EUR, got %s", got.Currency)
	}
}
