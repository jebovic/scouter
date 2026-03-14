package template

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves template endpoints.
// Routes are registered directly on the top-level router — no sub-resource.
type Handler struct {
	reg *Registry
}

// NewHandler creates a template handler backed by the given registry.
func NewHandler(reg *Registry) *Handler {
	return &Handler{reg: reg}
}

// List handles GET /api/templates.
func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	items := h.reg.All()
	httputil.WriteJSON(w, http.StatusOK, map[string][]Template{"items": items})
}

// Get handles GET /api/templates/{slug}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	tmpl, err := h.reg.GetBySlug(slug)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "template not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tmpl)
}
