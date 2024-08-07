// Package logger implements a logger interface with an implementation using the zap logger
package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface that for all the basic logging methods
// this package also provides a global implementation of the methods in this interface
type Logger interface {
	NewChild(fields ...Field) Logger
	WithOptions(opts ...Option) Logger

	AddFields(fields ...Field)
	Sync()

	// levels
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// context
	InjectIntoContext(ctx context.Context) context.Context
	PassContext(ctx context.Context)
}

type logger struct {
	Logger

	underlyingLogger *zap.Logger
	attachedContext  context.Context
}

type loggerCtxKey struct{}

// InjectIntoContext injects the logger into the context
func (l *logger) InjectIntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l.NewChild())
}

// PassContext passes the context to the logger
func (l *logger) PassContext(ctx context.Context) {
	l.attachedContext = ctx
}

// NewChild creates a new logger based on the default logger with the given default fields
func (l *logger) NewChild(fields ...Field) Logger {
	newLogger := l.underlyingLogger.With(fields...)
	return &logger{underlyingLogger: newLogger}
}

// WithOptions adds a new field to the default logger
func (l *logger) WithOptions(opts ...Option) Logger {
	newLogger := l.underlyingLogger.WithOptions(opts...)
	return &logger{underlyingLogger: newLogger}
}

// AddFields adds new fields to the default logger
func (l *logger) AddFields(fields ...Field) {
	l.underlyingLogger = l.underlyingLogger.With(fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() {
	err := tracer.ForceFlush(l.attachedContext)
	if err != nil {
		l.Warn("tracer.ForceFlush failed", Err(err))
	}

	if os.Getenv("SENTRY_DSN") != "" {
		success := sentry.Flush(1000 * time.Millisecond)

		if !success {
			l.Error("sentry.Flush failed")
		}
	}

	_ = l.underlyingLogger.Sync()
}

// Debug logs a message with the debug level
func (l *logger) Debug(msg string, fields ...Field) {
	l.underlyingLogger.Debug(msg, fields...)
}

// Info logs a message with the info level
func (l *logger) Info(msg string, fields ...Field) {
	l.underlyingLogger.Info(msg, fields...)
}

// Warn logs a message with the warn level
func (l *logger) Warn(msg string, fields ...Field) {
	l.captureExceptions(fields)
	l.underlyingLogger.Warn(msg, fields...)
}

// Error logs a message with the error level
func (l *logger) Error(msg string, fields ...Field) {
	trace.SpanFromContext(l.attachedContext).SetStatus(codes.Error, msg)
	l.captureExceptions(fields)
	l.underlyingLogger.Error(msg, fields...)
}

// Fatal logs a message with the fatal level and then calls os.Exit(1)
func (l *logger) Fatal(msg string, fields ...Field) {
	trace.SpanFromContext(l.attachedContext).SetStatus(codes.Error, msg)
	l.captureExceptions(fields)
	l.Sync()

	l.underlyingLogger.Fatal(msg, fields...)
}

// captureExceptions captures exceptions from fields and sends them to sentry
func (l *logger) captureExceptions(fields []Field) {
	if os.Getenv("SENTRY_DSN") == "" {
		return
	}

	for _, f := range fields {
		if f.Type != zapcore.ErrorType {
			continue
		}

		// cast the interface to an error
		err, ok := f.Interface.(error)
		if !ok {
			continue
		}

		span := trace.SpanFromContext(l.attachedContext)
		span.RecordError(err, trace.WithStackTrace(true))

		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}

		// provide trace context to sentry
		sentry.WithScope(func(scope *sentry.Scope) {
			if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
				scope.SetContext("aws", map[string]interface{}{
					"lambda": os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
					"logsURL": formatLogsURL(
						region,
						os.Getenv("AWS_LAMBDA_LOG_GROUP_NAME"),
						os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME"),
					),
				})
			} else if os.Getenv("ECS_SERVICE_NAME") != "" {
				scope.SetContext("aws", map[string]interface{}{
					"ecs": os.Getenv("ECS_SERVICE_NAME"),
					"logsURL": formatLogsURL(
						region,
						os.Getenv("AWS_ECS_LOG_GROUP_NAME"),
						os.Getenv("AWS_ECS_LOG_STREAM_NAME"),
					),
				})
			}

			scope.SetContext("trace", map[string]interface{}{
				"traceID": span.SpanContext().TraceID().String(),
				"spanID":  span.SpanContext().SpanID().String(),
			})

			sentry.CaptureException(err)
		})
	}
}

func formatLogsURL(region string, logGroupName string, logStreamName string) string {
	logGroupName = strings.ReplaceAll(logGroupName, "/", "$252F")

	logStreamName = strings.ReplaceAll(logStreamName, "$", "$2524")
	logStreamName = strings.ReplaceAll(logStreamName, "/", "$252F")
	logStreamName = strings.ReplaceAll(logStreamName, "[", "$255B")
	logStreamName = strings.ReplaceAll(logStreamName, "]", "$255D")

	return fmt.Sprintf("https://%s.console.aws.amazon.com/cloudwatch/home?region=%s#logsV2:log-groups/log-group/%s/log-events/%s",
		region, region, logGroupName, logStreamName,
	)
}
