package translation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// SubmitFn is the function signature for submitting a UUID to the translation worker.
type SubmitFn func(uuid.UUID) bool

// Handler exposes the retranslate endpoint.
type Handler struct {
	submit SubmitFn
}

// NewHandler creates a retranslate handler.
func NewHandler(submit SubmitFn) *Handler {
	return &Handler{submit: submit}
}

// Retranslate handles POST /{optionID}/retranslate — re-queues the option.
func (h *Handler) Retranslate(w http.ResponseWriter, r *http.Request) {
	optionID, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option id")
		return
	}
	// Drop-on-full is intentional: retranslate is best-effort.
	// 202 is returned regardless so the client can retry naturally via polling.
	_ = h.submit(optionID)
	w.WriteHeader(http.StatusAccepted)
}
