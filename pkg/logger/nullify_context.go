package logger

import (
	"context"

	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type nullifyContextKeyType string

const nullifyContextKey nullifyContextKeyType = "NullifyContext"

// NullifyContext holds a mutable tree of BotActions and other useful data
type NullifyContext struct {
	AWSConfig aws.Config
	Span      trace.Span
	LogConfig *LogConfig // Changed to pointer to make it optional
}

type LogConfig struct {
	Repository *Repository
	Service    *Service
	Tool       *Tool
}

type Repository struct {
	Name     string  `json:"repository_name"`
	Owner    string  `json:"repository_owner"`
	Platform string  `json:"repository_platform"`
	CommitID *string `json:"repository_commit_id"`
	PRNumber *string `json:"repository_pr_number"`
}

type Service struct {
	Name            string `json:"service_name"`
	ServiceCategory string `json:"service_category"`
}

type Tool struct {
	Name   string `json:"tool_name"`
	Status string `json:"tool_status"`
}

// NewNullifyContext should be called at the entrypoint of a new request to initialize the NullifyContext.
// it creates a new OpenTelemetry span and initialises the NullifyContext.
func NewNullifyContext(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := tracer.StartNewRootSpan(ctx, spanName)
	ctx, nullifyContext := GetNullifyContext(ctx)
	nullifyContext.Span = span
	return ctx, span
}

// GetNullifyContext creates a new NullifyContext if one does not already exist in the context.
func GetNullifyContext(ctx context.Context) (context.Context, *NullifyContext) {
	if nullifyContext, ok := ctx.Value(nullifyContextKey).(*NullifyContext); ok {
		return ctx, nullifyContext
	} else {
		L(ctx).Info("creating new NullifyContext")
		nullifyContext = &NullifyContext{}

		return context.WithValue(ctx, nullifyContextKey, nullifyContext), nullifyContext
	}
}

func GetTraceID(ctx context.Context) string {
	_, nullifyContext := GetNullifyContext(ctx)

	return nullifyContext.Span.SpanContext().TraceID().String()
}

func GetAWSConfig(ctx context.Context) (aws.Config, error) {
	_, nullifyContext := GetNullifyContext(ctx)

	if nullifyContext.AWSConfig.Region == "" {
		awsConfig, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			L(ctx).Error("error loading AWS config", Err(err))
			return aws.Config{}, err
		}

		nullifyContext.AWSConfig = awsConfig
	}

	return nullifyContext.AWSConfig, nil
}

func SetServiceInfo(ctx context.Context, serviceName string, serviceCategory string) {
	_, nullifyContext := GetNullifyContext(ctx)

	nullifyContext.LogConfig.Service = &Service{
		Name:            serviceName,
		ServiceCategory: serviceCategory,
	}
}

type contextKey string

func SetMetadata(ctx context.Context, metadata map[string]string) context.Context {
	for key, value := range metadata {
		ctx = setContextMetadata(ctx, contextKey(key), value)
	}
	return ctx
}

func setContextMetadata(ctx context.Context, inputKey contextKey, inputValue string) context.Context {
	return context.WithValue(ctx, contextKey(inputKey), inputValue)
}

func GetContextMetadataAsFields(ctx context.Context, s interface{}, metadata map[string]string) []zapcore.Field {
	var fields []zapcore.Field
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Ensure the input is a struct
	if v.Kind() != reflect.Struct {
		return fields
	}

	// Iterate through all struct fields
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonKey := field.Tag.Get("json")

		// Skip fields without a JSON tag
		if jsonKey == "" || jsonKey == "-" {
			continue
		}

		// Retrieve the value from context using the field name as the key
		key := contextKey(field.Name)
		if value, ok := ctx.Value(key).(string); ok {
			// Append zap field
			fields = append(fields, zap.String(jsonKey, value))
		}
	}

	return fields
}
