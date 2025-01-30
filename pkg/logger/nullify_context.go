package logger

import (
	"context"
	"reflect"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Define custom context key type to avoid collisions
type nullifyContextKeyType string

const nullifyContextKey nullifyContextKeyType = "NullifyContext"

// NullifyContext holds a mutable tree of BotActions and other useful data
type NullifyContext struct {
	AWSConfig   aws.Config
	Span        trace.Span
	LogConfig   LogConfig
	TraceConfig TraceConfig
}

// LogConfig holds configuration for logging
type LogConfig struct {
	Repository Repository
	Service    Service
	Tool       Tool
	Platform   Platform
}

// Platform holds platform-specific information
type Platform struct {
	Name      string `json:"platform_name"`
	Component string `json:"platform_component"`
}

// TraceConfig holds tracing configuration
type TraceConfig struct {
	SpanName string `json:"span_name"`
	Span     trace.Span
}

// Repository holds repository-specific information
type Repository struct {
	Name           string `json:"repository_name"`
	Owner          string `json:"repository_owner"`
	ID             string `json:"repository_id"`
	CommitID       string `json:"commit_id"`
	PrNumber       string `json:"pr_number"`
	BranchID       string `json:"branch_id"`
	BranchName     string `json:"branch_name"`
	InstallationID string `json:"installation_id"`
	AppID          string `json:"app_id"`
	Action         string `json:"action"`
	ProjectID      string `json:"project_id"`
	OrganizationID string `json:"organization_id"`
	StartCommitSha string `json:"start_commit_sha"`
	EndCommitSha   string `json:"end_commit_sha"`
	CloneURL       string `json:"clone_url"`
}

// Service holds service-specific information
type Service struct {
	Name     string `json:"service_name"`
	Category string `json:"service_category"`
	Event    string `json:"service_event"`
}

// Tool holds tool-specific information
type Tool struct {
	Name   string `json:"tool_name"`
	Status string `json:"tool_status"`
}

// GetNullifyContext creates a new NullifyContext if one does not already exist in the context.
func GetNullifyContext(ctx context.Context) (context.Context, *NullifyContext) {
	if nullifyContext, ok := ctx.Value(nullifyContextKey).(*NullifyContext); ok {
		return ctx, nullifyContext
	} else {
		nullifyContext = &NullifyContext{}
		ctx, span := tracer.StartNewRootSpan(ctx, "context.InitializeNullifyContext")
		nullifyContext.Span = span
		return context.WithValue(ctx, nullifyContextKey, nullifyContext), nullifyContext
	}
}

// GetAWSConfig retrieves the AWS configuration from the context or loads it if not present
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

// contextKey is a custom type for context keys
type contextKey string

// getContextMetadataAsFields extracts fields from the attached context
func (l *logger) getContextMetadataAsFields(fields []zapcore.Field) []zapcore.Field {
	return extractFieldsFromStruct(l.attachedContext, reflect.ValueOf(LogConfig{}), fields)
}

// extractFieldsFromStruct extracts fields from a struct and appends them as zap fields
func extractFieldsFromStruct(ctx context.Context, v reflect.Value, fields []zapcore.Field) []zapcore.Field {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fields
	}

	t := v.Type()

	// Iterate through all struct fields
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// If the field is a struct, recurse into it
		if fieldValue.Kind() == reflect.Struct || (fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.Struct) {
			fields = extractFieldsFromStruct(ctx, fieldValue, fields)
			continue
		}

		jsonKey := field.Tag.Get("json")

		// Skip fields without a JSON tag
		if jsonKey == "" || jsonKey == "-" {
			continue
		}

		// Retrieve the value from context using the field name as the key
		key := contextKey(field.Name)
		if value, ok := ctx.Value(key).(string); ok {
			// Append zap field
			fields = append(fields, zap.String(snakeToCamel(jsonKey), value))
		}
	}

	return fields
}

// snakeToCamel converts snake_case to camelCase
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		r := []rune(parts[i])
		if len(r) > 0 {
			r[0] = unicode.ToTitle(r[0])
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "")
}

// SetSpanAttributes sets attributes on the current span based on the LogConfig
func (l *logger) SetSpanAttributes(spanName string) {
	nullifyContext := l.attachedContext.Value(nullifyContextKey).(*NullifyContext)
	l.attachedContext, nullifyContext.Span = tracer.StartNewSpan(l.attachedContext, spanName)
	defer nullifyContext.Span.End()
	if nullifyContext.TraceConfig.SpanName != "" {
		v := reflect.ValueOf(nullifyContext.LogConfig)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)
			nullifyContext.Span.SetAttributes(attribute.String(snakeToCamel(fieldType.Name), field.Kind().String()))
			// If the field is a struct, recurse into it
			if field.Kind() == reflect.Struct {
				l.setStructAttributes(field)
				continue
			}

			jsonKey := fieldType.Tag.Get("json")
			if jsonKey == "" || jsonKey == "-" {
				continue
			}

			nullifyContext.Span.SetAttributes(attribute.String(snakeToCamel(jsonKey), field.String()))
		}
	}
}

// setStructAttributes sets attributes for struct fields recursively
func (l *logger) setStructAttributes(v reflect.Value) {
	t := v.Type()
	nullifyContext := l.attachedContext.Value(nullifyContextKey).(*NullifyContext)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		// If the field is a struct, recurse into it
		if field.Kind() == reflect.Struct {
			l.setStructAttributes(field)
			continue
		}
		jsonKey := fieldType.Tag.Get("json")
		if jsonKey == "" || jsonKey == "-" {
			continue
		}
		nullifyContext.Span.SetAttributes(attribute.String(snakeToCamel(jsonKey), field.String()))
	}
}
