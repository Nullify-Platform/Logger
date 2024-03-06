package tracer

import (
	"context"
	"go.opentelemetry.io/otel/trace"

	otelsdk "go.opentelemetry.io/otel/sdk/trace"
)

type tracerCtxKey struct{}
type traceProviderCtxKey struct{}

// FromContext returns the tracer from the context
func FromContext(ctx context.Context) trace.Tracer {
	t, _ := ctx.Value(tracerCtxKey{}).(trace.Tracer)
	return t
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
