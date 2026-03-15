package settings

import (
	"encoding/json"
	"net/http"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves GET /api/settings and PATCH /api/settings.
type Handler struct {
	repo Repository
}

// NewHandler returns a Handler backed by the given Repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// GetAll handles GET /api/settings — returns all settings as a key->value map.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	s, err := h.repo.GetAll(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

// Update handles PATCH /api/settings — accepts a partial map of key->value
// and updates only the provided keys. Unknown keys are rejected.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	allowedKeys := map[string]bool{
		"currency":     true,
		"locale":       true,
		"llm_provider": true,
	}

	for key, val := range req {
		if !allowedKeys[key] {
			httputil.WriteError(w, http.StatusBadRequest, "unknown setting key: "+key)
			return
		}
		if err := h.repo.Set(r.Context(), key, val); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update setting")
			return
		}
	}

	s, err := h.repo.GetAll(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}
