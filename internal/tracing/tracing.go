// Package tracing provides OpenTelemetry distributed tracing setup and middleware.
package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds tracing configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	ExporterEndpoint string // OTLP gRPC endpoint (e.g., "localhost:4317")
	Enabled        bool
}

// ShutdownFunc is returned by Init and should be called to flush traces on shutdown.
type ShutdownFunc func(ctx context.Context) error

// Init initializes OpenTelemetry tracing with OTLP gRPC exporter.
// Returns a shutdown function that should be deferred.
func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	if !cfg.Enabled {
		slog.Info("distributed tracing disabled")
		return func(ctx context.Context) error { return nil }, nil
	}

	if cfg.ExporterEndpoint == "" {
		cfg.ExporterEndpoint = "localhost:4317"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
		otlptracegrpc.WithInsecure(), // Use insecure for local dev; production should use TLS
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("distributed tracing initialized",
		"endpoint", cfg.ExporterEndpoint,
		"service", cfg.ServiceName,
	)

	return func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	}, nil
}

// Tracer returns a named tracer for the given instrumentation scope.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
