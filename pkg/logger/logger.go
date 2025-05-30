// Package logger implements a logger interface with an implementation using the zap logger
package logger

import (
	"context"
	"errors"

	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
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

	SetSpanAttributes(spanName string) context.Context
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

	_ = l.underlyingLogger.Sync()
}

// Debug logs a message with the debug level
func (l *logger) Debug(msg string, fields ...Field) {
	updateFields := l.getContextMetadataAsFields(fields)
	l.underlyingLogger.Debug(msg, updateFields...)
}

// Info logs a message with the info level
func (l *logger) Info(msg string, fields ...Field) {
	updateFields := l.getContextMetadataAsFields(fields)
	l.underlyingLogger.Info(msg, updateFields...)
}

// Warn logs a message with the warn level
func (l *logger) Warn(msg string, fields ...Field) {
	updateFields := l.getContextMetadataAsFields(fields)
	l.underlyingLogger.Warn(msg, updateFields...)
}

// Error logs a message with the error level
func (l *logger) Error(msg string, fields ...Field) {
	trace.SpanFromContext(l.attachedContext).RecordError(errors.New(msg))
	trace.SpanFromContext(l.attachedContext).SetStatus(codes.Error, msg)
	updateFields := l.getContextMetadataAsFields(fields)
	l.underlyingLogger.Error(msg, updateFields...)
}

// Fatal logs a message with the fatal level and then calls os.Exit(1)
func (l *logger) Fatal(msg string, fields ...Field) {
	trace.SpanFromContext(l.attachedContext).SetStatus(codes.Error, msg)
	updateFields := l.getContextMetadataAsFields(fields)
	l.Sync()

	l.underlyingLogger.Fatal(msg, updateFields...)
}
