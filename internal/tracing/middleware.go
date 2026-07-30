package tracing

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware returns middleware that automatically creates spans for HTTP requests.
// It extracts trace context from incoming requests and propagates it to outgoing requests.
func HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http.server",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
			otelhttp.WithPropagators(nil), // Use global propagators
		)
	}
}

// AddSpanAttributes adds custom attributes to the current span.
// Call this from within a handler to add request-specific metadata.
func AddSpanAttributes(r *http.Request, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attrs...)
}

// WithClientIP adds the client IP attribute to the current span.
func WithClientIP(r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	AddSpanAttributes(r, attribute.String("http.client_ip", ip))
}
