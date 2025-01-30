package logger

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Version is the current version of the application
// override this value with ldflags
// e.g. -ldflags "-X 'github.com/nullify-platform/logger/pkg/logger.Version=$(VERSION)'"
var Version = "0.0.0"

var extraTags = map[string]string{}

// ConfigureDevelopmentLogger configures a development logger which is more human readable instead of JSON
func ConfigureDevelopmentLogger(ctx context.Context, level string, syncs ...io.Writer) (context.Context, error) {
	// configure level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Error("failed to parse log level, using info", zap.Error(err))
		zapLevel = zapcore.InfoLevel
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
		zap.L().Error("failed to create trace exporter, continuing...", zap.Error(err))
	}

	tp, err := newTraceProvider(traceExporter)
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
func ConfigureProductionLogger(ctx context.Context, spanName string, level string, syncs ...io.Writer) (context.Context, trace.Span, error) {
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zap.L().Error("failed to parse log level, using info", zap.Error(err))
		zapLevel = zapcore.InfoLevel
	}

	var sync io.Writer = os.Stdout
	if len(syncs) > 0 {
		sync = syncs[0]
	}

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

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
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
		zap.L().Error("failed to create trace exporter, continuing", zap.Error(err))
	}

	tp, err := newTraceProvider(traceExporter)
	if err != nil {
		return nil, nil, err
	}

	otel.SetTracerProvider(tp)
	tc := propagation.TraceContext{}
	otel.SetTextMapPropagator(tc)

	l := &logger{underlyingLogger: zapLogger}
	ctx = l.InjectIntoContext(ctx)
	ctx = tracer.NewContext(ctx, tp, "prod-logger-tracer")
	ctx, span := NewNullifyContext(ctx, spanName)
	return ctx, span, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersion(Version),
		),
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

	fixMechanismTypeInSentryEvents()
}

// AddLambdaTagsToSentryEvents Sets `Environment`, `ServerName` and adds `service`, `tenant` and `region` tags to Sentry events
func AddLambdaTagsToSentryEvents(ctx context.Context, awsConfig aws.Config) error {
	if os.Getenv("SENTRY_DSN") == "" {
		return nil
	}

	lambdaClient := lambda.NewFromConfig(awsConfig)

	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if functionName != "" {
		err := os.Setenv("OTEL_SERVICE_NAME", functionName)
		if err != nil {
			zap.L().Error("failed to set OTEL_SERVICE_NAME", zap.Error(err))
		}

		functionDetails, err := lambdaClient.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		})
		if err != nil {
			zap.L().Error("failed to get lambda function details", zap.Error(err))
			return err
		}

		addTagsToSentryEvents(functionName, os.Getenv("AWS_REGION"), functionDetails.Tags)
	}

	return nil
}

type ecsMetadata struct {
	DockerID string `json:"DockerId"`
	Name     string `json:"Name"`
	Labels   struct {
		ClusterARN    string `json:"com.amazonaws.ecs.cluster"`
		ContainerName string `json:"com.amazonaws.ecs.container-name"`
		TaskARN       string `json:"com.amazonaws.ecs.task-arn"`
	} `json:"Labels"`
	LogOptions struct {
		Region    string `json:"awslogs-region"`
		LogGroup  string `json:"awslogs-group"`
		LogStream string `json:"awslogs-stream"`
	} `json:"LogOptions"`
}

// AddECSTagsToSentryEvents Sets `Environment`, `ServerName` and adds `service`, `tenant` and `region` tags to Sentry events
func AddECSTagsToSentryEvents(ctx context.Context, awsConfig aws.Config) error {
	if os.Getenv("SENTRY_DSN") == "" {
		return nil
	}

	cluster := os.Getenv("ECS_CLUSTER")
	ecsName := os.Getenv("FARGATE_TASK_NAME") // set by application
	if ecsName == "" {
		ecsName = os.Getenv("ECS_SERVICE_NAME") // set by ECS if task is launched as part of an ECS service
	}

	metadataEndpoint := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if metadataEndpoint != "" {
		resp, err := http.Get(metadataEndpoint)
		if err != nil {
			zap.L().Error("failed to get ECS metadata", zap.Error(err))
			return err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				zap.L().Error("failed to close ECS metadata body", zap.Error(err))
			}
		}()

		var metadata ecsMetadata
		if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
			zap.L().Error("failed to parse ECS metadata", zap.Error(err))
		} else {
			if ecsName == "" {
				ecsName = metadata.Name
			}
			cluster = metadata.Labels.ClusterARN

			if err = os.Setenv("AWS_ECS_LOG_GROUP_NAME", metadata.LogOptions.LogGroup); err != nil {
				zap.L().Error("failed to set AWS_ECS_LOG_GROUP_NAME", zap.Error(err))
			}
			if err = os.Setenv("AWS_ECS_LOG_STREAM_NAME", metadata.LogOptions.LogStream); err != nil {
				zap.L().Error("failed to set AWS_ECS_LOG_STREAM_NAME", zap.Error(err))
			}
		}
	}

	tags := map[string]string{}

	if cluster == "" {
		zap.L().Warn("ECS_CLUSTER is not set, will not be able to add cluster tags to sentry error events")
	} else {
		ecsClient := ecs.NewFromConfig(awsConfig)
		clusterDetails, err := ecsClient.DescribeClusters(ctx, &ecs.DescribeClustersInput{
			Clusters: []string{cluster},
			Include:  []types.ClusterField{types.ClusterFieldTags},
		})
		if err != nil {
			zap.L().Error("failed to get cluster details", zap.Error(err))
		} else {
			if len(clusterDetails.Clusters) == 0 {
				zap.L().Warn("cluster not found", zap.String("clusterName", cluster))
			} else {
				for _, tag := range clusterDetails.Clusters[0].Tags {
					tags[*tag.Key] = *tag.Value
				}
			}
		}
	}

	err := os.Setenv("ECS_SERVICE_NAME", ecsName)
	if err != nil {
		zap.L().Error("failed to set ECS_SERVICE_NAME", zap.Error(err))
	}
	err = os.Setenv("OTEL_SERVICE_NAME", ecsName)
	if err != nil {
		zap.L().Error("failed to set OTEL_SERVICE_NAME", zap.Error(err))
	}

	region := os.Getenv("AWS_REGION")
	addTagsToSentryEvents(ecsName, region, tags)

	return nil
}

// AddSentryTag allows the application to add arbitrary tags for Sentry events at runtime
func AddSentryTag(key string, value string) {
	extraTags[key] = value
}

func addTagsToSentryEvents(functionName string, region string, tags map[string]string) {
	zap.L().Info("adding tags to sentry events", zap.String("functionName", functionName), zap.String("region", region), zap.Any("tags", tags))

	// export tags so that they can be used by python agents & other scripts
	if err := os.Setenv("TAG_ENVIRONMENT", tags["Environment"]); err != nil {
		zap.L().Error("failed to set TAG_ENVIRONMENT", zap.Error(err))
	}
	if err := os.Setenv("TAG_TENANT", tags["Tenant"]); err != nil {
		zap.L().Error("failed to set TAG_TENANT", zap.Error(err))
	}
	if err := os.Setenv("TAG_SERVICE", tags["Service"]); err != nil {
		zap.L().Error("failed to set TAG_SERVICE", zap.Error(err))
	}

	// called by client.CaptureEvent() -> .processEvent() -> .prepareEvent()
	sentry.CurrentHub().Client().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		event.Environment = tags["Environment"]
		event.ServerName = functionName

		event.Tags["region"] = region
		event.Tags["environment"] = tags["Environment"]
		event.Tags["tenant"] = tags["Tenant"]
		event.Tags["service"] = tags["Service"]

		for k, v := range extraTags {
			event.Tags[k] = v
		}

		return event
	})
}

func fixMechanismTypeInSentryEvents() {
	// called by client.CaptureEvent() -> .processEvent() -> .prepareEvent()
	sentry.CurrentHub().Client().AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		for i := range event.Exception {
			if event.Exception[i].Mechanism != nil && event.Exception[i].Mechanism.Type == "" {
				// avoid "list[function-after[check_type_value(), function-wrap[_run_root_validator()]]]",
				event.Exception[i].Mechanism.Type = "error"
			}
		}

		return event
	})
}

func getSecretFromParamStore(varName string) *string {
	// check if the param name is defined in the environment
	paramName := os.Getenv(varName)
	if paramName == "" {
		return nil
	}

	ctx := context.Background()
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
