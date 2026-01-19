// Package tracer provides a way to get the tracer from the context and to create a new context with a tracer. It also provides a way to force the trace provider to flush all the traces to the exporter.
package tracer

import (
	"context"
	"time"

	otelsdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type (
	tracerCtxKey        struct{}
	traceProviderCtxKey struct{}
)

// FromContext returns the tracer from the context
func FromContext(ctx context.Context) trace.Tracer {
	t, _ := ctx.Value(tracerCtxKey{}).(trace.Tracer)
	return t
}

// StartNewSpan loads the Tracer from the context and starts a new span.
// opts can be used to provide additional options to t.Start():
//   - trace.WithAttributes()
//   - trace.WithSpanKind(trace.SpanKindServer)  or Internal/Client/Producer/Consumer
//   - trace.WithLinks({SpanContext, Attributes})
//   - trace.WithStackTrace(true)
func StartNewSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	t := FromContext(ctx)
	return t.Start(ctx, spanName, opts...)
}

// StartNewRootSpan should be called at from API handlers to start a new root span for each request.
// opts can be used to provide additional options to t.Start():
//   - trace.WithAttributes()
//   - trace.WithSpanKind(trace.SpanKindServer)  or Internal/Client/Producer/Consumer
//   - trace.WithLinks({SpanContext, Attributes})
//   - trace.WithStackTrace(true)
func StartNewRootSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return StartNewSpan(ctx, spanName, append(opts, trace.WithNewRoot())...)
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

// ForceFlushWithReplacedTimeout forces the trace provider to flush all the traces to the exporter, replacing any timeout on ctx with a new one.
// This is useful so that if ctx is cancelled, we still have a grace period to complete the flush.
func ForceFlushWithReplacedTimeout(ctx context.Context, timeout time.Duration) error {
	flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	return ForceFlush(flushCtx)
}
