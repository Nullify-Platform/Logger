# Nullify Logging Library

Abstraction above the Uber Zap logging library so we aren't locked into a particular logger.
Provides some standard logging practices across all Nullify services.

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
