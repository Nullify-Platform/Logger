package logger

import (
	"context"

	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/attribute"
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
	LogConfig LogConfig
}

// GetNullifyContext creates a new NullifyContext if one does not already exist in the context.
func GetNullifyContext(ctx context.Context) (context.Context, *NullifyContext) {
	if nullifyContext, ok := ctx.Value(nullifyContextKey).(*NullifyContext); ok {
		return ctx, nullifyContext
	} else {
		nullifyContext = &NullifyContext{}

		return context.WithValue(ctx, nullifyContextKey, nullifyContext), nullifyContext
	}
}

func NewNullifyContext(ctx context.Context, spanName string) (context.Context, trace.Span) {
	if existingSpan := trace.SpanFromContext(ctx); existingSpan.SpanContext().IsValid() {
		// Create child span if parent exists
		ctx, span := tracer.FromContext(ctx).Start(ctx, spanName)
		ctx, nullifyContext := GetNullifyContext(ctx)
		nullifyContext.Span = span
		return ctx, span
	}

	// Create root span if no parent
	ctx, span := tracer.StartNewRootSpan(ctx, spanName)
	ctx, nullifyContext := GetNullifyContext(ctx)
	nullifyContext.Span = span
	return ctx, span
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

type LogConfig struct {
	Repository *Repository
	Service    *Service
	Tool       *Tool
}

type Repository struct {
	Name           string `json:"repository_name"`
	Owner          string `json:"repository_owner"`
	Platform       string `json:"platform"`
	ID             string `json:"repository_id"`
	CommitID       string `json:"commit_id"` //head commit id
	PRNumber       string `json:"pr_number"`
	Component      string `json:"component"`
	BranchID       string `json:"branch_id"`
	BranchName     string `json:"branch_name"`
	InstallationID string `json:"installation_id"`
	AppID          string `json:"app_id"`
	Action         string `json:"action"`
	ProjectID      string `json:"project_id"`
	OrganizationID string `json:"organization_id"`
}

type Service struct {
	Name            string `json:"service_name"`
	ServiceCategory string `json:"service_category"`
	Event           string `json:"service_event"`
}

type Tool struct {
	Name   string `json:"tool_name"`
	Status string `json:"tool_status"`
}

type contextKey string

func setMetadataLogConfig(ctx context.Context, logConfig LogConfig) context.Context {
	return context.WithValue(ctx, nullifyContextKey, logConfig)
}

func SetTraceAttributes(trace trace.Span, metadata map[string]string) trace.Span {
	for key, value := range metadata {
		trace.SetAttributes(attribute.String(key, value))
	}
	return trace
}

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
			fields = append(fields, zap.String(jsonKey, value))
		}
	}

	return fields
}

func SetMetadataFromLogConfig(ctx context.Context, newLogConfig LogConfig) context.Context {
	// Get existing context and LogConfig
	ctx, nullifyContext := GetNullifyContext(ctx)

	// If existing LogConfig is empty, just use the new one
	if nullifyContext.LogConfig == (LogConfig{}) {
		nullifyContext.LogConfig = newLogConfig
	} else {
		// Merge Repository fields if provided
		if newLogConfig.Repository != nil {
			if nullifyContext.LogConfig.Repository == nil {
				nullifyContext.LogConfig.Repository = &Repository{}
			}
			mergeRepositoryFields(nullifyContext.LogConfig.Repository, newLogConfig.Repository)
		}

		// Merge Service fields if provided
		if newLogConfig.Service != nil {
			if nullifyContext.LogConfig.Service == nil {
				nullifyContext.LogConfig.Service = &Service{}
			}
			mergeServiceFields(nullifyContext.LogConfig.Service, newLogConfig.Service)
		}
	}
	ctx = setMetadataLogConfig(ctx, newLogConfig)
	return ctx
}

func mergeRepositoryFields(existing, new *Repository) {
	if new.Name != "" {
		existing.Name = new.Name
	}
	if new.Owner != "" {
		existing.Owner = new.Owner
	}
	if new.Platform != "" {
		existing.Platform = new.Platform
	}
	if new.ID != "" {
		existing.ID = new.ID
	}
	if new.CommitID != "" {
		existing.CommitID = new.CommitID
	}
	if new.PRNumber != "" {
		existing.PRNumber = new.PRNumber
	}
	if new.Component != "" {
		existing.Component = new.Component
	}
	if new.BranchID != "" {
		existing.BranchID = new.BranchID
	}
	if new.BranchName != "" {
		existing.BranchName = new.BranchName
	}
	if new.InstallationID != "" {
		existing.InstallationID = new.InstallationID
	}
	if new.AppID != "" {
		existing.AppID = new.AppID
	}
	if new.Action != "" {
		existing.Action = new.Action
	}
	if new.ProjectID != "" {
		existing.ProjectID = new.ProjectID
	}
	if new.OrganizationID != "" {
		existing.OrganizationID = new.OrganizationID
	}
}

func mergeServiceFields(existing, new *Service) {
	if new.Name != "" {
		existing.Name = new.Name
	}
	if new.ServiceCategory != "" {
		existing.ServiceCategory = new.ServiceCategory
	}
	if new.Event != "" {
		existing.Event = new.Event
	}
}

func (l *logger) getFieldsFromLogConfigWithoutTags(fields []zapcore.Field) []zapcore.Field {
	if l.attachedContext == nil {
		L(l.attachedContext).Error("context is nil")
		return fields
	}
	_, nullifyContext := GetNullifyContext(l.attachedContext)
	if nullifyContext == nil {
		return fields
	}

	return extractFieldsFromStructWithoutTags(l.attachedContext, reflect.ValueOf(nullifyContext.LogConfig), fields)
}

func extractFieldsFromStructWithoutTags(ctx context.Context, v reflect.Value, fields []zapcore.Field) []zapcore.Field {
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
			fields = extractFieldsFromStructWithoutTags(ctx, fieldValue, fields)
			continue
		}

		// Use the field name directly as the key
		key := contextKey(field.Name)
		if value, ok := ctx.Value(key).(string); ok {
			// Append zap field using the field name
			fields = append(fields, zap.String(field.Name, value))
		}
	}

	return fields
}
