package shopping

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes shopping item CRUD and price history as chi routes.
// Routes are mounted under /api/missions/{missionID}/shopping.
type Handler struct {
	svc *Service
}

// NewHandler creates a new shopping handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts shopping routes. Expects chi URL param "missionID" from parent router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/pinned", h.deletePinned)
	r.Get("/{itemID}", h.get)
	r.Put("/{itemID}", h.update)
	r.Delete("/{itemID}", h.delete)
	r.Patch("/{itemID}/pin", h.pin)
	r.Post("/{itemID}/snapshots", h.addSnapshot)
	r.Get("/{itemID}/snapshots", h.listSnapshots)
	r.Get("/{itemID}/deal-score", h.getDealScore)
	r.Get("/{itemID}/price-history", h.priceHistory)
	r.Get("/{itemID}/price-history/export", h.ExportPriceHistoryCSV)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	p := httputil.ParsePageParams(r)
	items, err := h.svc.ListByMissionPaged(r.Context(), missionID, p.Cursor, p.Limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := httputil.BuildPagedResponse(items, p.Limit, func(item Item) string {
		return item.CreatedAt.UTC().Format(time.RFC3339Nano)
	})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
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

	item, err := h.svc.Create(r.Context(), missionID, req)
	if err != nil {
		var ve ValidationError
		if errors.As(err, &ve) {
			httputil.WriteError(w, http.StatusBadRequest, ve.Error())
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
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

	item, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addSnapshot(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	var req PriceSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteValidationError(w, err)
		return
	}

	snap, err := h.svc.AddPriceSnapshot(r.Context(), itemID, req)
	if err != nil {
		var ve ValidationError
		if errors.As(err, &ve) {
			httputil.WriteError(w, http.StatusBadRequest, ve.Error())
		} else {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, snap)
}

func (h *Handler) listSnapshots(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	snapshots, err := h.svc.ListPriceHistory(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, snapshots)
}

func (h *Handler) pin(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.svc.Pin(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getDealScore(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	score, found, err := h.svc.GetDealScore(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}
	// score may be nil when fewer than 3 price history snapshots exist — that is valid
	httputil.WriteJSON(w, http.StatusOK, score)
}

func (h *Handler) priceHistory(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	const defaultLimit = 50
	const maxLimit = 200
	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	pts, err := h.svc.GetPriceHistory(r.Context(), itemID, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, pts)
}

func (h *Handler) deletePinned(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	if err := h.svc.DeletePinned(r.Context(), missionID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ExportPriceHistoryCSV exports price history for a shopping item as a CSV file.
func (h *Handler) ExportPriceHistoryCSV(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.svc.GetByID(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	snapshots, err := h.svc.ListPriceHistory(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Build CSV
	buf := new(bytes.Buffer)
	// Write UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xef, 0xbb, 0xbf})

	writer := csv.NewWriter(buf)

	// Write header in French
	header := []string{"Date", "Prix", "Devise", "Source", "Détaillant"}
	writer.Write(header)

	// Write rows with French formatting (DD/MM/YYYY, comma decimal separator)
	for _, snap := range snapshots {
		row := []string{
			snap.RecordedAt.Format("02/01/2006"), // DD/MM/YYYY
			formatPriceFrench(snap.Price),        // 12,99 format
			"EUR",                                 // Currency
			snap.Note,                             // Source (note field)
			item.Merchant,                         // Retailer
		}
		writer.Write(row)
	}

	// Flush the CSV writer
	writer.Flush()

	// Create filename slug from item name
	itemNameSlug := strings.ToLower(strings.TrimSpace(item.Name))
	itemNameSlug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, itemNameSlug)
	itemNameSlug = strings.Trim(itemNameSlug, "-")

	// Set response headers
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"prix-%s-%s.csv\"", itemNameSlug, time.Now().Format("2006-01-02")))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

// formatPriceFrench formats a price with French conventions: comma as decimal separator.
func formatPriceFrench(price float64) string {
	s := fmt.Sprintf("%.2f", price)
	return strings.ReplaceAll(s, ".", ",")
}
