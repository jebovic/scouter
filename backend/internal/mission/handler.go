package mission

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes mission CRUD as chi routes.
type Handler struct {
	svc *Service
}

// NewHandler creates a new mission handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := httputil.ParsePageParams(r)
	includeArchived := r.URL.Query().Get("include_archived") == "true"
	missions, err := h.svc.ListPaged(r.Context(), p.Cursor, p.Limit, includeArchived)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := httputil.BuildPagedResponse(missions, p.Limit, func(m Mission) string {
		return m.CreatedAt.UTC().Format(time.RFC3339Nano)
	})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	m, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, m)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteValidationError(w, err)
		return
	}

	m, err := h.svc.Create(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, m)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	m, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteValidationError(w, err)
		return
	}

	updated, err := h.svc.Update(r.Context(), m.ID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	m, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	if err := h.svc.Delete(r.Context(), m.ID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Duplicate creates a copy of the mission identified by slug.
// The copy gets a " (copie)" name suffix, phase reset to "researching",
// and lessons/shareToken/archivedAt cleared.
func (h *Handler) Duplicate(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	m, err := h.svc.Duplicate(r.Context(), slug)
	if err != nil {
		// Distinguish not-found from infrastructure errors using a simple string check.
		// The service returns a sentinel message when the mission is not found.
		if isNotFound(err) {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, m)
}

// Clone creates a deep copy of a mission — including all options and shopping
// items — identified by slug. POST /api/missions/{slug}/clone
func (h *Handler) Clone(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	cloned, err := h.svc.CloneMission(r.Context(), slug)
	if err != nil {
		if isNotFound(err) {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "clone failed")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, cloned)
}

// isNotFound reports whether err wraps the ErrNotFound sentinel from the service.
func isNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// Archive sets archived_at on a mission, hiding it from the default dashboard list.
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	m, err := h.svc.Archive(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, m)
}

// Unarchive clears archived_at, restoring a mission to the dashboard.
func (h *Handler) Unarchive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	m, err := h.svc.Unarchive(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, m)
}

// Share generates or returns the existing share token for a mission.
func (h *Handler) Share(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	m, err := h.svc.GenerateShareToken(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"shareToken": *m.ShareToken,
		"shareURL":   "/shared/" + *m.ShareToken,
	})
}

// RevokeShare removes the share token from a mission.
func (h *Handler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	if err := h.svc.RevokeShareToken(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
