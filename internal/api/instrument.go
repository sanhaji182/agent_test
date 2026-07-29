package api

import (
	"net/http"

	"github.com/go-go-golems/gotest-agent/internal/appmetrics"
)

// statusRecorder wraps http.ResponseWriter to capture the status code written.
// Used by instrumentMiddleware to distinguish successful vs error responses.
type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

// instrumentMiddleware counts API requests and error responses for Prometheus metrics.
// Every request increments gotest_api_requests_total; responses with status >= 400
// increment gotest_api_errors_total.
func instrumentMiddleware(m *appmetrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.APIRequests.Add(1)
			rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.code >= 400 {
				m.APIErrors.Add(1)
			}
		})
	}
}

// handlePrometheus serves application metrics in Prometheus text exposition format.
// GET /metrics (outside auth — scrape endpoint)
func (s *Server) handlePrometheus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(s.metrics.Render()))
}
