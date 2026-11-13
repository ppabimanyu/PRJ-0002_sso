package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/ppabimanyu/sso/internal/config"
)

// Telemetry owns the providers that must flush during shutdown.
type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

// SetupTelemetry installs the global propagator and the trace and metric
// providers. When an OTLP base endpoint is configured, both export over OTLP
// at /v1/traces and /v1/metrics. Without one, spans still get valid trace ids
// for logs and response headers, and exports are simply discarded.
func SetupTelemetry(ctx context.Context, cfg config.Config) (*Telemetry, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("service.name", cfg.OTelServiceName),
		attribute.String("deployment.environment.name", cfg.AppEnv),
	)

	traceOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.OTelTraceSampleRatio))),
	}
	metricOptions := []sdkmetric.Option{
		sdkmetric.WithResource(res),
	}

	if cfg.OTelEndpoint != "" {
		// WithEndpointURL does not append the signal path, so build the full
		// per-signal URL from the base endpoint.
		baseURL := strings.TrimRight(cfg.OTelEndpoint, "/")

		traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
			otlptracehttp.WithEndpointURL(baseURL+"/v1/traces"),
		))
		if err != nil {
			return nil, fmt.Errorf("create trace exporter: %w", err)
		}
		traceOptions = append(traceOptions, sdktrace.WithBatcher(traceExporter))

		metricExporter, err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpointURL(baseURL+"/v1/metrics"),
		)
		if err != nil {
			return nil, fmt.Errorf("create metric exporter: %w", err)
		}
		metricOptions = append(metricOptions,
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(15*time.Second))),
		)
	}

	tracerProvider := sdktrace.NewTracerProvider(traceOptions...)
	otel.SetTracerProvider(tracerProvider)

	meterProvider := sdkmetric.NewMeterProvider(metricOptions...)
	otel.SetMeterProvider(meterProvider)

	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		_ = meterProvider.Shutdown(ctx)
		_ = tracerProvider.Shutdown(ctx)
		return nil, fmt.Errorf("start runtime metrics: %w", err)
	}

	return &Telemetry{tracerProvider: tracerProvider, meterProvider: meterProvider}, nil
}

// Shutdown flushes pending spans and metrics.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	if t.tracerProvider != nil {
		errs = append(errs, t.tracerProvider.Shutdown(ctx))
	}
	if t.meterProvider != nil {
		errs = append(errs, t.meterProvider.Shutdown(ctx))
	}
	return errors.Join(errs...)
}
