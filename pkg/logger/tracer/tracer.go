package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type tracerCtxKey struct{}

// FromContext returns the tracer from the context
func FromContext(ctx context.Context) trace.Tracer {
	t, _ := ctx.Value(tracerCtxKey{}).(trace.Tracer)
	return t
}

// NewContext returns a new context with the given tracer
func NewContext(parent context.Context, t trace.Tracer) context.Context {
	return context.WithValue(parent, tracerCtxKey{}, t)
}
