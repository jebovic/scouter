package collaborator

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves collaborator and invite endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new collaborator handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// InviteRoutes returns routes mounted at /api/missions/{missionID}/invites.
func (h *Handler) InviteRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createInvite)
	return r
}

// CollaboratorRoutes returns routes mounted at /api/missions/{missionID}/collaborators.
func (h *Handler) CollaboratorRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listCollaborators)
	r.Delete("/{id}", h.removeCollaborator)
	return r
}

// GetInvite handles GET /api/invites/{token}.
func (h *Handler) GetInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	c, err := h.svc.GetInviteInfo(r.Context(), token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "invite not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, c)
}

// JoinMission handles POST /api/invites/{token}/join.
func (h *Handler) JoinMission(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.svc.JoinMission(r.Context(), token, req.DisplayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "invite not found")
			return
		}
		// display name validation error
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) createInvite(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if !req.Role.Valid() {
		httputil.WriteError(w, http.StatusBadRequest, "role must be viewer, editor, or voter")
		return
	}

	c, err := h.svc.CreateInvite(r.Context(), missionID, req.Email, req.Role)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create invite")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, c)
}

func (h *Handler) listCollaborators(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")

	collaborators, err := h.svc.ListByMission(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if collaborators == nil {
		collaborators = []Collaborator{}
	}
	httputil.WriteJSON(w, http.StatusOK, collaborators)
}

func (h *Handler) removeCollaborator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.RemoveCollaborator(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "collaborator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
