package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.8.0"
)

// NewProvider registers a global tracer provider with a default OTLP exporter pointing to a collector.
// The global trace provider allows you initialize a new trace from anywhere in code.
func NewProvider(
	ctx context.Context,
	serviceName string,
	collectorEndpoint string,
	sampler trace.Sampler,
) (tp *trace.TracerProvider, err error) {
	r, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing resource: %w", err)
	}

	traceOpts := []trace.TracerProviderOption{
		trace.WithSampler(sampler),
		trace.WithResource(r),
	}

	//traceOpts = append(traceOpts, trace.WithBatcher(exporter))

	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		fmt.Println("exp err", err)
	}

	traceOpts = append(traceOpts, trace.WithBatcher(exp))

	// Create a new tracer provider with a batch span processor and the otlp exporters.
	tp = trace.NewTracerProvider(traceOpts...)
	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return
}

// Cleanup calls the TraceProvider to shutdown any span processors
func Cleanup(ctx context.Context, tp *trace.TracerProvider) {
	if tp != nil {
		tp.ForceFlush(ctx)
		_ = tp.Shutdown(ctx)
	}
}
