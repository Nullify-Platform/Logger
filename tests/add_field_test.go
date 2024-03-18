package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/require"
)

// TestAddField tests that the logger.AddField function adds a new
// field to the default logger
func TestAddField(t *testing.T) {
	ctx := context.Background()
	var output bytes.Buffer

	// create new production logger
	ctx, err := logger.ConfigureProductionLogger(ctx, "info", &output)
	require.Nil(t, err)
	myLogger := logger.F(ctx)

	// log a line without the added field
	myLogger.Info("test")
	myLogger.Sync()
	fmt.Println("stdout: " + output.String())

	// check that the output doesn't include the added field
	var jsonOutput map[string]interface{}
	err = json.Unmarshal(output.Bytes(), &jsonOutput)
	require.Nil(t, err)
	_, ok := jsonOutput["my"]
	require.False(t, ok, "stdout included new field when not expected")

	// reset output and log again with the new default field
	output = bytes.Buffer{}
	myLogger.AddFields(logger.String("my", "field"))
	myLogger.Info("test")
	myLogger.Sync()
	fmt.Println("stdout: " + output.String())

	// check that the output now has the added field
	err = json.Unmarshal(output.Bytes(), &jsonOutput)
	require.Nil(t, err)
	_, ok = jsonOutput["my"]
	require.True(t, ok, "stdout didn't include the new field")
}
