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
	"github.com/nullify-platform/logger/pkg/logger/meter"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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

	ctx, err = configureOTel(ctx, "dev-logger")
	if err != nil {
		return nil, err
	}

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
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

	ctx, err = configureOTel(ctx, "prod-logger")
	if err != nil {
		return nil, err
	}

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	return ctx, nil
}

// configureOTel configures the OTel tracer and meter providers, returning a new context with providers attached.
func configureOTel(ctx context.Context, scopeName string) (context.Context, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceVersion(Version),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, err
	}

	headers := resolveOTLPHeaders(ctx)

	traceExporter, err := newSpanExporter(ctx, headers)
	if err != nil {
		zap.L().Error("failed to create trace exporter, continuing", zap.Error(err))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx = tracer.NewContext(ctx, tp, scopeName+"-tracer")

	metricExporter, err := newMetricExporter(ctx, headers)
	if err != nil {
		zap.L().Error("failed to create metric exporter, continuing", zap.Error(err))
	}

	mpOpts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if metricExporter != nil {
		mpOpts = append(mpOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
	}
	mp := sdkmetric.NewMeterProvider(mpOpts...)
	otel.SetMeterProvider(mp)
	ctx = meter.NewContext(ctx, mp, scopeName+"-meter")

	return ctx, nil
}

// resolveOTLPHeaders fetches OTLP exporter headers once from SSM parameter store.
// Returns nil if no endpoint is configured or no headers are needed.
func resolveOTLPHeaders(ctx context.Context) map[string]string {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return nil
	}

	raw := getSecretFromParamStore(ctx, "OTEL_EXPORTER_OTLP_HEADERS_NAME")
	if raw == nil {
		return nil
	}

	headerMap := make(map[string]string)
	for header := range strings.SplitSeq(*raw, ",") {
		parts := strings.SplitN(header, "=", 2)
		if len(parts) != 2 {
			zap.L().Error("invalid header format", zap.String("header", header))
			continue
		}
		headerMap[parts[0]] = parts[1]
	}
	return headerMap
}

func newSpanExporter(ctx context.Context, headers map[string]string) (sdktrace.SpanExporter, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		if headers != nil {
			return otlptracehttp.New(ctx, otlptracehttp.WithHeaders(headers))
		}
		return otlptracehttp.New(ctx)
	}

	if os.Getenv("TRACE_OUTPUT_DEBUG") != "" {
		return stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
	}

	return nil, nil
}

func newMetricExporter(ctx context.Context, headers map[string]string) (sdkmetric.Exporter, error) {
	// Grafana Cloud (Mimir) requires cumulative temporality for all metric types.
	cumulativeTemporality := otlpmetrichttp.WithTemporalitySelector(
		func(sdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.CumulativeTemporality
		},
	)

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		if headers != nil {
			return otlpmetrichttp.New(ctx, otlpmetrichttp.WithHeaders(headers), cumulativeTemporality)
		}
		return otlpmetrichttp.New(ctx, cumulativeTemporality)
	}

	if os.Getenv("TRACE_OUTPUT_DEBUG") != "" {
		return stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	}

	return nil, nil
}

func getSecretFromParamStore(ctx context.Context, varName string) *string {
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
