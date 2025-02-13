from loguru import logger
import sys
import json

class StructuredLogger:
    def __init__(self):
        logger.remove()  # Remove default handlers
        logger.add(sys.stdout, format=self.format_json)

    def format_json(self, message):
        record = message.record
        log_data = {
            "timestamp": record["time"].strftime("%Y-%m-%dT%H:%M:%S.%fZ"),
            "level": record["level"].name,
            "message": record["message"],
            "file": record["file"].name,
            "line": record["line"],
            "context": record["extra"].get("context", {}),
        }
        return json.dumps(log_data)

    def log(self, level, msg, **context):
        logger.bind(context=context).log(level.upper(), msg)

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
