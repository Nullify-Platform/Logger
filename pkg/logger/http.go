package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/nullify-platform/logger/pkg/logger/tracer"
	"go.uber.org/zap"
)

// NewLoggingTransport creates a new http.RoundTripper that logs requests and responses
// with the global logge
func NewLoggingTransport(baseCtx context.Context, baseTransport http.RoundTripper, service string) http.RoundTripper {
	return &LoggingTransport{
		baseTransport: baseTransport,
		service:       service,
		ctx:           baseCtx,
	}
}

// LoggingTransport is an http.RoundTripper that logs HTTP requests and responses
type LoggingTransport struct {
	baseTransport http.RoundTripper
	logger        Logger
	service       string
	ctx           context.Context
}

// RoundTrip executes the HTTP request and logs the request and response summary
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, span := tracer.F(t.ctx).Start(t.ctx, req.URL.EscapedPath())
	defer F(ctx).Sync()
	defer span.End()

	start := time.Now()

	res, err := t.baseTransport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	reqHeaders := []string{}
	for header, values := range req.Header {
		for _, value := range values {
			reqHeaders = append(reqHeaders, header+": "+value)
		}
	}

	resHeaders := []string{}
	for header, values := range res.Header {
		for _, value := range values {
			resHeaders = append(resHeaders, header+": "+value)
		}
	}

	summary := Any("requestSummary", &HTTPRequestSummary{
		Service:         t.service,
		Method:          req.Method,
		URL:             req.URL.String(),
		StatusCode:      res.StatusCode,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: resHeaders,
		Duration:        time.Since(start),
	})

	if t.logger == nil {
		F(ctx).Info("request summary", summary)
	} else {
		t.logger.Info("request summary", summary)
	}

	return res, err
}

// HTTPRequestSummary is a JSON struct definition for the summary of an HTTP request
type HTTPRequestSummary struct {
	Service         string        `json:"service"`
	Host            string        `json:"host"`
	Method          string        `json:"method"`
	URL             string        `json:"url"`
	StatusCode      int           `json:"statusCode"`
	RequestHeaders  []string      `json:"requestHeaders"`
	ResponseHeaders []string      `json:"responseHeaders"`
	Duration        time.Duration `json:"duration"`
}

// HTTPRequest logs a summary of an HTTP request
// service is the name of the service the request is being made to
func (l *logger) HTTPRequest(service string, duration time.Duration, req *http.Request, res *http.Response) {
	reqHeaders := []string{}
	for header, values := range req.Header {
		for _, value := range values {
			reqHeaders = append(reqHeaders, header+": "+value)
		}
	}

	resHeaders := []string{}
	for header, values := range res.Header {
		for _, value := range values {
			resHeaders = append(resHeaders, header+": "+value)
		}
	}

	l.Info(
		"request summary",
		Any("httpSummary", &HTTPRequestSummary{
			Service:         service,
			Method:          req.Method,
			URL:             req.URL.String(),
			StatusCode:      res.StatusCode,
			RequestHeaders:  reqHeaders,
			ResponseHeaders: resHeaders,
			Duration:        duration,
		}),
	)
}

// HTTPRequest logs a summary of an HTTP request
// service is the name of the service the request is being made to
func HTTPRequest(service string, duration time.Duration, req *http.Request, res *http.Response) {
	reqHeaders := []string{}
	for header, values := range req.Header {
		for _, value := range values {
			reqHeaders = append(reqHeaders, header+": "+value)
		}
	}

	resHeaders := []string{}
	for header, values := range res.Header {
		for _, value := range values {
			resHeaders = append(resHeaders, header+": "+value)
		}
	}

	zap.L().Info(
		"request summary",
		Any("httpSummary", &HTTPRequestSummary{
			Service:         service,
			Host:            req.Host,
			Method:          req.Method,
			URL:             req.URL.String(),
			StatusCode:      res.StatusCode,
			RequestHeaders:  reqHeaders,
			ResponseHeaders: resHeaders,
			Duration:        duration,
		}),
	)
}
