from functools import wraps
from opentelemetry import trace
from opentelemetry.propagate import extract
from opentelemetry.context import get_current
from pylogtrace import tracer
from pylogtrace.logger import structured_logger

def trace_span(func):
    """Decorator to trace function execution with OpenTelemetry"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        # Extract trace context from request
        parent_context = extract(kwargs.pop("context", {}))
        session_id = kwargs.pop("session_id", None)  # Optional session grouping

        # Check if a span is already active (nested tracing)
        current_span = trace.get_current_span()
        if current_span.get_span_context().is_valid():
            parent_span = current_span
        else:
            parent_span = None

        with tracer.start_as_current_span(
            func.__name__, context=parent_context, parent=parent_span
        ) as span:
            trace_id = span.get_span_context().trace_id
            span_id = span.get_span_context().span_id

            structured_logger.info(f"Started span {trace_id}, Span ID: {span_id}", session_id=session_id)

            result = func(*args, **kwargs)

            structured_logger.info(f"Ending span {trace_id}, Span ID: {span_id}", session_id=session_id)
            return result
    return wrapper
