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

// Define custom context key type to avoid collisions
type nullifyContextKeyType string

const nullifyContextKey nullifyContextKeyType = "NullifyContext"

// NullifyContext holds a mutable tree of BotActions and other useful data
type NullifyContext struct {
	AWSConfig aws.Config
	Span      trace.Span
	LogConfig LogConfig
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
	Name      string `json:"platformName"`
	Component string `json:"platformComponent"`
}

// TraceConfig holds tracing configuration

// Repository holds repository-specific information
type Repository struct {
	Name           string `json:"repositoryName"`
	Owner          string `json:"repositoryOwner"`
	ID             string `json:"repositoryId"`
	CommitID       string `json:"commitId"`
	PrNumber       string `json:"prNumber"`
	BranchID       string `json:"branchId"`
	BranchName     string `json:"branchName"`
	InstallationID string `json:"installationId"`
	AppID          string `json:"appId"`
	Action         string `json:"action"`
	ProjectName    string `json:"projectName"`
	ProjectID      string `json:"projectId"`
	OrganizationID string `json:"organizationId"`
	StartCommitSha string `json:"startCommitSha"`
	EndCommitSha   string `json:"endCommitSha"`
	CloneURL       string `json:"cloneUrl"`
}

// Service holds service-specific information
type Service struct {
	Name     string `json:"serviceName"`
	Category string `json:"serviceCategory"`
	Event    string `json:"serviceEvent"`
}

// Tool holds tool-specific information
type Tool struct {
	Name   string `json:"toolName"`
	Status string `json:"toolStatus"`
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
		if value, ok := ctx.Value(key).(string); ok && value != "" {
			// Append zap field
			fields = append(fields, zap.String(jsonKey, value))
		}
	}

	return fields
}

// TODO: This is a temporary function to get the function name. We need to fine tune it to get the function name as there are some edge cases and we need to handle them
// func (l *logger) getFunctionName() string {
// 	pc, _, _, _ := runtime.Caller(2)
// 	fullName := runtime.FuncForPC(pc).Name()

// 	// Step 1: Remove full module and directory path
// 	if lastSlash := strings.LastIndex(fullName, "/"); lastSlash != -1 {
// 		fullName = fullName[lastSlash+1:] // Keep only package and function parts
// 	}

// 	// Step 2: Remove ".funcN" suffix, if it exists
// 	if funcSuffix := strings.LastIndex(fullName, ".func"); funcSuffix != -1 {
// 		fullName = fullName[:funcSuffix]
// 	}

// 	// Step 3: Handle struct method pattern (*StructName).Method
// 	if structMethodIdx := strings.Index(fullName, ")."); structMethodIdx != -1 {
// 		// Extract the part before ")." and after the last dot (e.g., "*GitHubTokenTransport")
// 		lastDotBeforeStruct := strings.LastIndex(fullName[:structMethodIdx], ".")
// 		fullName = fullName[:lastDotBeforeStruct] + fullName[structMethodIdx+2:]
// 	}

// 	return fullName
// }

// SetSpanAttributes sets attributes on the current span based on the LogConfig
func (l *logger) SetSpanAttributes(spanName string) {
	nullifyContext := l.attachedContext.Value(nullifyContextKey).(*NullifyContext)
	l.attachedContext, nullifyContext.Span = tracer.StartNewSpan(l.attachedContext, spanName)
	v := reflect.ValueOf(nullifyContext.LogConfig)
	t := v.Type()
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

		nullifyContext.Span.SetAttributes(attribute.String(jsonKey, field.String()))
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
		if field.String() != "" {
			nullifyContext.Span.SetAttributes(attribute.String(jsonKey, field.String()))
		}
	}
}
