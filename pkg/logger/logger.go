// Package logger implements a logger interface with an implementation using the zap logger
package logger

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
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
}

type logger struct {
	Logger

	underlyingLogger *zap.Logger
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

// AddField adds a new field to the default logger
func (l *logger) AddFields(fields ...Field) {
	l.underlyingLogger = l.underlyingLogger.With(fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() {
	_ = l.underlyingLogger.Sync()
	sentry.Flush(2 * time.Second)
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
	l.underlyingLogger.Warn(msg, fields...)
}

// Error logs a message with the error level
func (l *logger) Error(msg string, fields ...Field) {
	l.underlyingLogger.Error(msg, fields...)
}

// Fatal logs a message with the fatal level and then calls os.Exit(1)
func (l *logger) Fatal(msg string, fields ...Field) {
	l.underlyingLogger.Fatal(msg, fields...)
}
