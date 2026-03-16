package imagefetch

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockScanner implements scanner for unit testing scanImage.
type mockScanner struct {
	id          uuid.UUID
	optionID    uuid.UUID
	minioKey    string
	contentType string
	width       *int
	height      *int
	sourceURL   string
	sortOrder   int
	createdAt   time.Time
	err         error
}

func (m *mockScanner) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	*dest[0].(*uuid.UUID) = m.id
	*dest[1].(*uuid.UUID) = m.optionID
	*dest[2].(*string) = m.minioKey
	*dest[3].(*string) = m.contentType
	*dest[4].(**int) = m.width
	*dest[5].(**int) = m.height
	*dest[6].(*string) = m.sourceURL
	*dest[7].(*int) = m.sortOrder
	*dest[8].(*time.Time) = m.createdAt
	return nil
}

func TestScanImage_OK(t *testing.T) {
	id := uuid.New()
	optionID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	w := 800
	ms := &mockScanner{
		id: id, optionID: optionID,
		minioKey: "options/abc/img.jpg", contentType: "image/jpeg",
		width: &w, sortOrder: 0, createdAt: now,
	}
	img, err := scanImage(ms)
	require.NoError(t, err)
	assert.Equal(t, id, img.ID)
	assert.Equal(t, optionID, img.OptionID)
	assert.Equal(t, "options/abc/img.jpg", img.MinioKey)
	assert.Equal(t, "image/jpeg", img.ContentType)
	require.NotNil(t, img.Width)
	assert.Equal(t, 800, *img.Width)
	assert.Equal(t, 0, img.SortOrder)
	assert.Equal(t, now, img.CreatedAt)
}

func TestScanImage_Error(t *testing.T) {
	ms := &mockScanner{err: errors.New("scan failed")}
	_, err := scanImage(ms)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan image")
}

func TestInsertParams_Defaults(t *testing.T) {
	p := InsertParams{
		OptionID:    uuid.New(),
		MinioKey:    "options/abc/test.jpg",
		ContentType: "image/jpeg",
		SourceURL:   "https://example.com/img.jpg",
		SortOrder:   0,
	}
	assert.Nil(t, p.Width)
	assert.Nil(t, p.Height)
	assert.Equal(t, 0, p.SortOrder)
}
