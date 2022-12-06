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

// NewHandler creates a tracer
func NewHandler() graphql.HandlerExtension {
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
