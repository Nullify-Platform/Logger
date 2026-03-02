"""Tests for pylogtracer.config.configure_logging."""

import logging
import sys

from pylogtracer.config import configure_logging
from pylogtracer.formatter import JSONLogFormatter


class TestConfigureLogging:
    def setup_method(self):
        """Reset root logger between tests."""
        root = logging.getLogger()
        root.handlers.clear()
        root.setLevel(logging.WARNING)
        # Reset langfuse logger level (may be set by previous tests).
        logging.getLogger("langfuse").setLevel(logging.NOTSET)

    def test_prod_mode_uses_json_formatter(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        configure_logging()
        root = logging.getLogger()
        assert len(root.handlers) == 1
        assert isinstance(root.handlers[0].formatter, JSONLogFormatter)

    def test_dev_mode_uses_text_formatter(self, monkeypatch):
        monkeypatch.setenv("DEVELOPMENT_MODE", "true")
        configure_logging()
        root = logging.getLogger()
        assert len(root.handlers) == 1
        assert not isinstance(root.handlers[0].formatter, JSONLogFormatter)

    def test_dev_mode_flag_one(self, monkeypatch):
        monkeypatch.setenv("DEVELOPMENT_MODE", "1")
        configure_logging()
        root = logging.getLogger()
        assert not isinstance(root.handlers[0].formatter, JSONLogFormatter)

    def test_respects_python_log_level_env(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        monkeypatch.setenv("PYTHON_LOG_LEVEL", "DEBUG")
        configure_logging()
        root = logging.getLogger()
        assert root.level == logging.DEBUG

    def test_default_level_is_info(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        monkeypatch.delenv("PYTHON_LOG_LEVEL", raising=False)
        configure_logging()
        root = logging.getLogger()
        assert root.level == logging.INFO

    def test_outputs_to_stderr(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        configure_logging()
        root = logging.getLogger()
        handler = root.handlers[0]
        assert isinstance(handler, logging.StreamHandler)
        assert handler.stream is sys.stderr

    def test_suppresses_langfuse_by_default(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        configure_logging()
        langfuse_logger = logging.getLogger("langfuse")
        assert langfuse_logger.level == logging.ERROR

    def test_does_not_suppress_langfuse_when_disabled(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        configure_logging(suppress_langfuse=False)
        langfuse_logger = logging.getLogger("langfuse")
        assert langfuse_logger.level != logging.ERROR

    def test_clears_existing_handlers(self, monkeypatch):
        monkeypatch.delenv("DEVELOPMENT_MODE", raising=False)
        root = logging.getLogger()
        before = len(root.handlers)
        root.addHandler(logging.StreamHandler())
        root.addHandler(logging.StreamHandler())
        assert len(root.handlers) == before + 2
        configure_logging()
        # configure_logging clears all existing handlers and adds exactly one.
        assert len(root.handlers) == 1
