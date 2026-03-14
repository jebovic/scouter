package export

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves mission export endpoints.
type Handler struct {
	gatherer *Gatherer
}

// NewHandler creates a new export handler.
func NewHandler(gatherer *Gatherer) *Handler {
	return &Handler{gatherer: gatherer}
}

// Export handles GET /api/missions/{missionID}/export?format=json|markdown|pdf
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, err := h.gatherer.Gather(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if data == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	slug := data.Mission.Slug

	switch format {
	case "json":
		b, err := ToJSON(data)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "export failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.json"`, slug))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

	case "markdown":
		b, err := ToMarkdown(data)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "export failed")
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, slug))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

	case "pdf":
		b, err := ToPDF(data)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "export failed")
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, slug))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)

	default:
		httputil.WriteError(w, http.StatusBadRequest, "format must be json, markdown, or pdf")
	}
}
