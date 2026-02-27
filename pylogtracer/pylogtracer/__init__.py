"""
PyLogTracer -- Structured logging and optional tracing for Python services.

Core (zero external dependencies):
    - ``JSONLogFormatter``: stdlib ``logging.Formatter`` producing Go zap-compatible JSON
    - ``configure_logging``: one-call root logger setup (JSON in prod, text in dev)

Optional extras (install via ``pip install pylogtracer[tracing]``, etc.):
    - ``structured_logger``: loguru-based structured logger (requires ``[loguru]``)
    - ``track``: OpenTelemetry span tracing decorator (requires ``[tracing,loguru]``)
    - ``initialize_tracer``: OpenTelemetry tracer setup (requires ``[tracing]``)
"""

# Always available -- stdlib only
from .config import configure_logging
from .formatter import JSONLogFormatter

__all__ = [
    "JSONLogFormatter",
    "configure_logging",
]


def __getattr__(name: str):  # noqa: N807
    """Lazy-load optional modules to avoid import errors when extras are not installed."""
    if name == "structured_logger":
        from .logger import structured_logger

        return structured_logger
    if name == "track":
        from .tracer import track

        return track
    if name == "initialize_tracer":
        from .tracing_setup import initialize_tracer

        return initialize_tracer
    if name == "tracer":
        from .tracing_setup import tracer

        return tracer
    msg = f"module 'pylogtracer' has no attribute {name!r}"
    raise AttributeError(msg)
