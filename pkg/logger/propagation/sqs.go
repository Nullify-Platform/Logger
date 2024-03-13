// Package propagation is a package for injecting and extracting tracing from AWS SNS and SQS messages.
package propagation

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type sqsMessageAttributeCarrier struct {
	Attributes *map[string]sqsTypes.MessageAttributeValue
}

func (c *sqsMessageAttributeCarrier) Get(key string) string {
	if val, ok := (*c.Attributes)[key]; ok {
		return *val.StringValue
	}
	return ""
}

func (c *sqsMessageAttributeCarrier) Set(key, value string) {
	(*c.Attributes)[key] = sqsTypes.MessageAttributeValue{StringValue: &value}
}

func (c *sqsMessageAttributeCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.Attributes))
	for k := range *c.Attributes {
		keys = append(keys, k)
	}
	return keys
}

func newSQSMessageCarrier(attributes *map[string]sqsTypes.MessageAttributeValue) propagation.TextMapCarrier {
	return &sqsMessageAttributeCarrier{Attributes: attributes}
}

// InjectTracingIntoSQSMessage inserts tracing from context into the SQS message attributes.
func InjectTracingIntoSQSMessage(ctx context.Context, sqsMessage *sqsTypes.Message) {
	if sqsMessage.MessageAttributes == nil {
		sqsMessage.MessageAttributes = make(map[string]sqsTypes.MessageAttributeValue)
	}

	otel.GetTextMapPropagator().Inject(ctx, newSQSMessageCarrier(&sqsMessage.MessageAttributes))
}

type sqsEventMessageAttributeCarrier struct {
	Attributes *map[string]events.SQSMessageAttribute
}

func (c *sqsEventMessageAttributeCarrier) Get(key string) string {
	if val, ok := (*c.Attributes)[key]; ok {
		return *val.StringValue
	}
	return ""
}

func (c *sqsEventMessageAttributeCarrier) Set(key, value string) {
	(*c.Attributes)[key] = events.SQSMessageAttribute{StringValue: &value}
}

func (c *sqsEventMessageAttributeCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.Attributes))
	for k := range *c.Attributes {
		keys = append(keys, k)
	}
	return keys
}

func newSQSEventMessageCarrier(attributes *map[string]events.SQSMessageAttribute) propagation.TextMapCarrier {
	return &sqsEventMessageAttributeCarrier{Attributes: attributes}
}

// ExtractTracingFromSQSEventMessage extracts tracing from SQS event message attributes.
func ExtractTracingFromSQSEventMessage(ctx context.Context, sqsMessage events.SQSMessage) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, newSQSEventMessageCarrier(&sqsMessage.MessageAttributes))
}
