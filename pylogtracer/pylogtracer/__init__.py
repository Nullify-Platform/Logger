"""
PyLogTracer -- Structured logging and optional tracing for Python services.

Core (zero external dependencies):
    - ``JSONLogFormatter``: stdlib ``logging.Formatter`` producing Go zap-compatible JSON
    - ``configure_logging``: one-call root logger setup (JSON in prod, text in dev)

Optional extras (install via ``pip install pylogtracer[tracing]``, etc.):
    - ``get_structured_logger()``: loguru-based structured logger (requires ``[loguru]``)
    - ``get_tracer()``: OpenTelemetry tracer (requires ``[tracing]``)
    - ``initialize_tracer()``: OpenTelemetry tracer setup (requires ``[tracing]``)
"""

# Always available -- stdlib only
from .config import configure_logging
from .formatter import JSONLogFormatter

__all__ = [
    "JSONLogFormatter",
    "configure_logging",
    "get_structured_logger",
    "get_tracer",
]


def get_structured_logger():
    """Return the loguru-based structured logger.

    Requires the ``[loguru]`` extra::

        pip install pylogtracer[loguru]
    """
    from .logger import structured_logger

    return structured_logger


def get_tracer():
    """Return an initialised OpenTelemetry tracer.

    Requires the ``[tracing]`` extra::

        pip install pylogtracer[tracing]
    """
    from .tracing_setup import initialize_tracer

    return initialize_tracer()
