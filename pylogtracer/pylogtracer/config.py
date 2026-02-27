"""
Logging configuration for Python services.

Provides ``configure_logging()`` which sets up the root logger with either
structured JSON output (production) or human-readable text (development).
"""

import logging
import os
import sys

from .formatter import JSONLogFormatter


def configure_logging(
    log_level: str = "INFO",
    enable_http_debug: bool = False,
    suppress_langfuse: bool = True,
) -> None:
    """Configure the root logger for service execution.

    In production (default) logs are emitted as structured JSON matching
    the Go zap schema.  Set the ``DEVELOPMENT_MODE`` env var to ``true``
    or ``1`` to fall back to human-readable text output.

    All output goes to stderr so that processes can write structured data
    to stdout without corruption.

    Args:
        log_level: Default log level (overridden by ``PYTHON_LOG_LEVEL`` env var).
        enable_http_debug: Enable debug logging for urllib3/requests.
        suppress_langfuse: Suppress verbose Langfuse logging.
    """
    level_str = os.getenv("PYTHON_LOG_LEVEL", log_level).upper()
    level = getattr(logging, level_str, logging.INFO)

    dev_mode = os.getenv("DEVELOPMENT_MODE", "").lower() in ("true", "1")

    # Clear any previously-configured handlers so behaviour is deterministic.
    root = logging.getLogger()
    root.handlers.clear()
    root.setLevel(level)

    handler = logging.StreamHandler(sys.stderr)
    handler.setLevel(level)

    if dev_mode:
        handler.setFormatter(
            logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
        )
    else:
        handler.setFormatter(JSONLogFormatter())

    root.addHandler(handler)

    if suppress_langfuse:
        logging.getLogger("langfuse").setLevel(logging.ERROR)

    if enable_http_debug:
        logging.getLogger("urllib3").setLevel(logging.DEBUG)
        logging.getLogger("requests").setLevel(logging.DEBUG)
