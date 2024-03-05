package logger

import (
	"go.uber.org/zap"
)

// Option an enum type for logger options
type Option = zap.Option

// Options

// AddCallerSkip adds n to the number of callers skipped by the logger
func AddCallerSkip(n int) Option {
	return zap.AddCallerSkip(n)
}
