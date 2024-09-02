package tracer

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// CreateClientContextForLambdaInvoke creates a client context string pointer which causes the traceparent header to be sent to the lambda function.
// eg:
//
//	_, err = awslambda.NewFromConfig(awsConfig).Invoke(ctx, &awslambda.InvokeInput{
//		   FunctionName:  aws.String(functionName),
//		   Payload:       payload,
//		   ClientContext: tracer.CreateClientContextForLambdaInvoke(ctx, nil),
//	})
//
// `custom` can be nil, an empty map or include arbitrary key-value pairs.
// The returned string is base64 encoded JSON with the following structure:
//
//	ClientContext: {
//	  "Client": {
//	    "installation_id":"",
//	    "app_title":"",
//	    "app_version_code":"",
//	    "app_package_name":""
//	  },
//	  "env": null,
//	  "custom": {
//	    "traceparent":"00-{traceID}-{spanID}-01"
//	  }
//	}
func CreateClientContextForLambdaInvoke(ctx context.Context, custom map[string]string) *string {
	if custom == nil {
		custom = map[string]string{}
	}
	InjectTracingIntoCustomMessage(ctx, custom)

	var clientContextBase64 *string
	if marshalled, err := json.Marshal(map[string]interface{}{"custom": custom}); err == nil {
		clientContextJSON := marshalled
		clientContextBase64 = aws.String(base64.StdEncoding.EncodeToString(clientContextJSON))
	}
	return clientContextBase64
}

// ExtractTracingFromLambdaInvokeClientContext extracts the tracing information from the client context.
func ExtractTracingFromLambdaInvokeClientContext(ctx context.Context) context.Context {
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		if len(lc.ClientContext.Custom) != 0 {
			ctx = ExtractTracingFromCustomEventMessage(ctx, lc.ClientContext.Custom)
		}
	}

	return ctx
}
