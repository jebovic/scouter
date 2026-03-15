package comment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// MissionResolver resolves a mission slug to its UUID string.
type MissionResolver interface {
	GetMissionIDBySlug(ctx context.Context, slug string) (id string, found bool, err error)
}

// Handler exposes comment CRUD as chi routes under /api/missions/{slug}/comments.
type Handler struct {
	repo     Repository
	missions MissionResolver
}

// NewHandler creates a new comment handler.
func NewHandler(repo Repository, missions MissionResolver) *Handler {
	return &Handler{repo: repo, missions: missions}
}

// Routes returns a chi.Router with all comment routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{commentId}", h.Delete)
	return r
}

// List handles GET /api/missions/{slug}/comments?optionId=xxx
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	missionID, found, err := h.missions.GetMissionIDBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	var optionID *string
	if raw := r.URL.Query().Get("optionId"); raw != "" {
		optionID = &raw
	}

	comments, err := h.repo.List(r.Context(), missionID, optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if comments == nil {
		comments = []Comment{}
	}
	httputil.WriteJSON(w, http.StatusOK, comments)
}

// Create handles POST /api/missions/{slug}/comments
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	missionID, found, err := h.missions.GetMissionIDBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteValidationError(w, err)
		return
	}

	comment, err := h.repo.Create(r.Context(), missionID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, comment)
}

// Delete handles DELETE /api/missions/{slug}/comments/{commentId}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	_, found, err := h.missions.GetMissionIDBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	commentID := chi.URLParam(r, "commentId")
	if err := h.repo.Delete(r.Context(), commentID); err != nil {
		if IsNotFound(err) {
			httputil.WriteError(w, http.StatusNotFound, "comment not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
