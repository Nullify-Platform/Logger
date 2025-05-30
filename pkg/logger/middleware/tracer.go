package middleware

import (
	"fmt"
	"net/http"

	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/attribute"
)

func TracerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx = tracer.ExtractTracingFromHTTPHeaders(ctx, r.Header)

		ctx, span := tracer.FromContext(ctx).Start(ctx, fmt.Sprintf("http call: %s %s", r.Method, r.URL.EscapedPath()))
		defer tracer.ForceFlush(ctx)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.host", r.Host),
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.EscapedPath()),
			attribute.String("http.query", r.URL.Query().Encode()),
		)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
