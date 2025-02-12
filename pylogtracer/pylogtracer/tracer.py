from functools import wraps
from opentelemetry.propagate import extract
from pylogtracer.logger import structured_logger

# Import tracer initialized in __init__.py
from pylogtracer import tracer

def trace_span(func):
    """Decorator to trace function execution with OpenTelemetry"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        parent_context = extract(kwargs.pop("context", {}))  # Extract trace context
        with tracer.start_as_current_span(func.__name__, context=parent_context) as span:
            trace_id = span.get_span_context().trace_id
            span_id = span.get_span_context().span_id
            structured_logger.info(f"Started span: {trace_id}, Span ID: {span_id}")

            result = func(*args, **kwargs)

            structured_logger.info(f"Ending span: {trace_id}, Span ID: {span_id}")
            return result
    return wrapper
