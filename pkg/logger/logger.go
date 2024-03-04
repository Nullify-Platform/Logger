// Package logger implements a logger interface with an implementation using the zap logger
package logger

import (
	"context"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface that for all the basic logging methods
// this package also provides a global implementation of the methods in this interface
type Logger interface {
	NewChild(fields ...Field) Logger
	WithOptions(opts ...Option) Logger

	AddField(fields ...Field)
	Sync()

	// levels
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// context
	WithContext(ctx context.Context) context.Context
	Tracer() trace.Tracer
}

type logger struct {
	Logger

	underlyingLogger *zap.Logger
	tracer           trace.Tracer
}

type loggerCtxKey struct{}

// NewContext returns a new context with the given logger
func (l *logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l.NewChild())
}

func (l *logger) Tracer() trace.Tracer {
	return l.tracer
}

// NewChild creates a new logger based on the default logger with the given default fields
func (l *logger) NewChild(fields ...Field) Logger {
	newLogger := l.underlyingLogger.With(fields...)
	return &logger{underlyingLogger: newLogger, tracer: l.tracer}
}

// WithOptions adds a new field to the default logger
func (l *logger) WithOptions(opts ...Option) Logger {
	newLogger := l.underlyingLogger.WithOptions(opts...)
	return &logger{underlyingLogger: newLogger, tracer: l.tracer}
}

// AddField adds a new field to the default logger
func (l *logger) AddFields(fields ...Field) {
	l.underlyingLogger = l.underlyingLogger.With(fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() {
	if os.Getenv("SENTRY_DSN") != "" {
		success := sentry.Flush(200 * time.Millisecond)

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
	captureExceptions(fields)
	l.underlyingLogger.Warn(msg, fields...)
}

// Error logs a message with the error level
func (l *logger) Error(msg string, fields ...Field) {
	captureExceptions(fields)
	l.underlyingLogger.Error(msg, fields...)
}

// Fatal logs a message with the fatal level and then calls os.Exit(1)
func (l *logger) Fatal(msg string, fields ...Field) {
	captureExceptions(fields)
	l.Sync()

	l.underlyingLogger.Fatal(msg, fields...)
}

// captureExceptions captures exceptions from fields and sends them to sentry
func captureExceptions(fields []Field) {
	if os.Getenv("SENTRY_DSN") == "" {
		return
	}

	for _, f := range fields {
		if f.Type != zapcore.ErrorType {
			continue
		}

		// cast the interface to an error
		err, ok := f.Interface.(error)
		if ok {
			sentry.CaptureException(err)
		}
	}
}
