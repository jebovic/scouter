package imagefetch

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

type Handler struct {
	repo     *Repository
	uploader *Uploader
}

func NewHandler(repo *Repository, uploader *Uploader) *Handler {
	return &Handler{repo: repo, uploader: uploader}
}

// Routes returns a chi router for image sub-routes.
// Mount at: r.Mount("/{optionID}/images", imageHandler.Routes())
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Patch("/{imageID}", h.updateSortOrder)
	r.Delete("/{imageID}", h.delete)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	optionID, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid optionID")
		return
	}

	images, err := h.repo.ListByOption(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "list images")
		return
	}

	// Populate presigned URLs
	for _, img := range images {
		u, err := h.uploader.PresignURL(r.Context(), img.MinioKey)
		if err != nil {
			log.Printf("imagefetch: presign %s: %v", img.MinioKey, err)
			continue
		}
		img.URL = u
	}
	httputil.WriteJSON(w, http.StatusOK, images)
}

type sortOrderBody struct {
	SortOrder int `json:"sortOrder"`
}

func (h *Handler) updateSortOrder(w http.ResponseWriter, r *http.Request) {
	imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid imageID")
		return
	}
	var body sortOrderBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.repo.UpdateSortOrder(r.Context(), imageID, body.SortOrder); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "update sort order")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid imageID")
		return
	}

	img, err := h.repo.Delete(r.Context(), imageID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "delete image")
		return
	}

	// Best-effort MinIO cleanup; log but don't fail.
	if err := h.uploader.Delete(r.Context(), img.MinioKey); err != nil {
		log.Printf("imagefetch: best-effort minio delete %s: %v", img.MinioKey, err)
	}

	w.WriteHeader(http.StatusNoContent)
}
