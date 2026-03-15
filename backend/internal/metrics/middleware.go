package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Middleware returns chi middleware that records HTTP metrics using the provided HTTPRecorder.
func Middleware(recorder HTTPRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			// Use chi route pattern to avoid cardinality explosion from UUIDs/slugs.
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}

			recorder.RecordHTTPRequest(
				r.Method,
				route,
				fmt.Sprintf("%d", ww.Status()),
				time.Since(start),
			)
		})
	}
}
