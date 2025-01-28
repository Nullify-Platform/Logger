package tracer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestCustomAttributeCarrier(t *testing.T) {
	ctx := context.Background()
	tc := propagation.TraceContext{}
	tp := sdktrace.NewTracerProvider()
	ctx = NewContext(ctx, tp, "test-logger-tracer")

	// extract

	extractAttributes := map[string]string{
		"traceparent": "00-adabf450d7a37f3a7b708aaaffffe150-95f2f55992ea2f1e-01",
	}

	ctx = tc.Extract(ctx, newCustomMessageCarrier(extractAttributes))
	ctx, span := FromContext(ctx).Start(ctx, "TestCustomAttributeCarrier")

	assert.Equal(t, "adabf450d7a37f3a7b708aaaffffe150", span.SpanContext().TraceID().String())

	// inject
	injectAttributes := map[string]string{}

	tc.Inject(ctx, newCustomMessageCarrier(injectAttributes))

	expectedInjectAttributes := map[string]string{
		"traceparent": fmt.Sprintf("00-adabf450d7a37f3a7b708aaaffffe150-%s-01", span.SpanContext().SpanID().String()),
	}

	assert.Equal(t, expectedInjectAttributes, injectAttributes)
}
