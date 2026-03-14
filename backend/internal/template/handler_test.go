package template

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func newTestRouter() chi.Router {
	reg := NewRegistry()
	h := NewHandler(reg)
	r := chi.NewRouter()
	r.Get("/api/templates", h.List)
	r.Get("/api/templates/{slug}", h.Get)
	return r
}

func TestHandler_List_Returns200(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_List_ReturnsItemsArray(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Items []Template `json:"items"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) < 10 {
		t.Errorf("items count = %d; want >= 10", len(resp.Items))
	}
}

func TestHandler_List_ContentTypeJSON(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q; want %q", ct, "application/json")
	}
}

func TestHandler_Get_Returns200ForKnownSlug(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates/laptop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get(laptop) status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_Get_ReturnsCorrectTemplate(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates/laptop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var tmpl Template
	if err := json.NewDecoder(w.Body).Decode(&tmpl); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tmpl.Slug != "laptop" {
		t.Errorf("Slug = %q; want %q", tmpl.Slug, "laptop")
	}
}

func TestHandler_Get_Returns404ForUnknownSlug(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates/does-not-exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get(unknown) status = %d; want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_Get_404BodyHasErrorField(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/templates/does-not-exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] == "" {
		t.Error("404 body missing 'error' field")
	}
}
