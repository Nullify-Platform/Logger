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

func Err(err error) Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

func String(key string, val string) Field {
	return zap.String(key, val)
}

func Strings(key string, val string) Field {
	return zap.String(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}
