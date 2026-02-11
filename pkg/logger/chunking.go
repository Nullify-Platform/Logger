package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MaxStringFieldSize is the maximum size in bytes for a single string field
	// before it gets split across multiple log entries. This prevents oversized
	// log entries from being rejected by log aggregation backends
	// (e.g. Grafana Loki's 256KB per-entry limit).
	MaxStringFieldSize = 204800 // 200KB per chunk
)

// chunkOversizedFields checks if any string field exceeds MaxStringFieldSize.
// If so, it returns multiple sets of fields — one per chunk of the oversized
// field — each annotated with chunk/total metadata. Other oversized string
// fields beyond the first are truncated as a safety net.
// Returns nil if no chunking is needed.
func chunkOversizedFields(fields []zapcore.Field) [][]zapcore.Field {
	// Find the first oversized string field
	oversizedIdx := -1
	for i, f := range fields {
		if f.Type == zapcore.StringType && len(f.String) > MaxStringFieldSize {
			oversizedIdx = i
			break
		}
	}

	if oversizedIdx == -1 {
		return nil
	}

	oversized := fields[oversizedIdx]
	chunks := chunkString(oversized.String, MaxStringFieldSize)
	totalChunks := len(chunks)

	// Build base fields: everything except the oversized field.
	// Truncate any other oversized string fields as a safety measure.
	baseFields := make([]zapcore.Field, 0, len(fields)-1)
	for i, f := range fields {
		if i == oversizedIdx {
			continue
		}
		if f.Type == zapcore.StringType && len(f.String) > MaxStringFieldSize {
			f.String = f.String[:MaxStringFieldSize] + "...[truncated]"
		}
		baseFields = append(baseFields, f)
	}

	result := make([][]zapcore.Field, totalChunks)
	for i, chunk := range chunks {
		entryFields := make([]zapcore.Field, 0, len(baseFields)+3)
		entryFields = append(entryFields, baseFields...)
		entryFields = append(entryFields,
			zap.String(oversized.Key, chunk),
			zap.Int(oversized.Key+"_chunk", i+1),
			zap.Int(oversized.Key+"_total_chunks", totalChunks),
		)
		result[i] = entryFields
	}

	return result
}

// chunkString splits s into pieces of at most chunkSize bytes.
func chunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string
	for len(s) > chunkSize {
		chunks = append(chunks, s[:chunkSize])
		s = s[chunkSize:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}
