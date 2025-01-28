package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTracingIntoCustomMessage inserts tracing from context into the Custom message attributes.
func InjectTracingIntoCustomMessage(ctx context.Context, attributes map[string]string) {
	if attributes == nil {
		attributes = make(map[string]string)
	}

	otel.GetTextMapPropagator().Inject(ctx, newCustomMessageCarrier(attributes))
}

// ExtractTracingFromCustomEventMessage extracts tracing from Custom event message attributes.
func ExtractTracingFromCustomEventMessage(ctx context.Context, attributes map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, newCustomMessageCarrier(attributes))
}

func newCustomMessageCarrier(attributes map[string]string) propagation.TextMapCarrier {
	return &customMessageAttributeCarrier{Attributes: attributes}
}

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
