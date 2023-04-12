package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Fields

func Trace(trace []byte) zapcore.Field {
	return zap.ByteString("trace", trace)
}

func Err(err error) zapcore.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zapcore.Field {
	return zap.Any(key, val)
}

func String(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zapcore.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zapcore.Field {
	return zap.Int64(key, val)
}

func Duration(key string, val time.Duration) zapcore.Field {
	return zap.Duration(key, val)
}
