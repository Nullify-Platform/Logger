from functools import wraps
from opentelemetry import trace
from opentelemetry.propagate import extract
from pylogtracer.logger import structured_logger

# Correct import for `set_span_in_context`
from opentelemetry.trace import set_span_in_context

# Initialize the OpenTelemetry tracer
tracer = trace.get_tracer(__name__)

def trace_span(span_name=None):
    """Decorator to trace function execution with OpenTelemetry, allowing optional span names."""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            # Extract trace context from request headers, default to empty context
            parent_context = extract(kwargs.pop("context", {})) if "context" in kwargs else trace.set_span_in_context(trace.INVALID_SPAN)
            session_id = kwargs.pop("session_id", None)  # Optional session grouping

            # Check if there is an active span
            current_span = trace.get_current_span()
            if current_span is not trace.INVALID_SPAN:
                parent_context = set_span_in_context(current_span)  # Correct parent context

            # Determine the span name (use provided name or default to function name)
            span_label = span_name if span_name else func.__name__

            # Start a new span with the correct context
            with tracer.start_as_current_span(span_label, context=parent_context) as span:
                trace_id = span.get_span_context().trace_id
                span_id = span.get_span_context().span_id

                structured_logger.info(
                    f"Started span {span_label} - Trace ID: {trace_id}, Span ID: {span_id}",
                    session_id=session_id
                )

                result = func(*args, **kwargs)

                structured_logger.info(
                    f"Ending span {span_label} - Trace ID: {trace_id}, Span ID: {span_id}",
                    session_id=session_id
                )
                return result
        return wrapper
    return decorator
