package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionLogger(t *testing.T) {
	ctx := context.Background()
	var output bytes.Buffer

	ctx, err := logger.ConfigureProductionLogger(ctx, "info", &output)
	require.Nil(t, err)
	log := logger.L(ctx)

	log.Info("test")
	log.Sync()

	fmt.Println("stdout: " + output.String())

	var jsonOutput map[string]interface{}
	err = json.Unmarshal(output.Bytes(), &jsonOutput)
	require.Nil(t, err, "stdout was not valid json")

	assert.Equal(t, "info", jsonOutput["level"], "stdout didn't include INFO")
	assert.Equal(t, "test", jsonOutput["msg"], "stdout didn't include the 'test' log message")

	pwd, err := os.Getwd()
	assert.NoError(t, err, "did not get working directory successfully")
	assert.Equal(t, pwd+"/production_test.go:24", jsonOutput["caller"], "stdout didn't include the file path and line number")

	assert.Equal(t, "0.0.0", jsonOutput["version"], "stdout didn't include version")
}
