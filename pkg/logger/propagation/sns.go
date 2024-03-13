package propagation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	snsTypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/udhos/opentelemetry-trace-sqs/otelsns"
)

// InjectTracingIntoSNS inserts tracing from context into the SNS message attributes.
func InjectTracingIntoSNS(ctx context.Context, input *sns.PublishInput) {
	if input.MessageAttributes == nil {
		input.MessageAttributes = make(map[string]snsTypes.MessageAttributeValue)
	}

	otelsns.NewCarrier().Inject(ctx, input.MessageAttributes)
}
