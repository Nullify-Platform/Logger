package logger

import (
	"go.uber.org/zap"
)

// NewChild creates a new logger based on the default logger with the given default fields
func NewChild(fields ...Field) Logger {
	return &logger{underlyingLogger: zap.L().With(fields...)}
}

// AddField adds a new field to the default logger
func AddField(fields ...Field) {
	zap.ReplaceGlobals(zap.L().With(fields...))
}

// levels

func Debug(msg string, fields ...Field) {
	zap.L().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	zap.L().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	zap.L().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	zap.L().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	zap.L().Fatal(msg, fields...)
}
