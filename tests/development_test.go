package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevelopmentLogger(t *testing.T) {
	var output bytes.Buffer

	log, err := logger.ConfigureDevelopmentLogger("info", &output)
	require.Nil(t, err)

	logger.Info("test")
	log.Sync()

	fmt.Println("stdout: " + output.String())

	assert.True(t, strings.Contains(output.String(), "INFO"), "stdout didn't include INFO")
	assert.True(t, strings.Contains(output.String(), "test"), "stdout didn't include the 'test' log message")
	assert.True(t, strings.Contains(output.String(), "tests/development_test.go:20"), "stdout didn't include the file path and line number")
	assert.True(t, strings.Contains(output.String(), `{"version": "0.0.0"}`), "stdout didn't include version")
}
