package notification

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jibei/scouter/internal/httputil"
)

const defaultNotifLimit = 50

// Handler exposes notification endpoints as chi routes.
// Routes are mounted at /api/notifications.
type Handler struct {
	repo Repository
}

// NewHandler creates a new notification handler.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// Routes mounts notification routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/unread-count", h.unreadCount)
	r.Patch("/{id}/read", h.markRead)
	r.Post("/read-all", h.markAllRead)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	p := httputil.ParsePageParams(r)
	limit := p.Limit
	if limit <= 0 {
		limit = defaultNotifLimit
	}

	notifs, err := h.repo.List(r.Context(), limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if notifs == nil {
		notifs = []Notification{}
	}
	httputil.WriteJSON(w, http.StatusOK, notifs)
}

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.repo.UnreadCount(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.repo.MarkRead(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "notification not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.MarkAllRead(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
