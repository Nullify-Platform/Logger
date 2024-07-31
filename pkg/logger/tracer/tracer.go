// Package tracer provides a way to get the tracer from the context and to create a new context with a tracer. It also provides a way to force the trace provider to flush all the traces to the exporter.
package tracer

import (
	"context"

	otelsdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type tracerCtxKey struct{}
type traceProviderCtxKey struct{}

// FromContext returns the tracer from the context
func FromContext(ctx context.Context) trace.Tracer {
	t, _ := ctx.Value(tracerCtxKey{}).(trace.Tracer)
	return t
}

// CopyFromContext copies the tracer from the old context to the new context
func CopyFromContext(fromCtx context.Context, toCtx context.Context) context.Context {
	t := fromCtx.Value(tracerCtxKey{})
	tp := fromCtx.Value(traceProviderCtxKey{})

	toCtx = context.WithValue(toCtx, tracerCtxKey{}, t)
	toCtx = context.WithValue(toCtx, traceProviderCtxKey{}, tp)

	return toCtx
}

// NewContext returns a new context with the given tracer
func NewContext(parent context.Context, tp *otelsdk.TracerProvider, tracerName string) context.Context {
	tracer := tp.Tracer(tracerName)
	ctx := context.WithValue(parent, traceProviderCtxKey{}, tp)
	return context.WithValue(ctx, tracerCtxKey{}, tracer)
}

// ForceFlush forces the trace provider to flush all the traces to the exporter
func ForceFlush(ctx context.Context) error {
	tp, _ := ctx.Value(traceProviderCtxKey{}).(*otelsdk.TracerProvider)
	return tp.ForceFlush(ctx)
}
