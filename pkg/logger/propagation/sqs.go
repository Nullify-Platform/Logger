// package propagation is a package for injecting and extracting tracing from AWS SNS and SQS messages.
package propagation

import (
	"context"

	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/udhos/opentelemetry-trace-sqs/otelsqs"
)

// InjectTracingIntoSQS inserts tracing from context into the SQS message attributes.
func InjectTracingIntoSQS(ctx context.Context, sqsMessage *sqsTypes.Message) {
	if sqsMessage.MessageAttributes == nil {
		sqsMessage.MessageAttributes = make(map[string]sqsTypes.MessageAttributeValue)
	}

	// We are not accounting for the case where there are > 10 attributes in the context.
	_ = otelsqs.NewCarrier().Inject(ctx, sqsMessage.MessageAttributes)
}

// ExtractTracingFromSQS extracts tracing from SQS message attributes.
func ExtractTracingFromSQS(ctx context.Context, sqsMessage *sqsTypes.Message) context.Context {
	return otelsqs.NewCarrier().Extract(ctx, sqsMessage.MessageAttributes)
}
