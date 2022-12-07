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

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
	graphql.FieldInterceptor
} = GQLGenTracer{}

// GQLGenTracer must implement HandlerExtension, ResponseInterceptor and FieldInterceptor
type GQLGenTracer struct{}

// NewTracer creates a tracer
func NewTracer() graphql.HandlerExtension {
	return GQLGenTracer{}
}

func (GQLGenTracer) ExtensionName() string {
	return "GQLGenTracer"
}

func (GQLGenTracer) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (GQLGenTracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	ctx = startFieldExecution(ctx)
	defer endFieldExecution(ctx)

	return next(ctx)
}

func (GQLGenTracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {

	ctx = startOperationExecution(ctx)
	defer endOperationExecution(ctx)

	return next(ctx)
}

func startOperationExecution(ctx context.Context) context.Context {
	opCtx := graphql.GetOperationContext(ctx)

	ctx, span := otel.Tracer("").
		Start(
			ctx,
			operationName(opCtx.OperationName),
			trace.WithSpanKind(trace.SpanKindServer),
		)

	if !span.IsRecording() {
		return ctx
	}

	for key, val := range opCtx.Variables {
		span.SetAttributes(
			attribute.String(fmt.Sprintf("graphql.request.variables.%s", key), fmt.Sprintf("%+v", val)),
		)
	}

	return ctx
}

func endOperationExecution(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.End()
	}
}

func startFieldExecution(ctx context.Context) context.Context {
	fieldCtx := graphql.GetFieldContext(ctx)
	field := fieldCtx.Field
	if len(field.Arguments) == 0 {
		return ctx
	}

	ctx, span := otel.Tracer("").Start(
		ctx,
		fmt.Sprintf("%s/%s", field.ObjectDefinition.Name, field.Name),
		trace.WithSpanKind(trace.SpanKindServer),
	)

	if !span.IsRecording() {
		return ctx
	}

	span.SetAttributes(
		attribute.String("graphql.resolver.object", field.ObjectDefinition.Name),
		attribute.String("graphql.resolver.field", field.Name),
	)

	for _, arg := range field.Arguments {
		if arg.Value != nil {
			for _, arg := range field.Arguments {
				span.SetAttributes(
					attribute.String(fmt.Sprintf("graphql.resolver.args.%s", arg.Name), fmt.Sprintf("%+v", arg.Value)),
				)
			}
		}
	}

	return ctx
}

func endFieldExecution(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.End()
	}
}

// operationName in case there is no operation name, use "nameless-op"
func operationName(name string) string {
	nameless := "nameless-op"

	if name != "" {
		return name
	}

	return nameless
}
