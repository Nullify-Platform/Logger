package logger

import (
	"context"
	"io"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Release:          Version,
		EnableTracing:    false,
		Debug:            true,
	})
	if err != nil {
		zap.L().Error("failed to initialise sentry", zap.Error(err))
		return
	}
}

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
	if err != nil {
		return nil, err
	}

	return traceExporter, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("TODO_PLACEHOLDER_SERVICE_NAME"),
			semconv.ServiceVersion(Version),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
}

// ConfigureProductionLogger configures a JSON production logger
func ConfigureProductionLogger(ctx context.Context, level string, syncs ...io.Writer) (context.Context, error) {
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

	traceExporter, err := newExporter(ctx)
	if err != nil {
		zap.L().Fatal("failed to create trace exporter", zap.Error(err))
	}

	tp := newTraceProvider(traceExporter)
	otel.SetTracerProvider(tp)

	l := &logger{underlyingLogger: zapLogger, tracer: tp.Tracer("logger-tracer")}
	ctx = l.WithContext(ctx)
	ctx = tracer.NewContext(ctx, tp.Tracer("logger-tracer"))

	return ctx, nil
}
