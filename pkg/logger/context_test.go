package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTraceAndSpanIDInLogs(t *testing.T) {
	// Setup in-memory exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)

	// Create buffer to capture log output
	var buf bytes.Buffer

	// Configure logger with buffer
	ctx, err := ConfigureProductionLogger(context.Background(), "info", &buf)
	if err != nil {
		t.Fatalf("Failed to configure logger: %v", err)
	}

	// Create a span
	ctx, span := tracer.FromContext(ctx).Start(ctx, "test-operation")
	defer span.End()

	// Get span context to verify IDs later
	spanCtx := span.SpanContext()
	expectedTraceID := spanCtx.TraceID().String()
	expectedSpanID := spanCtx.SpanID().String()

	// Log a message
	L(ctx).Info("test message")

	// Parse the logged JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v\nLog output: %s", err, buf.String())
	}

	// Verify trace_id is present
	traceID, ok := logEntry["trace_id"].(string)
	if !ok {
		t.Errorf("trace_id not found in log output. Log: %v", logEntry)
	} else if traceID != expectedTraceID {
		t.Errorf("trace_id mismatch: got %s, want %s", traceID, expectedTraceID)
	}

	// Verify span_id is present
	spanID, ok := logEntry["span_id"].(string)
	if !ok {
		t.Errorf("span_id not found in log output. Log: %v", logEntry)
	} else if spanID != expectedSpanID {
		t.Errorf("span_id mismatch: got %s, want %s", spanID, expectedSpanID)
	}

	// Verify message is correct
	msg, ok := logEntry["msg"].(string)
	if !ok || msg != "test message" {
		t.Errorf("msg mismatch: got %v, want 'test message'", msg)
	}
}

func TestNoSpanContext(t *testing.T) {
	// Create buffer to capture log output
	var buf bytes.Buffer

	// Configure logger with buffer
	ctx, err := ConfigureProductionLogger(context.Background(), "info", &buf)
	if err != nil {
		t.Fatalf("Failed to configure logger: %v", err)
	}

	// Log without a span (should not crash)
	L(ctx).Info("test message without span")

	// Parse the logged JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v\nLog output: %s", err, buf.String())
	}

	// trace_id and span_id should not be present (or be empty)
	// This verifies the code doesn't crash when there's no valid span
	msg, ok := logEntry["msg"].(string)
	if !ok || msg != "test message without span" {
		t.Errorf("msg mismatch: got %v, want 'test message without span'", msg)
	}
}
