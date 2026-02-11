package tests

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/nullify-platform/logger/pkg/logger/meter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

func TestDevelopmentLogger(t *testing.T) {
	ctx := context.Background()
	var output bytes.Buffer

	ctx, err := logger.ConfigureDevelopmentLogger(ctx, "info", &output)
	require.Nil(t, err)
	log := logger.L(ctx)

	log.Info("test")
	log.Sync()

	fmt.Println("stdout: " + output.String())

	assert.True(t, strings.Contains(output.String(), "INFO"), "stdout didn't include INFO")
	assert.True(t, strings.Contains(output.String(), "test"), "stdout didn't include the 'test' log message")
	assert.True(t, strings.Contains(output.String(), "tests/development_test.go:25"), "stdout didn't include the file path and line number")
	assert.True(t, strings.Contains(output.String(), `{"version": "0.0.0"}`), "stdout didn't include version")
}

func TestDevelopmentLoggerMeterAvailable(t *testing.T) {
	ctx := context.Background()
	var output bytes.Buffer

	ctx, err := logger.ConfigureDevelopmentLogger(ctx, "info", &output)
	require.NoError(t, err)

	m := meter.FromContext(ctx)
	require.NotNil(t, m, "meter should be available from context after ConfigureDevelopmentLogger")

	counter, err := m.Int64Counter("test.counter", metric.WithDescription("test counter"))
	require.NoError(t, err)

	counter.Add(ctx, 1)
}
