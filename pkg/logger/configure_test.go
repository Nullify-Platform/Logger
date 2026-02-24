package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestOtelEnvFields(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		resourceAttrs  string
		expectedFields map[string]string
	}{
		{
			name:           "no env vars set",
			expectedFields: map[string]string{},
		},
		{
			name:        "only OTEL_SERVICE_NAME set",
			serviceName: "my-service",
			expectedFields: map[string]string{
				"service.name": "my-service",
			},
		},
		{
			name:          "only OTEL_RESOURCE_ATTRIBUTES set",
			resourceAttrs: "env=prod,region=us-east-1",
			expectedFields: map[string]string{
				"env":    "prod",
				"region": "us-east-1",
			},
		},
		{
			name:          "both set",
			serviceName:   "my-service",
			resourceAttrs: "env=staging",
			expectedFields: map[string]string{
				"service.name": "my-service",
				"env":          "staging",
			},
		},
		{
			name:          "malformed entries are skipped",
			resourceAttrs: "valid=yes,malformed,also-bad,good=ok",
			expectedFields: map[string]string{
				"valid": "yes",
				"good":  "ok",
			},
		},
		{
			name:          "value containing equals sign",
			resourceAttrs: "key=val=ue",
			expectedFields: map[string]string{
				"key": "val=ue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OTEL_SERVICE_NAME", tt.serviceName)
			t.Setenv("OTEL_RESOURCE_ATTRIBUTES", tt.resourceAttrs)

			fields := otelEnvFields()

			if len(fields) != len(tt.expectedFields) {
				t.Fatalf("expected %d fields, got %d: %v", len(tt.expectedFields), len(fields), fields)
			}

			for _, f := range fields {
				expected, ok := tt.expectedFields[f.Key]
				if !ok {
					t.Errorf("unexpected field key %q", f.Key)
					continue
				}
				if f.Type != zapcore.StringType {
					t.Errorf("expected string type for field %q, got %v", f.Key, f.Type)
					continue
				}
				if f.String != expected {
					t.Errorf("field %q: expected %q, got %q", f.Key, expected, f.String)
				}
			}
		})
	}
}

func TestLoggerOutputContainsOtelFields(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "test-svc")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=test,team=platform")

	var buf bytes.Buffer
	ctx, err := ConfigureProductionLogger(context.Background(), "info", &buf)
	if err != nil {
		t.Fatalf("failed to configure logger: %v", err)
	}

	l := L(ctx)
	l.Info("hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v\nraw: %s", err, buf.String())
	}

	expected := map[string]string{
		"service.name":           "test-svc",
		"deployment.environment": "test",
		"team":                   "platform",
	}
	for key, want := range expected {
		got, ok := entry[key]
		if !ok {
			t.Errorf("expected field %q in log output, not found", key)
			continue
		}
		if got != want {
			t.Errorf("field %q: expected %q, got %q", key, want, got)
		}
	}
}
