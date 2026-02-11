// Package meter provides a way to get the meter from the context and to create a new context with a meter.
// It also provides a way to force the meter provider to flush all metrics to the exporter.
package meter

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type (
	meterCtxKey         struct{}
	meterProviderCtxKey struct{}
)

// FromContext returns the meter from the context
func FromContext(ctx context.Context) metric.Meter {
	m, _ := ctx.Value(meterCtxKey{}).(metric.Meter)
	return m
}

// NewContext returns a new context with the given meter provider and a named meter
func NewContext(parent context.Context, mp *sdkmetric.MeterProvider, meterName string) context.Context {
	m := mp.Meter(meterName)
	ctx := context.WithValue(parent, meterProviderCtxKey{}, mp)
	return context.WithValue(ctx, meterCtxKey{}, m)
}

// CopyFromContext copies the meter from the old context to the new context
func CopyFromContext(fromCtx context.Context, toCtx context.Context) context.Context {
	m := fromCtx.Value(meterCtxKey{})
	mp := fromCtx.Value(meterProviderCtxKey{})

	toCtx = context.WithValue(toCtx, meterCtxKey{}, m)
	toCtx = context.WithValue(toCtx, meterProviderCtxKey{}, mp)

	return toCtx
}

// ForceFlush forces the meter provider to flush all metrics to the exporter
func ForceFlush(ctx context.Context) error {
	mp, ok := ctx.Value(meterProviderCtxKey{}).(*sdkmetric.MeterProvider)
	if !ok || mp == nil {
		return nil
	}
	return mp.ForceFlush(ctx)
}

// ForceFlushWithReplacedTimeout forces the meter provider to flush all metrics to the exporter, replacing any timeout on ctx with a new one.
func ForceFlushWithReplacedTimeout(ctx context.Context, timeout time.Duration) error {
	flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	return ForceFlush(flushCtx)
}

// Shutdown shuts down the meter provider, flushing all remaining metrics
func Shutdown(ctx context.Context) error {
	mp, ok := ctx.Value(meterProviderCtxKey{}).(*sdkmetric.MeterProvider)
	if !ok || mp == nil {
		return nil
	}
	return mp.Shutdown(ctx)
}
