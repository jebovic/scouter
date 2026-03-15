package product

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes product barcode lookup over HTTP.
type Handler struct {
	looker *Looker
}

// NewHandler returns a Handler backed by the given Looker.
func NewHandler(looker *Looker) *Handler {
	return &Handler{looker: looker}
}

// Routes returns a chi sub-router for the product package.
// Register with: r.Route("/api/products", product.Routes(looker))
func Routes(looker *Looker) func(chi.Router) {
	h := NewHandler(looker)
	return func(r chi.Router) {
		r.Get("/{barcode}", h.Lookup)
	}
}

// Lookup handles GET /{barcode} and returns ProductInfo JSON.
// Returns 404 when the barcode is absent from Open Food Facts.
func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	barcode := chi.URLParam(r, "barcode")
	if barcode == "" {
		httputil.WriteError(w, http.StatusBadRequest, "barcode is required")
		return
	}

	info, err := h.looker.LookupByBarcode(r.Context(), barcode)
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "barcode lookup failed")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, info)
}
