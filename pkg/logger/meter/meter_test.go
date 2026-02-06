package meter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// newTestProvider creates a MeterProvider with a ManualReader for deterministic test collection.
func newTestProvider() (*sdkmetric.MeterProvider, *sdkmetric.ManualReader) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return mp, reader
}

func TestNewContextAndFromContext(t *testing.T) {
	mp, _ := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")

	m := FromContext(ctx)
	require.NotNil(t, m)
}

func TestFromContextReturnsNilWhenNotSet(t *testing.T) {
	m := FromContext(t.Context())
	assert.Nil(t, m)
}

func TestCopyFromContext(t *testing.T) {
	mp, _ := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	srcCtx := NewContext(t.Context(), mp, "test-meter")
	dstCtx := CopyFromContext(srcCtx, t.Context())

	m := FromContext(dstCtx)
	require.NotNil(t, m)
}

func TestForceFlush(t *testing.T) {
	mp, _ := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")

	err := ForceFlush(ctx)
	assert.NoError(t, err)
}

func TestForceFlushWithNoProvider(t *testing.T) {
	err := ForceFlush(t.Context())
	assert.NoError(t, err)
}

func TestShutdown(t *testing.T) {
	mp, _ := newTestProvider()
	ctx := NewContext(t.Context(), mp, "test-meter")

	err := Shutdown(ctx)
	assert.NoError(t, err)
}

func TestShutdownWithNoProvider(t *testing.T) {
	err := Shutdown(t.Context())
	assert.NoError(t, err)
}

func TestForceFlushWithReplacedTimeout(t *testing.T) {
	mp, _ := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")

	err := ForceFlushWithReplacedTimeout(ctx, 5*time.Second)
	assert.NoError(t, err)
}

func TestCounterMetricCollection(t *testing.T) {
	mp, reader := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")
	m := FromContext(ctx)

	counter, err := m.Int64Counter("test.requests",
		metric.WithDescription("Total test requests"),
	)
	require.NoError(t, err)

	counter.Add(ctx, 5, metric.WithAttributes(attribute.String("method", "GET")))
	counter.Add(ctx, 3, metric.WithAttributes(attribute.String("method", "POST")))
	counter.Add(ctx, 2, metric.WithAttributes(attribute.String("method", "GET")))

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	require.NoError(t, err)

	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)

	collected := rm.ScopeMetrics[0].Metrics[0]
	assert.Equal(t, "test.requests", collected.Name)
	assert.Equal(t, "Total test requests", collected.Description)

	sum, ok := collected.Data.(metricdata.Sum[int64])
	require.True(t, ok)
	assert.True(t, sum.IsMonotonic)
	assert.Len(t, sum.DataPoints, 2)

	dpByMethod := make(map[string]int64)
	for _, dp := range sum.DataPoints {
		method, _ := dp.Attributes.Value(attribute.Key("method"))
		dpByMethod[method.AsString()] = dp.Value
	}
	assert.Equal(t, int64(7), dpByMethod["GET"])
	assert.Equal(t, int64(3), dpByMethod["POST"])
}

func TestHistogramMetricCollection(t *testing.T) {
	mp, reader := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")
	m := FromContext(ctx)

	histogram, err := m.Float64Histogram("test.duration",
		metric.WithDescription("Request duration"),
		metric.WithUnit("s"),
	)
	require.NoError(t, err)

	histogram.Record(ctx, 0.1)
	histogram.Record(ctx, 0.5)
	histogram.Record(ctx, 1.2)

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	require.NoError(t, err)

	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)

	collected := rm.ScopeMetrics[0].Metrics[0]
	assert.Equal(t, "test.duration", collected.Name)
	assert.Equal(t, "s", collected.Unit)

	hist, ok := collected.Data.(metricdata.Histogram[float64])
	require.True(t, ok)
	require.Len(t, hist.DataPoints, 1)
	assert.Equal(t, uint64(3), hist.DataPoints[0].Count)
	assert.InDelta(t, 1.8, hist.DataPoints[0].Sum, 0.001)
}

func TestGaugeMetricCollection(t *testing.T) {
	mp, reader := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	ctx := NewContext(t.Context(), mp, "test-meter")
	m := FromContext(ctx)

	gauge, err := m.Float64Gauge("test.temperature",
		metric.WithDescription("Current temperature"),
	)
	require.NoError(t, err)

	gauge.Record(ctx, 20.5)
	gauge.Record(ctx, 22.3)

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	require.NoError(t, err)

	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)

	collected := rm.ScopeMetrics[0].Metrics[0]
	assert.Equal(t, "test.temperature", collected.Name)

	g, ok := collected.Data.(metricdata.Gauge[float64])
	require.True(t, ok)
	require.Len(t, g.DataPoints, 1)
	assert.InDelta(t, 22.3, g.DataPoints[0].Value, 0.001)
}

func TestCopiedContextRetainsMetrics(t *testing.T) {
	mp, reader := newTestProvider()
	defer func() { _ = mp.Shutdown(t.Context()) }()

	srcCtx := NewContext(t.Context(), mp, "test-meter")
	dstCtx := CopyFromContext(srcCtx, t.Context())

	m := FromContext(dstCtx)
	require.NotNil(t, m)

	counter, err := m.Int64Counter("test.copied")
	require.NoError(t, err)

	counter.Add(dstCtx, 42)

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	require.NoError(t, err)

	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)

	sum, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Sum[int64])
	require.True(t, ok)
	require.Len(t, sum.DataPoints, 1)
	assert.Equal(t, int64(42), sum.DataPoints[0].Value)
}
