package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field an enum type for log message fields
type Field = zapcore.Field

// Fields

// Trace adds a stac trace field to the logger
func Trace(trace []byte) Field {
	return zap.ByteString("trace", trace)
}

// Any adds a field to the logger
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Err adds an error field to the logger
func Err(err error) Field {
	return zap.Error(err)
}

// Errs adds a ;ist of errors field to the logger
func Errs(msg string, errs []error) Field {
	return zap.Errors(msg, errs)
}

// String adds a string field to the logger
func String(key string, val string) Field {
	return zap.String(key, val)
}

// Strings adds a list of strings field to the logger
func Strings(key string, val []string) Field {
	return zap.Strings(key, val)
}

// Bool adds a bool field to the logger
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Bools adds a list of bools field to the logger
func Bools(key string, val []bool) Field {
	return zap.Bools(key, val)
}

// Int adds an int field to the logger
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Ints adds a list of ints field to the logger
func Ints(key string, val []int) Field {
	return zap.Ints(key, val)
}

// Int32 adds an int32 field to the logger
func Int32(key string, val int32) Field {
	return zap.Int32(key, val)
}

// Int32s adds a list of int32s field to the logger
func Int32s(key string, val []int32) Field {
	return zap.Int32s(key, val)
}

// Int64 adds an int64 field to the logger
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Int64s adds a list of int64s field to the logger
func Int64s(key string, val []int64) Field {
	return zap.Int64s(key, val)
}

// Duration adds a time duration field to the logger
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Durations adds a list of time durations field to the logger
func Durations(key string, val []time.Duration) Field {
	return zap.Durations(key, val)
}

// AgentFields represents agent-related logging fields
type AgentFields struct {
	Name       string
	Status     string
	TraceID    string
	traceIDSet bool
}

// RepositoryFields represents repository-related logging fields
type RepositoryFields struct {
	Name           string
	Platform       string
	InstallationID string
	Owner          string // optional
	ownerSet       bool   // internal tracking
}

// ServiceFields represents service-related logging fields
type ServiceFields struct {
	Name           string
	ToolName       string // optional
	ToolVersion    string // optional
	Category       string // optional
	toolNameSet    bool   // internal tracking
	toolVersionSet bool
	categorySet    bool
}

// ErrorType represents the type of error that occurred
type ErrorType string

const (
	// Common error types
	ErrorTypeUnknown    ErrorType = "unknown_error"
	ErrorTypeValidation ErrorType = "validation_error"
	ErrorTypeAgent      ErrorType = "agent_error"
	ErrorTypeSystem     ErrorType = "system_error"
	ErrorTypePostScan   ErrorType = "postscan_error"
	ErrorTypePreScan    ErrorType = "prescan_error"
	ErrorTypeScan       ErrorType = "scan_error"
	ErrorTypeConfig     ErrorType = "config_error"
	ErrorTypeNetwork    ErrorType = "network_error"
	ErrorTypeTimeout    ErrorType = "timeout_error"
	// Custom error types
	ErrorTypeFatalExecution ErrorType = "fatal_execution_error"
)

// ErrorFields represents error-related logging fields
type ErrorFields struct {
	Type         ErrorType
	Message      string
	Traceback    string // optional
	tracebackSet bool
}

// WithAgent adds agent-related fields to the log entry
func WithAgent(agent AgentFields) Field {
	fields := map[string]interface{}{
		"name":   agent.Name,
		"status": agent.Status,
	}

	if agent.traceIDSet {
		fields["trace_id"] = agent.TraceID
	}

	return Any("agent", fields)
}

// Add this method to AgentFields
func (a *AgentFields) WithTraceID(traceID string) *AgentFields {
	a.TraceID = traceID
	a.traceIDSet = true
	return a
}

// WithRepository adds repository-related fields to the log entry
func WithRepository(repo RepositoryFields) Field {
	fields := map[string]interface{}{
		"name":            repo.Name,
		"platform":        repo.Platform,
		"installation_id": repo.InstallationID,
	}

	if repo.ownerSet {
		fields["owner"] = repo.Owner
	}

	return Any("repository", fields)
}

// WithService adds service-related fields to the log entry
func WithService(service ServiceFields) Field {
	fields := map[string]interface{}{
		"name": service.Name,
	}

	if service.toolNameSet {
		fields["tool_name"] = service.ToolName
	}
	if service.toolVersionSet {
		fields["tool_version"] = service.ToolVersion
	}
	if service.categorySet {
		fields["category"] = service.Category
	}

	return Any("service", fields)
}

// WithErrorInfo adds error-related fields to the log entry
func WithErrorInfo(errFields ErrorFields) []Field {
	fields := []Field{
		String("error.type", string(errFields.Type)),
		String("error.message", errFields.Message),
	}

	if errFields.tracebackSet {
		fields = append(fields, String("error.traceback", errFields.Traceback))
	}

	return fields
}

// LogFields represents all logging-related fields
type LogFields struct {
	Agent      *AgentFields
	Repository *RepositoryFields
	Service    *ServiceFields
	Error      *ErrorFields
}

// NewLogFields creates a new LogFields instance
func NewLogFields() *LogFields {
	return &LogFields{}
}

// WithAgent adds agent-related fields
func (l *LogFields) WithAgent(name, status string) *LogFields {
	l.Agent = &AgentFields{
		Name:   name,
		Status: status,
	}
	return l
}

// Add this method
func (l *LogFields) WithAgentTraceID(traceID string) *LogFields {
	if l.Agent == nil {
		l.Agent = &AgentFields{}
	}
	l.Agent.WithTraceID(traceID)
	return l
}

// WithRepository adds required repository fields
func (l *LogFields) WithRepository(name, platform, installationID string) *LogFields {
	l.Repository = &RepositoryFields{
		Name:           name,
		Platform:       platform,
		InstallationID: installationID,
	}
	return l
}

// WithRepositoryOwner adds optional repository owner
func (l *LogFields) WithRepositoryOwner(owner string) *LogFields {
	if l.Repository == nil {
		l.Repository = &RepositoryFields{}
	}
	l.Repository.Owner = owner
	l.Repository.ownerSet = true
	return l
}

// WithService adds required service fields
func (l *LogFields) WithService(name string) *LogFields {
	l.Service = &ServiceFields{
		Name: name,
	}
	return l
}

// WithServiceTool adds optional service tool details
func (l *LogFields) WithServiceTool(name, version string) *LogFields {
	if l.Service == nil {
		l.Service = &ServiceFields{}
	}
	l.Service.ToolName = name
	l.Service.ToolVersion = version
	l.Service.toolNameSet = true
	l.Service.toolVersionSet = true
	return l
}

// WithServiceCategory adds optional service category
func (l *LogFields) WithServiceCategory(category string) *LogFields {
	if l.Service == nil {
		l.Service = &ServiceFields{}
	}
	l.Service.Category = category
	l.Service.categorySet = true
	return l
}

// WithError adds error fields
func (l *LogFields) WithError(errType ErrorType, message string) *LogFields {
	l.Error = &ErrorFields{
		Type:    errType,
		Message: message,
	}
	return l
}

// WithErrorTraceback adds optional error traceback
func (l *LogFields) WithErrorTraceback(traceback string) *LogFields {
	if l.Error == nil {
		l.Error = &ErrorFields{}
	}
	l.Error.Traceback = traceback
	l.Error.tracebackSet = true
	return l
}

// Build creates all the fields based on what was set
func (l *LogFields) Build() []Field {
	var fields []Field

	if l.Agent != nil {
		fields = append(fields, WithAgent(*l.Agent))
	}

	if l.Repository != nil {
		fields = append(fields, WithRepository(*l.Repository))
	}

	if l.Service != nil {
		fields = append(fields, WithService(*l.Service))
	}

	if l.Error != nil {
		fields = append(fields, WithErrorInfo(*l.Error)...)
	}

	return fields
}
