package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type customMessageAttributeCarrier struct {
	Attributes map[string]string
}

func (c *customMessageAttributeCarrier) Get(key string) string {
	return c.Attributes[key]
}

func (c *customMessageAttributeCarrier) Set(key, value string) {
	c.Attributes[key] = value
}

func (c *customMessageAttributeCarrier) Keys() []string {
	keys := make([]string, 0, len(c.Attributes))
	for k := range c.Attributes {
		keys = append(keys, k)
	}
	return keys
}

func newCustomMessageCarrier(attributes map[string]string) propagation.TextMapCarrier {
	return &customMessageAttributeCarrier{Attributes: attributes}
}

// InjectTracingIntoCustomMessage inserts tracing from context into the Custom message attributes.
func InjectTracingIntoCustomMessage(ctx context.Context, attributes map[string]string) {
	if attributes == nil {
		attributes = make(map[string]string)
	}

	otel.GetTextMapPropagator().Inject(ctx, newCustomMessageCarrier(attributes))
}

type customEventMessageAttributeCarrier struct {
	Attributes map[string]string
}

func (c *customEventMessageAttributeCarrier) Get(key string) string {
	return c.Attributes[key]
}

func (c *customEventMessageAttributeCarrier) Set(key, value string) {
	c.Attributes[key] = value
}

func (c *customEventMessageAttributeCarrier) Keys() []string {
	keys := make([]string, 0, len(c.Attributes))
	for k := range c.Attributes {
		keys = append(keys, k)
	}
	return keys
}

func newCustomEventMessageCarrier(attributes map[string]string) propagation.TextMapCarrier {
	return &customEventMessageAttributeCarrier{Attributes: attributes}
}

// ExtractTracingFromCustomEventMessage extracts tracing from Custom event message attributes.
func ExtractTracingFromCustomEventMessage(ctx context.Context, attributes map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, newCustomEventMessageCarrier(attributes))
}
