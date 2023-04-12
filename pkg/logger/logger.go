package logger

import (
	"io"
	"os"

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

func ConfigureDevelopmentLogger(level string, syncs ...io.Writer) (Logger, error) {
	// configure level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Fatal("failed to parse log level")
	}

	var sync io.Writer = os.Stdout
	if len(syncs) > 0 {
		sync = syncs[0]
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(sync),
		zapLevel,
	), zap.AddCaller(), zap.AddCallerSkip(1), zap.Fields(zap.String("version", Version)))
	zap.ReplaceGlobals(logger)
	return syncLogger{SyncImplementation: func() { _ = logger.Sync() }}, nil
}

func ConfigureProductionLogger(level string, syncs ...io.Writer) (Logger, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Fatal("failed to parse log level")
	}

	var sync io.Writer = os.Stdout
	if len(syncs) > 0 {
		sync = syncs[0]
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(sync),
		zapLevel,
	), zap.AddCaller(), zap.AddCallerSkip(1), zap.Fields(zap.String("version", Version)))
	zap.ReplaceGlobals(logger)
	return syncLogger{SyncImplementation: func() { _ = logger.Sync() }}, nil
}

// NewChild creates a new logger based on the default logger with the given default fields
func NewChild(fields ...zapcore.Field) *zap.Logger {
	return zap.L().With(fields...)
}

// AddField adds a new field to the default logger
func AddField(fields ...zapcore.Field) {
	zap.ReplaceGlobals(NewChild(fields...))
}
