# PyLogTrace - Structured Logging and Tracing for Python

## Overview
PyLogTrace is a Python logging and tracing library that wraps `loguru` for structured logging and integrates with OpenTelemetry to send traces to Grafana Tempo.

## Features
- JSON structured logging via `loguru`
- OpenTelemetry tracing with automatic span context propagation
- Supports external tracing backends like Grafana Tempo and Jaeger
- Configurable via environment variables

## Installation
```bash
pip install git+ssh://git@github.com/your-org/pylogtrace.git
