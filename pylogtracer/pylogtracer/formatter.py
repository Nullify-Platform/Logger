"""
Structured JSON log formatter compatible with Go's zap logger schema.

Outputs one JSON object per log line so that log aggregation tools can parse,
index, and alert on fields like level, caller, and any extra keys
passed via ``logging.info("msg", extra={...})``.

Only uses stdlib modules -- no external dependencies.
"""

import json
import logging
import traceback
from datetime import datetime, timezone

# Map Python level names to Go zap equivalents so queries on
# ``level="warn"`` or ``level="fatal"`` match Python log output.
_LEVEL_NAME_MAP: dict[str, str] = {
    "WARNING": "warn",
    "CRITICAL": "fatal",
}

# Standard LogRecord attributes that should NOT be forwarded as extras.
_STANDARD_LOG_RECORD_ATTRS: frozenset[str] = frozenset(
    {
        "args",
        "created",
        "exc_info",
        "exc_text",
        "filename",
        "funcName",
        "levelname",
        "levelno",
        "lineno",
        "message",
        "module",
        "msecs",
        "msg",
        "name",
        "pathname",
        "process",
        "processName",
        "relativeCreated",
        "stack_info",
        "taskName",
        "thread",
        "threadName",
    }
)


class JSONLogFormatter(logging.Formatter):
    """Emit each log record as a single JSON line matching the Go zap schema.

    Fields produced:
        timestamp  - ISO 8601 UTC (e.g. ``2025-06-01T12:34:56.789Z``)
        level      - zap-compatible (debug, info, warn, error, fatal)
        msg        - the formatted log message
        caller     - ``filename:lineno``
        logger     - the logger name (``record.name``)
        stacktrace - present only when ``exc_info`` is set

    Any *extra* dict entries that are not standard ``LogRecord`` attributes
    are promoted to top-level JSON keys.
    """

    def format(self, record: logging.LogRecord) -> str:
        dt = datetime.fromtimestamp(record.created, tz=timezone.utc)
        log_entry: dict[str, object] = {
            "timestamp": dt.strftime("%Y-%m-%dT%H:%M:%S.")
            + f"{dt.microsecond // 1000:03d}Z",
            "level": _LEVEL_NAME_MAP.get(record.levelname, record.levelname.lower()),
            "msg": record.getMessage(),
            "caller": f"{record.filename}:{record.lineno}",
            "logger": record.name,
        }

        # Attach stacktrace when exception info is present.
        if record.exc_info and record.exc_info[1] is not None:
            # Single-arg form (Python 3.10+) -- avoids DeprecationWarning
            # from the three-arg form removed in Python 3.14.
            log_entry["stacktrace"] = "".join(
                traceback.format_exception(record.exc_info[1])
            ).rstrip()

        # Forward extra fields as top-level keys.
        for key, value in record.__dict__.items():
            if key not in _STANDARD_LOG_RECORD_ATTRS and key not in log_entry:
                log_entry[key] = value

        return json.dumps(log_entry, default=str)
