// Package slogx adds trace correlation to structured logs.
package slogx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// NewHandler wraps inner so every record logged with a context that carries a
// valid span gets trace_id and span_id attributes.
func NewHandler(inner slog.Handler) slog.Handler {
	return &traceHandler{inner: inner}
}

type traceHandler struct {
	inner slog.Handler
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", span.TraceID().String()),
			slog.String("span_id", span.SpanID().String()),
		)
	}
	return h.inner.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{inner: h.inner.WithGroup(name)}
}
