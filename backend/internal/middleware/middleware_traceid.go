package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader lets a browser or upstream proxy read the trace id of the
// request it just made.
const TraceIDHeader = "X-Trace-Id"

// TraceID writes the active span's trace id to the response header.
// otelhttp must run before this middleware so a span exists.
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if span := trace.SpanContextFromContext(r.Context()); span.IsValid() {
			w.Header().Set(TraceIDHeader, span.TraceID().String())
		}
		next.ServeHTTP(w, r)
	})
}
