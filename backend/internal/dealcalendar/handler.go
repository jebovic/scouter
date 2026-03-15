package dealcalendar

import (
	"net/http"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves deal calendar endpoints.
type Handler struct {
	cal *Calendar
}

// NewHandler creates a new deal calendar handler with the static French calendar.
func NewHandler() *Handler {
	return &Handler{
		cal: NewCalendar(),
	}
}

// ListResponse is the response structure for GET /api/deal-calendar.
type ListResponse struct {
	Events []DealEvent `json:"events"`
	Count  int         `json:"count"`
}

// List handles GET /api/deal-calendar.
// Query parameters:
//   - category: filter by category hint (e.g., "electronics", "fashion", "all")
//   - upcoming: if "true", only return events starting today or later
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	upcomingRaw := r.URL.Query().Get("upcoming")
	upcoming := upcomingRaw == "true"

	// Start with all events
	events := h.cal.GetAll()

	// Filter by category if specified
	if category != "" {
		events = h.cal.FilterByCategory(events, category)
	}

	// Filter by upcoming if requested
	if upcoming {
		now := time.Now()
		events = h.cal.GetUpcoming(now)
		// Re-apply category filter if needed
		if category != "" {
			events = h.cal.FilterByCategory(events, category)
		}
	}

	resp := ListResponse{
		Events: events,
		Count:  len(events),
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
