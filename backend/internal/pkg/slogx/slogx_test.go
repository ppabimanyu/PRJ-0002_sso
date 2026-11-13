package slogx

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestHandlerAddsTraceAndSpanIDs(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(NewHandler(slog.NewJSONHandler(&buffer, nil)))

	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	logger.InfoContext(ctx, "hello")

	var record map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &record))
	assert.Equal(t, traceID.String(), record["trace_id"])
	assert.Equal(t, spanID.String(), record["span_id"])
}

func TestHandlerLeavesRecordsWithoutSpanAlone(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(NewHandler(slog.NewJSONHandler(&buffer, nil)))

	logger.Info("hello")

	var record map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &record))
	assert.NotContains(t, record, "trace_id")
	assert.NotContains(t, record, "span_id")
}
