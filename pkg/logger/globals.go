package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// FromContext returns the logger from the context
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return nil
	}

	var fields []zap.Field

	if l, ok := ctx.Value(loggerCtxKey{}).(Logger); ok {
		if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
			fields = append(fields, zap.String("trace-id", traceID.String()))
		}

		if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
			fields = append(fields, zap.String("span-id", spanID.String()))
		}

		l := l.NewChild(fields...)
		l.PassContext(ctx)

		return l
	}

	return nil
}
