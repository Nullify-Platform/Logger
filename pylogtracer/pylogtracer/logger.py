import json
import sys

from loguru import logger
from opentelemetry import trace


class StructuredLogger:
    def __init__(self):
        logger.remove()  # Remove default handlers
        logger.add(
            sys.stdout,
            format="{time} {level} {message} {file} {line} {function}",
            serialize=True,
        )

    def get_current_span_context(self):
        current_span = trace.get_current_span()
        if current_span:
            span_context = current_span.get_span_context()
            return {
                "trace_id": format(span_context.trace_id, "032x"),
                "span_id": format(span_context.span_id, "016x"),
            }
        return {"trace_id": "", "span_id": ""}

    def format_json(self, record):
        log_data = {
            "timestamp": record["time"].isoformat(),
            "level": record["level"].name,
            "message": record["message"],
            "file": record["file"].path,
            "line": record["line"],
            "function": record["function"],
            "module": record["module"],
            "process": record["process"].id,
            "thread": record["thread"].id,
            "trace_id": record["extra"].get("context", {}).get("trace_id", ""),
            "span_id": record["extra"].get("context", {}).get("span_id", ""),
            "context": record["extra"].get("context", {}),
        }
        return json.dumps(log_data) + "\n"

    def log(self, level, msg, **context):
        span_context = self.get_current_span_context()
        context.update(span_context)
        logger.opt(depth=2).bind(context=context).log(level.upper(), msg)

    def debug(self, msg, **context):
        self.log("DEBUG", msg, **context)

    def info(self, msg, **context):
        self.log("INFO", msg, **context)

    def warn(self, msg, **context):
        self.log("WARNING", msg, **context)

    def error(self, msg, **context):
        self.log("ERROR", msg, **context)

    def fatal(self, msg, **context):
        self.log("CRITICAL", msg, **context)
        sys.exit(1)  # Exit with error


structured_logger = StructuredLogger()
