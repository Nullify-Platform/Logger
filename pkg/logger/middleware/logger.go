// Package middleware provides a middleware for logging http requests and injecting tracing
package middleware

import (
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/nullify-platform/logger/pkg/logger"
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

// ToLogFields converts the httpRequestMetadata to logger fields for structured logging
func (m httpRequestMetadata) ToLogFields() []logger.Field {
	fields := []logger.Field{
		logger.String("host", m.Host),
		logger.String("method", m.Method),
		logger.String("path", m.Path),
		logger.Any("query", m.Query),
		logger.Int("statusCode", m.StatusCode),
		logger.Any("requestHeaders", m.RequestHeaders),
		logger.Any("responseHeaders", m.ResponseHeaders),
		logger.Duration("duration", m.Duration),
	}
	return fields
}

// LoggingMiddleware logs the incoming request and the outgoing response and adds relevant tracing information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				logger.L(ctx).Error(
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

			if strings.ToLower(header) == "cookie" {
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

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		resHeaders := []string{}
		for header, values := range rw.Header() {
			for _, value := range values {
				resHeaders = append(resHeaders, header+": "+value)
			}
		}

		metadata.StatusCode = rw.StatusCode
		metadata.ResponseHeaders = resHeaders
		metadata.Duration = duration

		if r.URL.EscapedPath() != "/healthcheck" {
			log := logger.L(ctx)
			switch {
			case metadata.StatusCode >= 500:
				// Server error
				log.Error("request summary", metadata.ToLogFields()...)
			case metadata.StatusCode >= 400:
				// Client error
				log.Warn("request summary", metadata.ToLogFields()...)
			default:
				log.Info("request summary", metadata.ToLogFields()...)
			}
		}
	})
}
