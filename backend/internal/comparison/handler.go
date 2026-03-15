package comparison

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler handles HTTP requests for comparison weights and matrix scoring.
type Handler struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewHandler returns a Handler.
func NewHandler(repo *Repository, pool *pgxpool.Pool) *Handler {
	return &Handler{repo: repo, pool: pool}
}

// WeightRoutes returns a chi route function for the comparison-weights sub-resource.
func WeightRoutes(repo *Repository, pool *pgxpool.Pool) func(chi.Router) {
	h := NewHandler(repo, pool)
	return func(r chi.Router) {
		r.Get("/", h.ListWeights)
		r.Put("/", h.SetWeight)
		r.Delete("/{criteria}", h.DeleteWeight)
	}
}

// ListWeights handles GET /api/missions/{missionID}/comparison-weights.
func (h *Handler) ListWeights(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")
	weights, err := h.repo.ListWeights(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list weights")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, weights)
}

// SetWeight handles PUT /api/missions/{missionID}/comparison-weights.
func (h *Handler) SetWeight(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")

	var req SetWeightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	weight, err := h.repo.SetWeight(r.Context(), missionID, req)
	if err != nil {
		if isValidationErr(err) {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to set weight")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, weight)
}

// DeleteWeight handles DELETE /api/missions/{missionID}/comparison-weights/{criteria}.
func (h *Handler) DeleteWeight(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")
	criteria := chi.URLParam(r, "criteria")

	err := h.repo.DeleteWeight(r.Context(), missionID, criteria)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "weight not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete weight")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Matrix handles GET /api/missions/{missionID}/matrix.
// It fetches all weights for the mission, then fetches all options with their
// attributes JSONB, computes weighted scores, and returns a ranked MatrixResponse.
func (h *Handler) Matrix(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")

	weights, err := h.repo.ListWeights(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch weights")
		return
	}

	opts, err := h.fetchRawOptions(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to fetch options")
		return
	}

	scores := computeScores(opts, weights)

	httputil.WriteJSON(w, http.StatusOK, MatrixResponse{
		Weights: weights,
		Scores:  scores,
	})
}

// fetchRawOptions queries the options table for mission and returns rawOption slices.
// Attributes JSONB is decoded into a flat map[string]interface{} where each key
// comes from the Attribute.Key field of the stored JSON array.
func (h *Handler) fetchRawOptions(ctx context.Context, missionID string) ([]rawOption, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, attributes::text
		FROM options
		WHERE mission_id = $1
		ORDER BY created_at ASC
	`, missionID)
	if err != nil {
		return nil, fmt.Errorf("query options: %w", err)
	}
	defer rows.Close()

	var opts []rawOption
	for rows.Next() {
		var id, name string
		var attrsText *string
		if err := rows.Scan(&id, &name, &attrsText); err != nil {
			return nil, fmt.Errorf("scan option row: %w", err)
		}

		attrs := map[string]interface{}{}
		if attrsText != nil && *attrsText != "" {
			attrs = parseAttributesJSON(*attrsText)
		}
		opts = append(opts, rawOption{id: id, name: name, attrs: attrs})
	}
	return opts, rows.Err()
}

// parseAttributesJSON decodes a JSONB attributes array ([]Attribute) into a
// flat key→value map. Non-numeric values are retained as-is; the scoring layer
// will skip them.
//
// The stored format is: [{"key":"price","value":499.99,"type":"price",...}, ...]
func parseAttributesJSON(raw string) map[string]interface{} {
	result := map[string]interface{}{}

	var attrs []struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
		return result
	}
	for _, a := range attrs {
		result[a.Key] = a.Value
	}
	return result
}

// isValidationErr returns true if err is a validation error from this package.
func isValidationErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "weight must be between 0 and 100" ||
		msg == "criteria must not be empty"
}
