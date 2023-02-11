package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Version string = "0.0.0"

type Logger interface {
	Sync()
}

type syncLogger struct {
	Logger
	SyncImplementation func()
}

func (logger syncLogger) Sync() {
	logger.SyncImplementation()
}

func ConfigureDevelopmentLogger(level string) (Logger, error) {
	// configure level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Fatal("failed to parse log level")
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	), zap.AddCallerSkip(1), zap.Fields(zap.String("version", Version)))
	zap.ReplaceGlobals(logger)
	return syncLogger{SyncImplementation: func() { _ = logger.Sync() }}, nil
}

func ConfigureProductionLogger(level string) (Logger, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Fatal("failed to parse log level")
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	), zap.AddCallerSkip(1), zap.Fields(zap.String("version", Version)))
	zap.ReplaceGlobals(logger)
	return syncLogger{SyncImplementation: func() { _ = logger.Sync() }}, nil
}

func Debug(msg string, fields ...zapcore.Field) {
	zap.L().Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	zap.L().Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	zap.L().Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	zap.L().Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	zap.L().Fatal(msg, fields...)
}

func NewChild(fields ...zapcore.Field) *zap.Logger {
	return zap.L().With(fields...)
}

// Fields

func Trace(trace []byte) zapcore.Field {
	return zap.ByteString("trace", trace)
}

func Err(err error) zapcore.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zapcore.Field {
	return zap.Any(key, val)
}

func String(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zapcore.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zapcore.Field {
	return zap.Int64(key, val)
}

func Duration(key string, val time.Duration) zapcore.Field {
	return zap.Duration(key, val)
}
