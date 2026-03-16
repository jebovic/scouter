package imagefetch

import (
	"time"

	"github.com/google/uuid"
)

// OptionImage is a stored image record from the DB.
type OptionImage struct {
	ID          uuid.UUID `json:"id"`
	OptionID    uuid.UUID `json:"optionId"`
	MinioKey    string    `json:"-"`
	ContentType string    `json:"contentType"`
	Width       *int      `json:"width,omitempty"`
	Height      *int      `json:"height,omitempty"`
	SourceURL   string    `json:"-"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	// URL is populated by the handler with a presigned URL — not stored in DB.
	URL string `json:"url,omitempty"`
}

// FetchedImage holds raw bytes from a successful scrape before MinIO upload.
type FetchedImage struct {
	Bytes       []byte
	ContentType string
	Width       *int
	Height      *int
	SourceURL   string
}
