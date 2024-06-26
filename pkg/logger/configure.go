package logger

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Version is the current version of the application
// override this value with ldflags
// e.g. -ldflags "-X 'github.com/nullify-platform/logger/pkg/logger.Version=$(VERSION)'"
var Version = "0.0.0"

// ConfigureDevelopmentLogger configures a development logger which is more human readable instead of JSON
func ConfigureDevelopmentLogger(ctx context.Context, level string, syncs ...io.Writer) (context.Context, error) {
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

	traceExporter, err := newExporter(ctx)
	if err != nil {
		zap.L().Fatal("failed to create trace exporter", zap.Error(err))
	}

	tp := newTraceProvider(traceExporter)
	otel.SetTracerProvider(tp)
	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	ctx = tracer.NewContext(ctx, tp, "dev-logger-tracer")

	return ctx, nil
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

func getSecretFromParamStore(varName string) *string {
	// Check if the param name is defined in the environment
	paramName := os.Getenv(varName)
	if len(paramName) == 0 {
		zap.L().Error("param name not defined in environment", zap.String("varName", varName))
		return nil
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		zap.L().Fatal("failed to load AWS config", zap.Error(err), zap.String("paramName", paramName))
		return nil
	}

	svc := ssm.NewFromConfig(cfg)
	param, err := svc.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		zap.L().Fatal("failed to fetch parameter", zap.Error(err), zap.String("paramName", paramName))
		return nil
	}

	return param.Parameter.Value
}

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		// Check if there are extra headers in the param store
		headers := getSecretFromParamStore("OTEL_EXPORTER_OTLP_HEADERS_NAME")
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

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
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
	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	ctx = tracer.NewContext(ctx, tp, "prod-logger-tracer")

	return ctx, nil
}
