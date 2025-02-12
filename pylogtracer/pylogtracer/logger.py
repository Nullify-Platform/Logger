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
        }
        return json.dumps(log_data)

    def info(self, msg):
        logger.info(msg)

    def warn(self, msg):
        logger.warning(msg)

    def error(self, msg):
        logger.error(msg)

    def fatal(self, msg):
        logger.critical(msg)
        sys.exit(1)  # Exit with error

structured_logger = StructuredLogger()
