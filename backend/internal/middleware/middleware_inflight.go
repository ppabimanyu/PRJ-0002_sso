package middleware

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// InFlight records the number of requests currently being served. The
// instrument name matches the OpenTelemetry HTTP server semantic convention.
func InFlight(scope string) func(http.Handler) http.Handler {
	counter, err := otel.Meter(scope).Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of requests currently being served."),
	)
	if err != nil {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			counter.Add(r.Context(), 1)
			defer counter.Add(context.Background(), -1)
			next.ServeHTTP(w, r)
		})
	}
}
