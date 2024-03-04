package tests

import (
	"bytes"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/require"
)

// TestAddField tests that the logger.AddField function adds a new
// field to the default logger
func TestWithOptions(t *testing.T) {
	var output bytes.Buffer

	// create new production logger
	myLogger, err := logger.ConfigureProductionLogger("info", &output)
	require.Nil(t, err)

	myLogger.NewChild().WithOptions(logger.AddCallerSkip(5))
}
