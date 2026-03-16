package httputil

import (
	"net/http"
	"strconv"
	"time"
)

const defaultLimit = 20
const maxLimit = 100

// PageParams contains parsed cursor pagination parameters.
type PageParams struct {
	Cursor *time.Time // nil means first page
	Limit  int
}

// ParsePageParams reads ?cursor=<RFC3339>&limit=<n> from the request.
// Cursor is an RFC3339 timestamp. Limit defaults to 20, max 100.
func ParsePageParams(r *http.Request) PageParams {
	p := PageParams{Limit: defaultLimit}

	if raw := r.URL.Query().Get("cursor"); raw != "" {
		if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			p.Cursor = &t
		}
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			p.Limit = n
		}
	}

	return p
}

// PagedResponse wraps a list with pagination metadata.
type PagedResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
}

// BuildPagedResponse trims the extra probe item and sets HasMore/NextCursor.
// items must come from a query that fetched limit+1 rows.
// cursorOf extracts the RFC3339Nano cursor string from the last visible item.
func BuildPagedResponse[T any](items []T, limit int, cursorOf func(T) string) PagedResponse[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	if items == nil {
		items = []T{}
	}
	resp := PagedResponse[T]{Items: items, HasMore: hasMore}
	if hasMore && len(items) > 0 {
		resp.NextCursor = cursorOf(items[len(items)-1])
	}
	return resp
}
