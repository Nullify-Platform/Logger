# PyLogTracer

Structured JSON logging and optional OpenTelemetry tracing for Python services.

## Core Features (zero dependencies)

- **`JSONLogFormatter`** — stdlib `logging.Formatter` that outputs one JSON object per line, matching the Go zap logger schema
- **`configure_logging()`** — one-call root logger setup with dev/prod dual mode

## Optional Features

Install extras for additional capabilities:

| Extra | Provides | Dependencies |
|-------|----------|-------------|
| `[tracing]` | OpenTelemetry tracer setup | opentelemetry-api, -sdk, -exporter-otlp |
| `[loguru]` | `StructuredLogger` (loguru-based) | loguru |
| `[all]` | Everything above | all of the above |

## Installation

### Core only (recommended for most services)

```bash
pip install pylogtracer
```

### With tracing support

```bash
pip install "pylogtracer[tracing]"
```

### With all extras

```bash
pip install "pylogtracer[all]"
```

## Usage

### JSON Structured Logging

```python
import logging
from pylogtracer import configure_logging

# Set up once at service startup
configure_logging()

logger = logging.getLogger(__name__)
logger.info("Processing request", extra={"request_id": "abc-123", "tool": "search"})
```

**Production output** (single JSON line):
```json
{"timestamp":"2025-06-01T12:34:56.789Z","level":"info","msg":"Processing request","caller":"app.py:8","logger":"__main__","request_id":"abc-123","tool":"search"}
```

**Development output** (set `DEVELOPMENT_MODE=true`):
```
2025-06-01 12:34:56,789 - __main__ - INFO - Processing request
```

### JSON Field Schema

| Field | Description | Go zap equivalent |
|-------|-------------|-------------------|
| `timestamp` | ISO 8601 UTC with `Z` suffix | `zapcore.ISO8601TimeEncoder` |
| `level` | `debug`, `info`, `warn`, `error`, `fatal` | `zapcore.LowercaseLevelEncoder` |
| `msg` | Log message | `MessageKey: "msg"` |
| `caller` | `filename:lineno` | `zapcore.ShortCallerEncoder` |
| `logger` | Logger name | `NameKey: "logger"` |
| `stacktrace` | Present on exceptions only | `StacktraceKey: "stacktrace"` |

Any `extra={}` dict entries are promoted to top-level JSON keys.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DEVELOPMENT_MODE` | Set to `true` or `1` for human-readable text output | _(JSON mode)_ |
| `PYTHON_LOG_LEVEL` | Log level (`DEBUG`, `INFO`, `WARNING`, `ERROR`) | `INFO` |

## Development

```bash
# Install dev dependencies
uv sync --all-groups

# Run tests
make test-python

# Lint and format
make lint-python
make fix-python
```
