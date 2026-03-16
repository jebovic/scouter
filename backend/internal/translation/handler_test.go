package translation_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/translation"
	"github.com/stretchr/testify/assert"
)

func TestRetranslateHandler_Returns202(t *testing.T) {
	submitted := make([]uuid.UUID, 0, 1)
	h := translation.NewHandler(func(id uuid.UUID) bool {
		submitted = append(submitted, id)
		return true
	})

	optID := uuid.New()
	r := chi.NewRouter()
	r.Post("/{optionID}/retranslate", h.Retranslate)

	req := httptest.NewRequest(http.MethodPost, "/"+optID.String()+"/retranslate", strings.NewReader(""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Len(t, submitted, 1)
	assert.Equal(t, optID, submitted[0])
}

func TestRetranslateHandler_BadID(t *testing.T) {
	h := translation.NewHandler(func(id uuid.UUID) bool { return true })
	r := chi.NewRouter()
	r.Post("/{optionID}/retranslate", h.Retranslate)

	req := httptest.NewRequest(http.MethodPost, "/not-a-uuid/retranslate", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
