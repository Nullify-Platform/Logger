package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace/noop"
)

// MetricsMiddleware returns HTTP middleware that records standard OpenTelemetry
// server metrics (request duration, request body size, response body size).
// Tracing is disabled since TracerMiddleware already handles that.
//
// Usage:
//
//	service.Use(resthandler.MetricsMiddleware())
//
// To skip health checks:
//
//	service.Use(resthandler.MetricsMiddleware(
//	    otelhttp.WithFilter(func(r *http.Request) bool {
//	        return r.URL.Path != "/health"
//	    }),
//	))
func MetricsMiddleware(opts ...otelhttp.Option) func(http.Handler) http.Handler {
	defaults := []otelhttp.Option{
		otelhttp.WithTracerProvider(noop.NewTracerProvider()),
	}
	return otelhttp.NewMiddleware("http.server", append(defaults, opts...)...)
}
