# Nullify Logging Library

Abstraction above the Uber Zap logging library so we aren't locked into a particular logger.
Provides some standard logging practices across all Nullify services.

## How do I use this?

To make the most use of this library, you have to ensure that your service passes context down within each function call.

One big change is that the logger is no longer a global variable. It is now linked to the context, so you need to ensure that you pass the context.

```go
logger.L(ctx).Info("this is a log message")
```

First, declare the background context in your main function, and then inject the logger into the context.

```go
func main() {
  ctx := context.Background()
  ctx, err := logger.ConfigureProductionLogger(ctx, "info") // inject logger into context

  // There is no longer a global logger. It is always linked to the context, to ensure that log messages have the appropriate span and trace IDs linked to them.
  defer logger.L(ctx).Sync() // defer the sync of the logger to ensure all logs are written

  // Initiate a new span, and defer ending the span to record the end of this piece of work.
  ctx, span := tracer.FromContext(ctx).Start(ctx, "main")
  defer span.End()

  // You can add more contextual event information to spans through the span.AddEvent method. This is analogous to adding a log message, and has an associated timestamp that is recorded.
  span.AddEvent("a piece of work is starting within the span")

  // You can also set attributes to the span to provide more context to the span, typically as key-value pairs.
  span.SetAttributes(
    attribute.String("customer-id", "018e4fc1-c079-70ce-84a0-9591295d96aa"),
    attribute.Int64("request-count", 80),
  )

  // You need to ensure that you pass context to other functions so that the logger and tracer are available to them.
  anotherFunction(ctx)
}

func anotherFunction(ctx context.Context) {
  // Start a child span - the parent span is automatically linked via context.
  ctx, span := tracer.FromContext(ctx).Start(ctx, "some-other-work")
  defer span.End()

  // To record errors, use logger.L(ctx).Error(). This automatically sets the span to errored, so it is highlighted in Grafana.
  err := errors.New("this is an error")
  logger.L(ctx).Error("this is an error", logger.Err(err))
}
```

### Spans

Spans are created via the `tracer` sub-package. Both the tracer and meter are automatically injected into context by `ConfigureProductionLogger` / `ConfigureDevelopmentLogger`.

```go
// Start a child span (inherits parent from context)
ctx, span := tracer.StartNewSpan(ctx, "span-name")
defer span.End()

// Start a root span (no parent, e.g. at the entry point of an API handler)
ctx, span := tracer.StartNewRootSpan(ctx, "request-handler")
defer span.End()

// With options
ctx, span := tracer.StartNewSpan(ctx, "span-name",
  trace.WithAttributes(attribute.String("key", "value")),
  trace.WithSpanKind(trace.SpanKindServer),
)
defer span.End()

// Flush traces before shutdown (e.g. in Lambda handlers)
tracer.ForceFlush(ctx)
```

### Metrics

Metrics are created via the `meter` sub-package. The meter is retrieved from context.

```go
m := meter.FromContext(ctx)

// Create instruments
counter, _ := m.Int64Counter("requests.total")
histogram, _ := m.Float64Histogram("request.duration_ms")

// Record measurements
counter.Add(ctx, 1, metric.WithAttributes(attribute.String("method", "GET")))
histogram.Record(ctx, 42.5)

// Flush metrics before shutdown
meter.ForceFlush(ctx)
```

## OpenTelemetry Exporting

To actually have your traces exported, you need to set a few environment variables in your service:

- `OTEL_EXPORTER_OTLP_PROTOCOL`: typically set to `http/protobuf`; this is the protocol that the traces are sent over.
- `OTEL_EXPORTER_OTLP_ENDPOINT`: the endpoint that the traces are sent to.
- `OTEL_EXPORTER_OTLP_HEADERS_NAME`: the name of the parameter in aws parameter store that contains the headers for the OTLP exporter.
- `OTEL_RESOURCE_ATTRIBUTES`: comma-separated `key=value` attributes associated with the service (e.g. `deployment.environment=production`). These are propagated to traces, metrics, and as default log fields.
- `OTEL_SERVICE_NAME`: the name of the service. Propagated to traces, metrics, and as a default `service.name` log field.

## Install

Dependencies:

- make
- golang

Install code dependencies

```
go mod download
```

## Build

Compile the code locally into the `bin` directory

```
make
```

## Test

```
make unit # without coverage
make cov  # with coverage
```

## Lint

Run golangci-lint with the following tools:

- gofmt
- stylecheck
- gosec

```
make lint
```
