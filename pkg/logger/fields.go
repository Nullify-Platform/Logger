package logger

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field an enum type for log message fields
type Field = zapcore.Field

// Fields

// Trace adds a stac trace field to the logger
func Trace(trace []byte) Field {
	return zap.ByteString("trace", trace)
}

// Any adds a field to the logger
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Err adds an error field to the logger
func Err(err error) Field {
	if os.Getenv("SENTRY_DSN") != "" {
		sentry.CaptureException(err)
	}

	return zap.Error(err)
}

// Errs adds a ;ist of errors field to the logger
func Errs(msg string, errs []error) Field {
	if os.Getenv("SENTRY_DSN") != "" {
		for _, err := range errs {
			sentry.CaptureException(err)
		}
	}

	return zap.Errors(msg, errs)
}

// String adds a string field to the logger
func String(key string, val string) Field {
	return zap.String(key, val)
}

// Strings adds a list of strings field to the logger
func Strings(key string, val []string) Field {
	return zap.Strings(key, val)
}

// Bool adds a bool field to the logger
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Bools adds a list of bools field to the logger
func Bools(key string, val []bool) Field {
	return zap.Bools(key, val)
}

// Int adds an int field to the logger
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Ints adds a list of ints field to the logger
func Ints(key string, val []int) Field {
	return zap.Ints(key, val)
}

// Int32 adds an int32 field to the logger
func Int32(key string, val int32) Field {
	return zap.Int32(key, val)
}

// Int32s adds a list of int32s field to the logger
func Int32s(key string, val []int32) Field {
	return zap.Int32s(key, val)
}

// Int64 adds an int64 field to the logger
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Int64s adds a list of int64s field to the logger
func Int64s(key string, val []int64) Field {
	return zap.Int64s(key, val)
}

// Duration adds a time duration field to the logger
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Durations adds a list of time durations field to the logger
func Durations(key string, val []time.Duration) Field {
	return zap.Durations(key, val)
}
