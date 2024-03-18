package tracer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snsTypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type snsPublishInputAttributeCarrier struct {
	Attributes *map[string]snsTypes.MessageAttributeValue
}

func (c *snsPublishInputAttributeCarrier) Get(key string) string {
	if val, ok := (*c.Attributes)[key]; ok {
		return *val.StringValue
	}
	return ""
}

func (c *snsPublishInputAttributeCarrier) Set(key, value string) {
	(*c.Attributes)[key] = snsTypes.MessageAttributeValue{
		StringValue: &value,
		DataType:    aws.String(stringType),
	}
}

func (c *snsPublishInputAttributeCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.Attributes))
	for k := range *c.Attributes {
		keys = append(keys, k)
	}
	return keys
}

func newSNSPublishInputCarrier(attributes *map[string]snsTypes.MessageAttributeValue) propagation.TextMapCarrier {
	return &snsPublishInputAttributeCarrier{Attributes: attributes}
}

// InjectTracingIntoSNS inserts tracing from context into the SNS message attributes. Call this function right before calling the SNS Publish API.
func InjectTracingIntoSNS(ctx context.Context, input *sns.PublishInput) {
	if input.MessageAttributes == nil {
		input.MessageAttributes = make(map[string]snsTypes.MessageAttributeValue)
	}

	otel.GetTextMapPropagator().Inject(ctx, newSNSPublishInputCarrier(&input.MessageAttributes))
}
