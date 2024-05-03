package logger

import (
	"context"
	"net/http"
	"strings"
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
	ctx, span := tracer.FromContext(t.ctx).Start(t.ctx, req.URL.EscapedPath())
	defer L(ctx).Sync()
	defer span.End()

	start := time.Now()

	res, err := t.baseTransport.RoundTrip(req)
	if err != nil {
		return res, err
	}

	summary := createHttpRequestSummary(t.service, time.Since(start), req, res)
	logger := t.logger
	if logger == nil {
		logger = L(ctx)
	}

	if res.StatusCode >= 500 {
		logger.Warn("request summary", summary)
	} else if res.StatusCode >= 400 {
		logger.Info("request summary", summary)
	} else {
		logger.Debug("request summary", summary)
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
	httpSummary := createHttpRequestSummary(service, duration, req, res)

	if res.StatusCode >= 500 {
		l.Warn("request summary", httpSummary)
	} else if res.StatusCode >= 400 {
		l.Info("request summary", httpSummary)
	} else {
		l.Debug("request summary", httpSummary)
	}
}

// HTTPRequest logs a summary of an HTTP request
// service is the name of the service the request is being made to
func HTTPRequest(service string, duration time.Duration, req *http.Request, res *http.Response) {
	httpSummary := createHttpRequestSummary(service, duration, req, res)

	if res.StatusCode >= 500 {
		zap.L().Warn("request summary", httpSummary)
	} else if res.StatusCode >= 400 {
		zap.L().Info("request summary", httpSummary)
	} else {
		zap.L().Debug("request summary", httpSummary)
	}
}

func createHttpRequestSummary(service string, duration time.Duration, req *http.Request, res *http.Response) zap.Field {
	reqHeaders := []string{}
	for header, values := range req.Header {
		if strings.ToLower(header) == "authorization" {
			reqHeaders = append(reqHeaders, header+": ****")
			continue
		}

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

	return Any("requestSummary", &HTTPRequestSummary{
		Service:         service,
		Host:            req.Host,
		Method:          req.Method,
		URL:             req.URL.String(),
		StatusCode:      res.StatusCode,
		RequestHeaders:  reqHeaders,
		ResponseHeaders: resHeaders,
		Duration:        duration,
	})
}
