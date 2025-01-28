package tracer

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTracingIntoHTTPHeaders inserts tracing from context into the Custom message attributes.
func InjectTracingIntoHTTPHeaders(ctx context.Context, headers http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractTracingFromHTTPHeaders extracts tracing from Custom event message attributes.
func ExtractTracingFromHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
}
