package logger

import (
	"go.uber.org/zap"
)

type Logger interface {
	NewChild(fields ...Field) Logger
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

// AddField adds a new field to the default logger
func (l *logger) AddFields(fields ...Field) {
	l.underlyingLogger = l.underlyingLogger.With(fields...)
}

func (l *logger) Sync() {
	_ = l.underlyingLogger.Sync()
}

func (l *logger) Debug(msg string, fields ...Field) {
	l.underlyingLogger.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	l.underlyingLogger.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.underlyingLogger.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.underlyingLogger.Error(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.underlyingLogger.Fatal(msg, fields...)
}
