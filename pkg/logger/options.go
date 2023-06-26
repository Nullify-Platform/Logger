package logger

import (
	"go.uber.org/zap"
)

type Option = zap.Option

// Options

func AddCallerSkip(n int) zap.Option {
	return zap.AddCallerSkip(n)
}
