/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package telemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.8.0"
	"google.golang.org/grpc"
)

// NewProvider registers a global tracer provider with a default OTLP exporter pointing to a collector.
// The global trace provider allows you to initialize a new trace from anywhere in code.
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

	// Create default otlp exporter
	exporter, err := otlptrace.New(
		ctx,
		otlpgrpc.NewClient(
			otlpgrpc.WithInsecure(),
			otlpgrpc.WithEndpoint(collectorEndpoint),
			otlpgrpc.WithDialOption(grpc.WithBlock()),
		),
	)

	traceOpts = append(traceOpts, trace.WithBatcher(exporter))

	// Create a new tracer provider with a batch span processor and the otlp exporters.
	tp = trace.NewTracerProvider(traceOpts...)
	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return
}

// CleanupTraceProvider calls the TraceProvider to shut down any span processors
func CleanupTraceProvider(ctx context.Context, tp *trace.TracerProvider) {
	if tp != nil {
		_ = tp.ForceFlush(ctx)
		_ = tp.Shutdown(ctx)
	}
}
