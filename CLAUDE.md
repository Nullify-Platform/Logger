# CLAUDE.md

## What is this?

A shared Go logging library (`github.com/nullify-platform/logger`) used across Nullify services. It wraps Uber Zap with context-based logging, OpenTelemetry tracing/metrics, and HTTP middleware. There is also a parallel Python implementation in `pylogtracer/`.

## Commands

```bash
make build        # Build (CGO_ENABLED=0, linux/amd64)
make unit         # Run all Go tests
make cov          # Tests with coverage (generates coverage.html, coverage.xml, coverage.txt)
make lint         # golangci-lint v2.10.1 (gofmt, stylecheck, gosec)
make format       # gofmt

# Run a single test
go test ./pkg/logger/... -run TestName -v

# Python
make lint-python  # ruff format + lint check
make fix-python   # ruff auto-fix
```

## Architecture

All logging is context-bound (`logger.L(ctx)`), never global. `ConfigureProductionLogger` / `ConfigureDevelopmentLogger` inject the logger into context and set up OTEL providers.

- **`pkg/logger/`** - Core package: `Logger` interface, configuration, field builders, chunking for oversized Loki entries
- **`pkg/logger/tracer/`** - OTEL tracer wrapping: span creation, AWS service context extraction (SQS, SNS, Lambda)
- **`pkg/logger/meter/`** - OTEL meter provider context management
- **`pkg/logger/middleware/`** - HTTP middleware for request logging, tracing, and metrics (redacts sensitive headers, skips healthchecks)

### Key patterns

- `logger.L(ctx)` auto-injects trace-id and span-id into every log line and records errors on the current span
- Field builders in `fields.go` provide both direct helpers (`logger.String()`, `logger.Err()`) and a fluent `LogFields` builder for domain-specific fields (agent, repository, tool calls)
- OTEL env vars (`OTEL_SERVICE_NAME`, `OTEL_RESOURCE_ATTRIBUTES`) are propagated as default log fields in addition to trace/metric resources
- Metrics use cumulative temporality (required by Grafana Mimir)
- OTLP headers can be fetched from AWS SSM Parameter Store via `OTEL_EXPORTER_OTLP_HEADERS_NAME`

### Spans (tracer sub-package)

Both tracer and meter are automatically injected into context by the logger configuration functions.

```go
ctx, span := tracer.StartNewSpan(ctx, "operation-name")  // child span
defer span.End()

// or

ctx, span := tracer.FromContext(ctx).Start(ctx, "operation-name")
defer span.End()


ctx, span := tracer.StartNewRootSpan(ctx, "handler")  // root span (no parent)
defer span.End()

tracer.ForceFlush(ctx)                                // flush before shutdown - also done by logger.L(ctx).Sync()
```

### Metrics (meter sub-package)

```go
m := meter.FromContext(ctx)
counter, _ := m.Int64Counter("requests.total")
counter.Add(ctx, 1, metric.WithAttributes(attribute.String("method", "GET")))
meter.ForceFlush(ctx) // flush before shutdown - also done by logger.L(ctx).Sync()
```

## Testing

Tests use `stretchr/testify`. Logger output tests capture to a `bytes.Buffer` writer and verify JSON structure.
