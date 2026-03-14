package option

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes option CRUD as chi routes.
// Routes are mounted under /api/missions/{missionID}/options.
type Handler struct {
	svc *Service
}

// NewHandler creates a new option handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts option routes. Expects chi URL param "missionID" from parent router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{optionID}", h.get)
	r.Put("/{optionID}", h.update)
	r.Delete("/{optionID}", h.delete)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	p := httputil.ParsePageParams(r)
	opts, err := h.svc.ListByMissionPaged(r.Context(), missionID, p.Cursor, p.Limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := httputil.BuildPagedResponse(opts, p.Limit, func(o Option) string {
		return o.CreatedAt.UTC().Format(time.RFC3339Nano)
	})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option id")
		return
	}

	o, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if o == nil {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, o)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteValidationError(w, err)
		return
	}

	o, err := h.svc.Create(r.Context(), missionID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, o)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option id")
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

	o, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if o == nil {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, o)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
