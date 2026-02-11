package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestChunkString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		chunkSize int
		expected  []string
	}{
		{
			name:      "under limit returns single chunk",
			input:     "hello",
			chunkSize: 100,
			expected:  []string{"hello"},
		},
		{
			name:      "exact limit returns single chunk",
			input:     "hello",
			chunkSize: 5,
			expected:  []string{"hello"},
		},
		{
			name:      "over limit splits into chunks",
			input:     "abcdefghij",
			chunkSize: 3,
			expected:  []string{"abc", "def", "ghi", "j"},
		},
		{
			name:      "empty string returns single chunk",
			input:     "",
			chunkSize: 100,
			expected:  []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chunkString(tt.input, tt.chunkSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChunkOversizedFields_NoChunking(t *testing.T) {
	fields := []zapcore.Field{
		zap.String("msg", "short message"),
		zap.Int("status", 200),
	}

	result := chunkOversizedFields(fields)
	assert.Nil(t, result, "should return nil when no fields exceed limit")
}

func TestChunkOversizedFields_SingleOversizedField(t *testing.T) {
	largeBody := strings.Repeat("x", MaxStringFieldSize*2+100)

	fields := []zapcore.Field{
		zap.String("status", "error"),
		zap.String("body", largeBody),
		zap.Int("code", 500),
	}

	chunks := chunkOversizedFields(fields)
	require.NotNil(t, chunks)
	assert.Len(t, chunks, 3, "should produce 3 chunks")

	// Each chunk should contain the base fields + chunk of body + metadata
	for i, chunkFields := range chunks {
		fieldMap := fieldsToMap(chunkFields)

		// Base fields present in every chunk
		assert.Equal(t, "error", fieldMap["status"])
		assert.Equal(t, int64(500), fieldMap["code"])

		// Chunk metadata
		assert.Equal(t, int64(i+1), fieldMap["body_chunk"])
		assert.Equal(t, int64(3), fieldMap["body_total_chunks"])

		// Body chunk is present and within limit
		bodyChunk := fieldMap["body"].(string)
		assert.LessOrEqual(t, len(bodyChunk), MaxStringFieldSize)
	}

	// Reassemble and verify full content
	var reassembled strings.Builder
	for _, chunkFields := range chunks {
		fieldMap := fieldsToMap(chunkFields)
		reassembled.WriteString(fieldMap["body"].(string))
	}
	assert.Equal(t, largeBody, reassembled.String())
}

func TestChunkOversizedFields_MultipleOversized(t *testing.T) {
	// First oversized field gets chunked, second gets truncated
	largeBody := strings.Repeat("a", MaxStringFieldSize+500)
	largeResponse := strings.Repeat("b", MaxStringFieldSize+300)

	fields := []zapcore.Field{
		zap.String("body", largeBody),
		zap.String("response", largeResponse),
	}

	chunks := chunkOversizedFields(fields)
	require.NotNil(t, chunks)
	assert.Len(t, chunks, 2, "first field splits into 2 chunks")

	// The second oversized field (response) should be truncated in each chunk
	for _, chunkFields := range chunks {
		fieldMap := fieldsToMap(chunkFields)
		response := fieldMap["response"].(string)
		assert.Contains(t, response, "...[truncated]")
		assert.LessOrEqual(t, len(response), MaxStringFieldSize+len("...[truncated]"))
	}
}

func TestChunkOversizedFields_ExactLimit(t *testing.T) {
	exactBody := strings.Repeat("x", MaxStringFieldSize)

	fields := []zapcore.Field{
		zap.String("body", exactBody),
	}

	result := chunkOversizedFields(fields)
	assert.Nil(t, result, "exactly at limit should not trigger chunking")
}

// fieldsToMap converts a slice of zap fields to a map for easier test assertions.
func fieldsToMap(fields []zapcore.Field) map[string]interface{} {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}
	return enc.Fields
}
