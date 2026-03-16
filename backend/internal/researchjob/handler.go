// backend/internal/researchjob/handler.go
package researchjob

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

// MissionGetter is the subset of mission.Service the handler needs.
type MissionGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*mission.Mission, error)
}

// Handler exposes research job endpoints.
type Handler struct {
	repo       Repository
	missionSvc MissionGetter
	agent      AgentRunner
	logger     *slog.Logger
}

// NewHandler constructs a Handler.
func NewHandler(repo Repository, missionSvc MissionGetter, agent AgentRunner) *Handler {
	return &Handler{
		repo:       repo,
		missionSvc: missionSvc,
		agent:      agent,
		logger:     slog.Default(),
	}
}

// Routes registers the three research job endpoints under the parent subrouter.
// Mount at "/research" under /api/missions/{missionID}.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.trigger)
	r.Get("/jobs", h.list)
	r.Get("/jobs/{jobID}", h.get)
	return r
}

// trigger: POST /api/missions/{missionID}/research
func (h *Handler) trigger(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	ms, err := h.missionSvc.GetByID(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ms == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	active, err := h.repo.HasActiveJob(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if active {
		httputil.WriteError(w, http.StatusConflict, "research already running")
		return
	}

	var feedback string
	var input FeedbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err == nil && input.Feedback != "" {
		feedback = input.Feedback
	}

	job, err := h.repo.Create(r.Context(), missionID, feedback)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	go RunResearchJob(job.ID, *ms, feedback, h.agent, h.repo, h.logger)

	httputil.WriteJSON(w, http.StatusAccepted, map[string]any{
		"job_id": job.ID,
		"status": StatusPending,
	})
}

// list: GET /api/missions/{missionID}/research/jobs
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	jobs, err := h.repo.GetByMission(r.Context(), missionID, 10)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if jobs == nil {
		jobs = []ResearchJob{}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

// get: GET /api/missions/{missionID}/research/jobs/{jobID}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.repo.GetByID(r.Context(), jobID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if job == nil {
		httputil.WriteError(w, http.StatusNotFound, "job not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, job)
}
