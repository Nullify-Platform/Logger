package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionLogger(t *testing.T) {
	var output bytes.Buffer

	log, err := logger.ConfigureProductionLogger("info", &output)
	require.Nil(t, err)

	logger.Info("test")
	log.Sync()

	fmt.Println("stdout: " + output.String())

	var jsonOutput map[string]interface{}
	err = json.Unmarshal(output.Bytes(), &jsonOutput)
	require.Nil(t, err, "stdout was not valid json")

	assert.Equal(t, "info", jsonOutput["level"], "stdout didn't include INFO")
	assert.Equal(t, "test", jsonOutput["msg"], "stdout didn't include the 'test' log message")
	assert.Equal(t, "tests/production_test.go:20", jsonOutput["caller"], "stdout didn't include the file path and line number")
	assert.Equal(t, "0.0.0", jsonOutput["version"], "stdout didn't include version")
}
