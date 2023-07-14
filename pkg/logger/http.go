package logger

import (
	"net/http"

	"go.uber.org/zap"
)

func NewLoggingTransport(baseTransport http.RoundTripper) http.RoundTripper {
	return &LoggingTransport{
		baseTransport: baseTransport,
	}
}

func NewLoggingTransportWithLogger(baseTransport http.RoundTripper, logger Logger) http.RoundTripper {
	return &LoggingTransport{
		logger:        logger,
		baseTransport: baseTransport,
	}
}

type LoggingTransport struct {
	logger        Logger
	baseTransport http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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

	if t.logger == nil {
		Info(
			"request summary",
			Any("requestSummary", &HTTPRequestMetadata{
				Service:         "cognito",
				Method:          req.Method,
				URL:             req.URL.String(),
				StatusCode:      res.StatusCode,
				RequestHeaders:  reqHeaders,
				ResponseHeaders: resHeaders,
			}),
		)
	} else {
		t.logger.Info(
			"request summary",
			Any("requestSummary", &HTTPRequestMetadata{
				Service:         "cognito",
				Method:          req.Method,
				URL:             req.URL.String(),
				StatusCode:      res.StatusCode,
				RequestHeaders:  reqHeaders,
				ResponseHeaders: resHeaders,
			}),
		)
	}

	return res, err
}

type HTTPRequestMetadata struct {
	Service         string   `json:"service"`
	Method          string   `json:"method"`
	URL             string   `json:"url"`
	StatusCode      int      `json:"statusCode"`
	RequestHeaders  []string `json:"requestHeaders"`
	ResponseHeaders []string `json:"responseHeaders"`
}

// HTTPRequest logs a summary of an HTTP request
// service is the name of the service the request is being made to
func (l *logger) HTTPRequest(service string, req *http.Request, res *http.Response) {
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
		Any("httpSummary", &HTTPRequestMetadata{
			Service:         service,
			Method:          req.Method,
			URL:             req.URL.String(),
			StatusCode:      res.StatusCode,
			RequestHeaders:  reqHeaders,
			ResponseHeaders: resHeaders,
		}),
	)
}

// HTTPRequest logs a summary of an HTTP request
// service is the name of the service the request is being made to
func HTTPRequest(service string, req *http.Request, res *http.Response) {
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
		Any("httpSummary", &HTTPRequestMetadata{
			Service:         service,
			Method:          req.Method,
			URL:             req.URL.String(),
			StatusCode:      res.StatusCode,
			RequestHeaders:  reqHeaders,
			ResponseHeaders: resHeaders,
		}),
	)
}
