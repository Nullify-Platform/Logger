package logger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

// helper function to convert Field to map for easier assertions
func fieldToMap(t *testing.T, field Field) map[string]interface{} {
	enc := zapcore.NewMapObjectEncoder()
	field.AddTo(enc)

	// Get the value and convert it to map
	raw := enc.Fields["agent"]
	if raw == nil {
		raw = enc.Fields["service"]
	}
	if raw == nil {
		raw = enc.Fields["repository"]
	}
	if raw == nil {
		return enc.Fields // for error fields
	}

	// Convert to JSON and back to ensure proper mapping
	jsonBytes, err := json.Marshal(raw)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)

	return result
}

func TestLogFields(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() []Field
		expected map[string]interface{}
	}{
		{
			name: "agent fields",
			builder: func() []Field {
				return NewLogFields().
					WithAgent("test-agent", "running").
					Build()
			},
			expected: map[string]interface{}{
				"agent": map[string]interface{}{
					"name":   "test-agent",
					"status": "running",
				},
			},
		},
		{
			name: "service with required fields only",
			builder: func() []Field {
				return NewLogFields().
					WithService("test-service").
					Build()
			},
			expected: map[string]interface{}{
				"service": map[string]interface{}{
					"name": "test-service",
				},
			},
		},
		{
			name: "service with all fields",
			builder: func() []Field {
				return NewLogFields().
					WithService("test-service").
					WithServiceTool("test-tool", "1.0.0").
					WithServiceCategory("test-category").
					Build()
			},
			expected: map[string]interface{}{
				"service": map[string]interface{}{
					"name":         "test-service",
					"tool_name":    "test-tool",
					"tool_version": "1.0.0",
					"category":     "test-category",
				},
			},
		},
		{
			name: "repository with required fields only",
			builder: func() []Field {
				return NewLogFields().
					WithRepository("test-repo", "github", "12345").
					Build()
			},
			expected: map[string]interface{}{
				"repository": map[string]interface{}{
					"name":            "test-repo",
					"platform":        "github",
					"installation_id": "12345",
				},
			},
		},
		{
			name: "repository with owner",
			builder: func() []Field {
				return NewLogFields().
					WithRepository("test-repo", "github", "12345").
					WithRepositoryOwner("test-owner").
					Build()
			},
			expected: map[string]interface{}{
				"repository": map[string]interface{}{
					"name":            "test-repo",
					"platform":        "github",
					"installation_id": "12345",
					"owner":           "test-owner",
				},
			},
		},
		{
			name: "error with required fields only",
			builder: func() []Field {
				return NewLogFields().
					WithError(ErrorTypeAgent, "test error").
					Build()
			},
			expected: map[string]interface{}{
				"error.type":    "agent_error",
				"error.message": "test error",
			},
		},
		{
			name: "error with traceback",
			builder: func() []Field {
				return NewLogFields().
					WithError(ErrorTypeAgent, "test error").
					WithErrorTraceback("test traceback").
					Build()
			},
			expected: map[string]interface{}{
				"error.type":      "agent_error",
				"error.message":   "test error",
				"error.traceback": "test traceback",
			},
		},
		{
			name: "all fields together",
			builder: func() []Field {
				return NewLogFields().
					WithAgent("test-agent", "failed").
					WithService("test-service").
					WithServiceTool("test-tool", "1.0.0").
					WithServiceCategory("test-category").
					WithRepository("test-repo", "github", "12345").
					WithRepositoryOwner("test-owner").
					WithError(ErrorTypeAgent, "test error").
					WithErrorTraceback("test traceback").
					Build()
			},
			expected: map[string]interface{}{
				"agent": map[string]interface{}{
					"name":   "test-agent",
					"status": "failed",
				},
				"service": map[string]interface{}{
					"name":         "test-service",
					"tool_name":    "test-tool",
					"tool_version": "1.0.0",
					"category":     "test-category",
				},
				"repository": map[string]interface{}{
					"name":            "test-repo",
					"platform":        "github",
					"installation_id": "12345",
					"owner":           "test-owner",
				},
				"error.type":      "agent_error",
				"error.message":   "test error",
				"error.traceback": "test traceback",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.builder()
			enc := zapcore.NewMapObjectEncoder()
			for _, f := range fields {
				f.AddTo(enc)
			}
			assert.Equal(t, tt.expected, enc.Fields)
		})
	}
}

func TestErrorType(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"unknown error", ErrorTypeUnknown, "unknown_error"},
		{"validation error", ErrorTypeValidation, "validation_error"},
		{"agent error", ErrorTypeAgent, "agent_error"},
		{"system error", ErrorTypeSystem, "system_error"},
		{"post scan error", ErrorTypePostScan, "postscan_error"},
		{"pre scan error", ErrorTypePreScan, "prescan_error"},
		{"scan error", ErrorTypeScan, "scan_error"},
		{"config error", ErrorTypeConfig, "config_error"},
		{"network error", ErrorTypeNetwork, "network_error"},
		{"timeout error", ErrorTypeTimeout, "timeout_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.errType))
		})
	}
}
