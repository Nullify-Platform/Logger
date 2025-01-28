package logger

import (
	"context"

	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type nullifyContextKeyType string

const nullifyContextKey nullifyContextKeyType = "NullifyContext"

type LogConfig struct {
	Repository *repository
	Service    *service
	Tool       *tool
}

type repository struct {
	Name           *string `json:"repository_name"`
	Owner          *string `json:"repository_owner"`
	Platform       *string `json:"platform"`
	ID             *string `json:"repository_id"`
	CommitID       *string `json:"commit_id"` //head commit id
	PRNumber       *string `json:"pr_number"`
	Component      *string `json:"component"`
	BranchID       *string `json:"branch_id"`
	BranchName     *string `json:"branch_name"`
	InstallationID *string `json:"installation_id"`
	AppID          *string `json:"app_id"`
	Action         *string `json:"action"`
	ProjectID      *string `json:"project_id"`
	OrganizationID *string `json:"organization_id"`
}

type service struct {
	Name            *string `json:"service_name"`
	ServiceCategory *string `json:"service_category"`
	Event           *string `json:"service_event"`
}

type tool struct {
	Name   *string `json:"tool_name"`
	Status *string `json:"tool_status"`
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

func SetTraceAttributes(trace trace.Span, metadata map[string]string) trace.Span {
	for key, value := range metadata {
		trace.SetAttributes(attribute.String(key, value))
	}
	return trace
}

func SetMetdataForLogsAndTraces(ctx context.Context, trace trace.Span, metadata map[string]string) (context.Context, trace.Span) {
	mctx := SetMetadata(ctx, metadata)
	mtrace := SetTraceAttributes(trace, metadata)
	return mctx, mtrace
}

func (l *logger) getContextMetadataAsFields(logConfig LogConfig, fields []zapcore.Field) []zapcore.Field {
	return extractFieldsFromStruct(l.attachedContext, reflect.ValueOf(logConfig), fields)
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
