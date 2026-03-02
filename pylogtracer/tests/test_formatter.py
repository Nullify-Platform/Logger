"""Tests for pylogtracer.formatter.JSONLogFormatter."""

import json
import logging
import warnings

import pytest
from pylogtracer.formatter import JSONLogFormatter


@pytest.fixture
def formatter():
    return JSONLogFormatter()


def _make_record(
    msg: str = "hello",
    level: int = logging.INFO,
    name: str = "test.logger",
    exc_info: tuple | None = None,
    extra: dict | None = None,
) -> logging.LogRecord:
    record = logging.LogRecord(
        name=name,
        level=level,
        pathname="example.py",
        lineno=42,
        msg=msg,
        args=(),
        exc_info=exc_info,
    )
    if extra:
        for key, value in extra.items():
            setattr(record, key, value)
    return record


class TestJSONOutput:
    def test_produces_valid_json(self, formatter: JSONLogFormatter):
        record = _make_record()
        output = formatter.format(record)
        parsed = json.loads(output)
        assert parsed["msg"] == "hello"
        assert parsed["level"] == "info"
        assert parsed["caller"] == "example.py:42"
        assert parsed["logger"] == "test.logger"
        assert "timestamp" in parsed

    def test_timestamp_format(self, formatter: JSONLogFormatter):
        record = _make_record()
        parsed = json.loads(formatter.format(record))
        ts = parsed["timestamp"]
        assert ts.endswith("Z"), f"Timestamp should end with Z: {ts}"
        # Should match YYYY-MM-DDTHH:MM:SS.mmmZ
        assert len(ts) == 24, f"Unexpected timestamp length: {ts}"

    def test_preserves_extra_fields(self, formatter: JSONLogFormatter):
        record = _make_record(extra={"request_id": "abc-123", "tool": "git_diff"})
        parsed = json.loads(formatter.format(record))
        assert parsed["request_id"] == "abc-123"
        assert parsed["tool"] == "git_diff"

    def test_extra_fields_do_not_override_core(self, formatter: JSONLogFormatter):
        # Core fields like "level" and "caller" cannot be overridden by extras.
        record = _make_record(extra={"level": "override", "caller": "bad:0"})
        parsed = json.loads(formatter.format(record))
        assert parsed["level"] == "info"  # original level preserved
        assert parsed["caller"] == "example.py:42"  # original caller preserved


class TestLevelMapping:
    @pytest.mark.parametrize(
        ("py_level", "expected_zap"),
        [
            (logging.DEBUG, "debug"),
            (logging.INFO, "info"),
            (logging.WARNING, "warn"),
            (logging.ERROR, "error"),
            (logging.CRITICAL, "fatal"),
        ],
        ids=["debug", "info", "warning->warn", "error", "critical->fatal"],
    )
    def test_maps_levels_to_zap_equivalents(
        self,
        formatter: JSONLogFormatter,
        py_level: int,
        expected_zap: str,
    ):
        record = _make_record(level=py_level)
        parsed = json.loads(formatter.format(record))
        assert parsed["level"] == expected_zap


class TestStacktrace:
    def test_includes_stacktrace_on_exception(self, formatter: JSONLogFormatter):
        try:
            raise ValueError("test error")
        except ValueError:
            import sys

            exc_info = sys.exc_info()

        record = _make_record(exc_info=exc_info)
        parsed = json.loads(formatter.format(record))
        assert "stacktrace" in parsed
        assert "ValueError: test error" in parsed["stacktrace"]

    def test_no_stacktrace_without_exception(self, formatter: JSONLogFormatter):
        record = _make_record()
        parsed = json.loads(formatter.format(record))
        assert "stacktrace" not in parsed

    def test_format_exception_no_deprecation_warning(self, formatter: JSONLogFormatter):
        """Verify single-arg form does not emit DeprecationWarning."""
        try:
            raise RuntimeError("boom")
        except RuntimeError:
            import sys

            exc_info = sys.exc_info()

        record = _make_record(exc_info=exc_info)
        with warnings.catch_warnings():
            warnings.simplefilter("error")
            formatter.format(record)  # should not raise
