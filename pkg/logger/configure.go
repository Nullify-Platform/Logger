package logger

import (
	"context"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Version is the current version of the application
// override this value with ldflags
// e.g. -ldflags "-X 'github.com/nullify-platform/logger/pkg/logger.Version=$(VERSION)'"
var Version = ""

// BuildInfoRevision https://icinga.com/blog/2022/05/25/embedding-git-commit-information-in-go-binaries/
var BuildInfoRevision = func() string {
	var revision, tainted string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" { //  The information is available only in binaries built with module support.
				revision = setting.Value
			}
			if setting.Key == "vcs.modified" { // if the source code has been modified since the last commit
				tainted = "-tainted"
			}
		}
	}
	if revision == "" {
		return "0.0.0" // Fallback if no revision is found
	}
	return revision + tainted
}()

// ConfigureDevelopmentLogger configures a development logger which is more human readable instead of JSON
func ConfigureDevelopmentLogger(ctx context.Context, level string, syncs ...io.Writer) (context.Context, error) {
	// configure level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Error("failed to parse log level, using info", zap.Error(err))
		zapLevel = zapcore.InfoLevel
	}

	var writers []io.Writer
	if len(syncs) > 0 {
		writers = syncs
	} else {
		writers = []io.Writer{os.Stdout}
	}

	// Convert io.Writers to zapcore.WriteSyncers
	writeSyncers := make([]zapcore.WriteSyncer, len(writers))
	for i, writer := range writers {
		writeSyncers[i] = zapcore.AddSync(writer)
	}

	// Combine multiple syncs into a single WriteSyncer
	multiSync := zapcore.NewMultiWriteSyncer(writeSyncers...)

	version := Version
	if version == "" {
		version = BuildInfoRevision
	}

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			multiSync,
			zapLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("version", version)),
	)
	zap.ReplaceGlobals(zapLogger)

	traceExporter, err := newExporter(ctx)
	if err != nil {
		zap.L().Error("failed to create trace exporter, continuing...", zap.Error(err))
	}

	tp, err := newTraceProvider(ctx, traceExporter)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	ctx = tracer.NewContext(ctx, tp, "dev-logger-tracer")
	return ctx, nil
}

// ConfigureProductionLogger configures a JSON production logger
func ConfigureProductionLogger(ctx context.Context, level string, syncs ...io.Writer) (context.Context, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Error("failed to parse log level, using info", zap.Error(err))
		zapLevel = zapcore.InfoLevel
	}

	var writers []io.Writer
	if len(syncs) > 0 {
		writers = syncs
	} else {
		writers = []io.Writer{os.Stdout}
	}

	// Convert io.Writers to zapcore.WriteSyncers
	writeSyncers := make([]zapcore.WriteSyncer, len(writers))
	for i, writer := range writers {
		writeSyncers[i] = zapcore.AddSync(writer)
	}

	// Combine multiple syncs into a single WriteSyncer
	multiSync := zapcore.NewMultiWriteSyncer(writeSyncers...)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	version := Version
	if version == "" {
		version = BuildInfoRevision
	}

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			multiSync,
			zapLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("version", version)),
	)
	zap.ReplaceGlobals(zapLogger)

	traceExporter, err := newExporter(ctx)
	if err != nil {
		zap.L().Error("failed to create trace exporter, continuing", zap.Error(err))
	}

	tp, err := newTraceProvider(ctx, traceExporter)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	ctx = tracer.NewContext(ctx, tp, "prod-logger-tracer")

	return ctx, nil
}

func newTraceProvider(ctx context.Context, exp sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	// Create a resource with our own attributes to avoid schema conflicts
	r, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceVersion(Version),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	), nil
}

func getSecretFromParamStore(ctx context.Context, varName string) *string {
	// check if the param name is defined in the environment
	paramName := os.Getenv(varName)
	if paramName == "" {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		zap.L().Error("failed to load aws config", zap.Error(err), zap.String("paramName", paramName))
		return nil
	}

	svc := ssm.NewFromConfig(cfg)
	param, err := svc.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		zap.L().Error("failed to fetch parameter", zap.Error(err), zap.String("paramName", paramName))
		return nil
	}

	return param.Parameter.Value
}

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		headers := getSecretFromParamStore(ctx, "OTEL_EXPORTER_OTLP_HEADERS_NAME")
		if headers == nil {
			traceExporter, err := otlptracehttp.New(ctx)
			if err != nil {
				return nil, err
			}

			return traceExporter, nil
		}

		var headerMap = make(map[string]string)
		for _, header := range strings.Split(*headers, ",") {
			parts := strings.SplitN(header, "=", 2)
			if len(parts) != 2 {
				zap.L().Error("invalid header format", zap.String("header", header))
				continue
			}

			headerMap[parts[0]] = parts[1]
		}

		traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithHeaders(headerMap))
		if err != nil {
			return nil, err
		}

		return traceExporter, nil
	}

	if os.Getenv("TRACE_OUTPUT_DEBUG") != "" {
		traceExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
		if err != nil {
			return nil, err
		}
		return traceExporter, nil
	}

	return nil, nil
}
