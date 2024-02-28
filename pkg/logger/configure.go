package logger

import (
	"io"
	"os"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Version is the current version of the application
// override this value with ldflags
// e.g. -ldflags "-X 'github.com/nullify-platform/logger/pkg/logger.Version=$(VERSION)'"
var Version = "0.0.0"

// ConfigureDevelopmentLogger configures a development logger which is more human readable instead of JSON
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

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(sync),
			zapLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("version", Version)),
	)
	zap.ReplaceGlobals(zapLogger)
	return &logger{underlyingLogger: zapLogger}, nil
}

func initialiseSentry() {
	if os.Getenv("SENTRY_DSN") == "" {
		zap.L().Fatal("Sentry DSN has not been configured; not capturing errors")
		return
	} else {
		zap.L().Info("Sentry DSN has been configured; capturing errors")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Release:          Version,
		EnableTracing:    false,
		Debug:            true,
	})
	if err != nil {
		zap.L().Fatal("failed to initialise sentry", zap.Error(err))
		return
	}
}

// ConfigureProductionLogger configures a JSON production logger
func ConfigureProductionLogger(level string, syncs ...io.Writer) (Logger, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Fatal("failed to parse log level")
	}

	var sync io.Writer = os.Stdout
	if len(syncs) > 0 {
		sync = syncs[0]
	}

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(sync),
			zapLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("version", Version)),
	)
	zap.ReplaceGlobals(zapLogger)

	initialiseSentry()

	return &logger{underlyingLogger: zapLogger}, nil
}
