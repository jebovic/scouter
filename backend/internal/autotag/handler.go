package autotag

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

// MissionGetter retrieves a mission by slug.
// Tests inject a fake; production wires the mission repository.
type MissionGetter interface {
	GetBySlug(ctx context.Context, slug string) (*mission.Mission, error)
}

// Handler exposes the suggest-category endpoint over HTTP.
type Handler struct {
	missions MissionGetter
	tagger   TaggerInterface
}

// NewHandler returns a Handler wired to the given mission getter and tagger.
func NewHandler(missions MissionGetter, tagger TaggerInterface) *Handler {
	return &Handler{missions: missions, tagger: tagger}
}

// Suggest handles POST /api/missions/{slug}/suggest-category.
// It looks up the mission, passes its name and constraints to the tagger,
// and falls back to keyword-based matching when the LLM is unavailable.
func (h *Handler) Suggest(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	m, err := h.missions.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Build tag input from mission fields.
	constraints := make([]string, 0, len(m.Constraints))
	for _, c := range m.Constraints {
		constraints = append(constraints, c.Label)
	}

	input := TagInput{
		MissionSlug: m.Slug,
		Name:        m.Name,
		Constraints: constraints,
	}

	suggestion, err := h.tagger.Suggest(r.Context(), input)
	if err != nil {
		// Fall back to deterministic keyword matching.
		suggestion = KeywordFallback(m.Name)
	}

	// Ensure alternatives is never null in JSON output.
	if suggestion.Alternatives == nil {
		suggestion.Alternatives = []string{}
	}

	httputil.WriteJSON(w, http.StatusOK, suggestion)
}
