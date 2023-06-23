package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

// Fields

func Trace(trace []byte) Field {
	return zap.ByteString("trace", trace)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

func Err(err error) Field {
	return zap.Error(err)
}

func Errs(msg string, errs []error) Field {
	return zap.Errors(msg, errs)
}

func String(key string, val string) Field {
	return zap.String(key, val)
}

func Strings(key string, val []string) Field {
	return zap.Strings(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func Bools(key string, val []bool) Field {
	return zap.Bools(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Ints(key string, val []int) Field {
	return zap.Ints(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Int64s(key string, val []int64) Field {
	return zap.Int64s(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

func Durations(key string, val []time.Duration) Field {
	return zap.Durations(key, val)
}
