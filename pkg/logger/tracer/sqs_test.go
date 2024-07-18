package tracer

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestSQSAttributeCarrier(t *testing.T) {
	ctx := context.Background()
	tc := propagation.TraceContext{}
	tp := sdktrace.NewTracerProvider()
	ctx = NewContext(ctx, tp, "test-logger-tracer")

	// extract

	lambdaEventAttributes := &map[string]events.SQSMessageAttribute{
		"traceparent": {
			DataType:    "String",
			StringValue: aws.String("00-adabf450d7a37f3a7b708aaaffffe150-95f2f55992ea2f1e-01"),
		},
	}

	ctx = tc.Extract(ctx, newSQSEventMessageCarrier(lambdaEventAttributes))
	ctx, span := FromContext(ctx).Start(ctx, "TestSQSAttributeCarrier")

	assert.Equal(t, "adabf450d7a37f3a7b708aaaffffe150", span.SpanContext().TraceID().String())

	// inject
	sdkAttributes := &map[string]types.MessageAttributeValue{}

	tc.Inject(ctx, newSQSMessageCarrier(sdkAttributes))

	expectedSDKAttributes := map[string]types.MessageAttributeValue{
		"traceparent": {
			DataType:    aws.String("String"),
			StringValue: aws.String(fmt.Sprintf("00-adabf450d7a37f3a7b708aaaffffe150-%s-01", span.SpanContext().SpanID().String())),
		},
	}

	assert.Equal(t, expectedSDKAttributes, *sdkAttributes)
}
