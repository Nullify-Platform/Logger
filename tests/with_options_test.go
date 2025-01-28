package tests

import (
	"bytes"
	"context"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/require"
)

// TestAddField tests that the logger.AddField function adds a new
// field to the default logger
func TestWithOptions(t *testing.T) {
	ctx := context.Background()
	var output bytes.Buffer

	// create new production logger
	ctx, _, err := logger.ConfigureProductionLogger(ctx, "test", "info", &output)
	require.Nil(t, err)
	myLogger := logger.L(ctx)

	myLogger.NewChild().WithOptions(logger.AddCallerSkip(5))
}
