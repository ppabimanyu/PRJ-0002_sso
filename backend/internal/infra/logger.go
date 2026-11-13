// Package infra builds the process-level clients: logger, database, Redis,
// and OpenTelemetry providers.
package infra

import (
	"log/slog"
	"os"

	"github.com/ppabimanyu/sso/internal/pkg/slogx"
)

// NewLogger returns a JSON logger whose records carry trace_id and span_id
// whenever they are logged with a request context.
func NewLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(slogx.NewHandler(handler))
}
