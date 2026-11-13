package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog records one structured line per request. The slogx handler adds
// trace_id and span_id when the request context carries a span.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			writer := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r)

			status := writer.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.InfoContext(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
				"bytes", writer.bytes,
				"request_id", RequestIDFromContext(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
