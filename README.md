# Nullify Logging Library

Abstraction above the Uber Zap logging library so we aren't locked into a particular logger.
Provides some standard logging practices across all Nullify services.

## How do I use this?

To make the most use of this library, you have to ensure that your service passes context down within each function call.

One big change is that the logger is no longer a global variable. It is now linked to the context, so you need to ensure that you pass the context.

```go
logger.F(ctx).Info("this is a log message")
```

First, declare the background context in your main function, and then inject the logger into the context.

```go
func main() {
  ctx := context.Background()
  ctx, err := logger.ConfigureProductionLogger(ctx, "info") // inject logger into context

  // There is no longer a global logger. It is always linked to the context, to ensure that log messages have the appropriate span and trace IDs linked to them.
  defer logger.F(ctx).Sync() // defer the sync of the logger to ensure all logs are written

  // Initiate a new span, and defer ending the span to record the end of this piece of work.
  ctx, span := tracer.F(ctx).Start(ctx, "main")
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
  // Retrieve the tracer from the context
  ctx, span := tracer.F(ctx).Start(ctx, "some-other-work")
  defer span.End()

  // This will automatically set the parent function span to be the parent span of this new span that has started, within this trace.

  // To record errors, you can use the logger.Fctx).Error() call. This will automatically capture any errors that you pass into it and pass them to GlitchTip. It will also set the span to errored, so it is highlighted in Grafana.

  error := errors.New("this is an error")
  logger.F(err).Error("this is an error", error)
}
```

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
make test # without coverage
make cov  # with coverage
```

## Lint

Run the golangci-lint docker image with the following tools:

- gofmt
- stylecheck
- gosec

```
make lint
```
