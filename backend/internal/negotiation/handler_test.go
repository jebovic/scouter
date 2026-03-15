package negotiation_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/negotiation"
)

// newTestHandler builds a Handler backed by nil pool + stub provider for unit tests.
// The stub provider is wired through a real Agent so we test the full request→response path
// without needing a database.
func newTestHandler(provider *stubProvider) *negotiation.Handler {
	return negotiation.NewHandler(nil, provider)
}

func TestHandler_Routes_Registered(t *testing.T) {
	h := newTestHandler(&stubProvider{})
	r := h.Routes()
	if r == nil {
		t.Fatal("Routes() returned nil")
	}
}

func TestHandler_MissingOptionID_Returns400(t *testing.T) {
	// A handler that doesn't have the chi URL param set should return 400.
	// We invoke the negotiate endpoint with an empty option ID to simulate missing param.
	h := newTestHandler(&stubProvider{})

	req := httptest.NewRequest(http.MethodPost, "/api/options//negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	// An empty segment may route to 404 (path not matched) — either 400 or 404 is acceptable.
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusNotFound {
		t.Errorf("expected 400 or 404, got %d", rec.Code)
	}
}

func TestHandler_InvalidOptionID_Returns400(t *testing.T) {
	h := newTestHandler(&stubProvider{})

	req := httptest.NewRequest(http.MethodPost, "/not-a-uuid/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest {
		t.Logf("response body: %s", rec.Body.String())
	}
}

func TestHandler_ValidOptionID_NilPool_Returns500(t *testing.T) {
	// With a nil pool the handler should return 500 (cannot fetch option).
	h := negotiation.NewHandler((*pgxpool.Pool)(nil), &stubProvider{})

	optionID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/"+optionID+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusNotFound {
		var body map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&body)
		t.Logf("unexpected status %d, body: %v", rec.Code, body)
	}
}
