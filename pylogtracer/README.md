# PyLogTracer - Structured Logging and Tracing for Python

## Overview
PyLogTracer is a Python logging and tracing library that wraps `loguru` for structured logging and integrates with OpenTelemetry to send traces to Grafana Tempo.

## Features
- JSON structured logging via `loguru`
- OpenTelemetry tracing with automatic span context propagation
- Supports external tracing backends like Grafana Tempo and Jaeger
- Configurable via environment variables

## Installation

### Using uv (recommended)
```bash
# Install uv if you haven't already
curl -LsSf https://astral.sh/uv/install.sh | sh

# Clone the repository
git clone https://github.com/nullify/pylogtracer.git
cd pylogtracer

# Install dependencies
uv sync

# Install in development mode
uv pip install -e .
```

### Using pip
```bash
pip install git+ssh://git@github.com/nullify/pylogtracer.git
```

## Development

### Setup with uv
```bash
# Install development dependencies
uv sync --extra dev

# Run tests
uv run pytest

# Run linting
uv run ruff check .

# Format code
uv run ruff format .
```

### Using traditional pip
```bash
# Install development dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run linting
ruff check .

# Format code
ruff format .
```
