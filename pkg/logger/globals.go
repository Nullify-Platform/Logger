package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// L returns the logger from the context
func L(ctx context.Context) Logger {
	if ctx == nil {
		return nil
	}

	var fields []zap.Field

	if l, ok := ctx.Value(loggerCtxKey{}).(Logger); ok {
		spanContext := trace.SpanFromContext(ctx).SpanContext()
		if traceID := spanContext.TraceID(); traceID.IsValid() {
			fields = append(fields, zap.String("trace-id", traceID.String()))
		}

		if spanID := spanContext.SpanID(); spanID.IsValid() {
			fields = append(fields, zap.String("span-id", spanID.String()))
		}

		l := l.NewChild(fields...)
		l.PassContext(ctx)

		return l
	}

	return nil
}

// CopyFromContext copies the logger from the old context to the new context
func CopyFromContext(fromCtx context.Context, toCtx context.Context) context.Context {
	l := fromCtx.Value(loggerCtxKey{})

	return context.WithValue(toCtx, loggerCtxKey{}, l)
}
