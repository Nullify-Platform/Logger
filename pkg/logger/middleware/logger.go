// Package middleware provides a middleware for logging http requests and injecting tracing
package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type responseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type httpRequestMetadata struct {
	Host            string        `json:"host"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Query           url.Values    `json:"query"`
	StatusCode      int           `json:"statusCode"`
	RequestHeaders  []string      `json:"requestHeaders"`
	ResponseHeaders []string      `json:"responseHeaders"`
	Duration        time.Duration `json:"duration"`
}

// LoggingMiddleware logs the incoming request and the outgoing response and adds relevant tracing information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.FromContext(r.Context()).Start(r.Context(), fmt.Sprint("http call", r.URL.EscapedPath()))
		defer func() {
			// Check if there is a parent span
			if parentSpan := trace.SpanFromContext(ctx); !parentSpan.SpanContext().IsValid() {
				logger.FromContext(ctx).Sync()
			}
		}()
		defer span.End()

		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					sentry.CaptureException(e)
				}

				w.WriteHeader(http.StatusInternalServerError)
				logger.FromContext(ctx).Error(
					"endpoint handler panicked",
					logger.Any("err", err),
					logger.Trace(debug.Stack()),
				)
			}
		}()

		reqHeaders := []string{}
		for header, values := range r.Header {
			if strings.ToLower(header) == "authorization" {
				reqHeaders = append(reqHeaders, header+": "+"[REDACTED]")
				continue
			}

			for _, value := range values {
				reqHeaders = append(reqHeaders, header+": "+value)
			}
		}

		metadata := httpRequestMetadata{
			Host:           r.Host,
			Method:         r.Method,
			Path:           r.URL.EscapedPath(),
			Query:          r.URL.Query(),
			RequestHeaders: reqHeaders,
		}

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.host", r.Host),
			attribute.String("http.Path", r.URL.EscapedPath()),
			attribute.String("http.Query", r.URL.Query().Encode()),
			attribute.StringSlice("http.RequestHeaders", reqHeaders),
		)

		if r.URL.EscapedPath() != "/healthcheck" {
			logger.FromContext(ctx).Info(
				"new request",
				logger.Any("requestSummary", metadata),
			)
		}

		span.AddEvent("delegating request")
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r.WithContext(ctx))
		span.AddEvent("request completed")

		resHeaders := []string{}
		for header, values := range rw.Header() {
			for _, value := range values {
				resHeaders = append(resHeaders, header+": "+value)
			}
		}

		metadata.StatusCode = rw.StatusCode
		metadata.ResponseHeaders = resHeaders
		metadata.Duration = time.Since(start)
		span.AddEvent("response parsing complete")

		if r.URL.EscapedPath() != "/healthcheck" {
			logger.FromContext(ctx).Info(
				"request summary",
				logger.Any("requestSummary", metadata),
			)
		}
	})
}
