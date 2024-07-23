package logger

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
)

func TestFormatLogsURL(t *testing.T) {
	region := "ap-southeast-2"
	logGroupName := "/aws/lambda/foxsports-the-force-pullrequest"
	logStreamName := "2024/07/22/[$LATEST]83b2196b0fb5466e9e71f9e6c9eceab8"

	expectedURL := "https://ap-southeast-2.console.aws.amazon.com/cloudwatch/home?region=ap-southeast-2#logsV2:log-groups/log-group/" +
		"$252Faws$252Flambda$252Ffoxsports-the-force-pullrequest/log-events/" +
		"2024$252F07$252F22$252F$255B$2524LATEST$255D83b2196b0fb5466e9e71f9e6c9eceab8"

	result := formatLogsURL(region, logGroupName, logStreamName)

	require.Equal(t, expectedURL, result)
}

func TestIntegrationCaptureException(t *testing.T) {
	t.Skip("Debug manually & check logs")
	t.Setenv("SENTRY_DSN", "https://d678cbc547e946d38a483bd3651fadea@app.glitchtip.com/6485")
	t.Setenv("AWS_REGION", "localhost")

	initialiseSentry()

	// fix the "missing mechanism type" error
	fixMechanismTypeInSentryEvents()

	// add tags to the sentry events
	addTagsToSentryEvents("logger_test.go", map[string]string{
		"Environment": "integration-test",
		"Service":     "logger",
		"Tenant":      "nullify-integration-test",
	})

	err := testFunctionA()
	fields := []Field{Err(err)}

	// when
	l := logger{}
	l.captureExceptions(fields)

	// need to manually check the logs for:
	// Sending event failed with the following error: {"detail": [{"type": "missing", "loc": ["body", "payload", "exception", "list[function-after[check_type_value(), function-wrap[_run_root_validator()]]]", 0, "mechanism", "type"], "msg": "Field required"},
	sentry.Flush(30 * time.Second)
}

func testFunctionA() error {
	return errors.WithStack(testFunctionB())
}

func testFunctionB() error {
	return errors.WithStack(testFunctionC())
}

func testFunctionC() error {
	return errors.New("test error")
}
