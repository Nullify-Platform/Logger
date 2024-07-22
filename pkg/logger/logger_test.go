package logger

import (
	"testing"

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
