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

func TestWithAgent(t *testing.T) {
	tests := []struct {
		name     string
		input    AgentFields
		expected map[string]interface{}
	}{
		{
			name: "all fields",
			input: AgentFields{
				Name:   "test-agent",
				Status: "running",
			},
			expected: map[string]interface{}{
				"name":   "test-agent",
				"status": "running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithAgent(tt.input)
			result := fieldToMap(t, field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithService(t *testing.T) {
	toolName := "test-tool"
	toolVersion := "1.0.0"
	category := "test-category"

	tests := []struct {
		name     string
		input    ServiceFields
		expected map[string]interface{}
	}{
		{
			name: "required fields only",
			input: ServiceFields{
				Name: "test-service",
			},
			expected: map[string]interface{}{
				"name": "test-service",
			},
		},
		{
			name: "all fields",
			input: ServiceFields{
				Name:        "test-service",
				ToolName:    &toolName,
				ToolVersion: &toolVersion,
				Category:    &category,
			},
			expected: map[string]interface{}{
				"name":         "test-service",
				"tool_name":    "test-tool",
				"tool_version": "1.0.0",
				"category":     "test-category",
			},
		},
		{
			name: "partial optional fields",
			input: ServiceFields{
				Name:     "test-service",
				ToolName: &toolName,
			},
			expected: map[string]interface{}{
				"name":      "test-service",
				"tool_name": "test-tool",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithService(tt.input)
			result := fieldToMap(t, field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithRepository(t *testing.T) {
	owner := "test-owner"

	tests := []struct {
		name     string
		input    RepositoryFields
		expected map[string]interface{}
	}{
		{
			name: "required fields only",
			input: RepositoryFields{
				Name:           "test-repo",
				Platform:       "github",
				InstallationID: "12345",
			},
			expected: map[string]interface{}{
				"name":            "test-repo",
				"platform":        "github",
				"installation_id": "12345",
			},
		},
		{
			name: "all fields",
			input: RepositoryFields{
				Name:           "test-repo",
				Owner:          &owner,
				Platform:       "github",
				InstallationID: "12345",
			},
			expected: map[string]interface{}{
				"name":            "test-repo",
				"owner":           "test-owner",
				"platform":        "github",
				"installation_id": "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := WithRepository(tt.input)
			result := fieldToMap(t, field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithErrorInfo(t *testing.T) {
	traceback := "test traceback"

	tests := []struct {
		name     string
		input    ErrorFields
		expected map[string]interface{}
	}{
		{
			name: "required fields only",
			input: ErrorFields{
				Type:    ErrorTypeAgent,
				Message: "test error",
			},
			expected: map[string]interface{}{
				"error.type":    "agent_error",
				"error.message": "test error",
			},
		},
		{
			name: "all fields",
			input: ErrorFields{
				Type:      ErrorTypeAgent,
				Message:   "test error",
				Traceback: &traceback,
			},
			expected: map[string]interface{}{
				"error.type":      "agent_error",
				"error.message":   "test error",
				"error.traceback": "test traceback",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := WithErrorInfo(tt.input)
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
