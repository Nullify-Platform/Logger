package middleware

import (
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/nullify-platform/logger/pkg/logger"
	"go.uber.org/zap"
)

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type HTTPRequestMetadata struct {
	Host            string        `json:"host"`
	Method          string        `json:"method"`
	Path            string        `json:"path"`
	Query           url.Values    `json:"query"`
	StatusCode      int           `json:"statusCode"`
	RequestHeaders  []string      `json:"requestHeaders"`
	ResponseHeaders []string      `json:"responseHeaders"`
	Duration        time.Duration `json:"duration"`
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					sentry.CaptureException(e)
				}

				w.WriteHeader(http.StatusInternalServerError)
				logger.Error(
					"endpoint handler panicked",
					logger.Any("err", err),
					logger.Trace(debug.Stack()),
				)
			}

			sentry.Flush(200 * time.Millisecond)
			zap.L().Sync()
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

		metadata := HTTPRequestMetadata{
			Host:           r.Host,
			Method:         r.Method,
			Path:           r.URL.EscapedPath(),
			Query:          r.URL.Query(),
			RequestHeaders: reqHeaders,
		}

		if r.URL.EscapedPath() != "/healthcheck" {
			logger.Info(
				"new request",
				logger.Any("requestSummary", metadata),
			)
		}

		start := time.Now()
		rw := &ResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		resHeaders := []string{}
		for header, values := range rw.Header() {
			for _, value := range values {
				resHeaders = append(resHeaders, header+": "+value)
			}
		}

		metadata.StatusCode = rw.StatusCode
		metadata.ResponseHeaders = resHeaders
		metadata.Duration = time.Since(start)

		if r.URL.EscapedPath() != "/healthcheck" {
			logger.Info(
				"request summary",
				logger.Any("requestSummary", metadata),
			)
		}
	})
}
