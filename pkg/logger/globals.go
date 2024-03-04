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

	if l, ok := ctx.Value(loggerCtxKey{}).(Logger); ok {
		if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
			l = l.NewChild(zap.String("trace-id", traceID.String()))
		}

		if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
			l = l.NewChild(zap.String("span-id", spanID.String()))
		}

		return l
	}

	return nil
}

// NewChild creates a new logger based on the default logger with the given default fields
func NewChild(fields ...Field) Logger {
	return &logger{underlyingLogger: zap.L().With(fields...)}
}

// AddField adds a new field to the default logger
func AddField(fields ...Field) {
	zap.ReplaceGlobals(zap.L().With(fields...))
}

// levels

// Debug logs a message with the debug level
func Debug(msg string, fields ...Field) {
	zap.L().Debug(msg, fields...)
}

// Info logs a message with the info level
func Info(msg string, fields ...Field) {
	zap.L().Info(msg, fields...)
}

// Warn logs a message with the warn level
func Warn(msg string, fields ...Field) {
	captureExceptions(fields)
	zap.L().Warn(msg, fields...)
}

// Error logs a message with the error level
func Error(msg string, fields ...Field) {
	captureExceptions(fields)
	zap.L().Error(msg, fields...)
}

// Fatal logs a message with the fatal level and then calls os.Exit(1)
func Fatal(msg string, fields ...Field) {
	captureExceptions(fields)
	zap.L().Fatal(msg, fields...)
}
