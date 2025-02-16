from pylogtracer import structured_logger
from pylogtracer import trace_span

# Basic logging
structured_logger.info("Application started", environment="production", version="1.0.0")
structured_logger.debug("Debug message", user_id="123")
structured_logger.error("Something went wrong", error_code=500)

@trace_span(span_name="process_order2")
def process_order2(order_id, user_id):
    structured_logger.info("Processing order", order_id=order_id, user_id=user_id)
    # ... your processing logic here ...
    return {"status": "success"}


@trace_span()
def process_order1(order_id, user_id):
    structured_logger.info("Processing order", order_id=order_id, user_id=user_id)
    # ... your processing logic here ...
    return process_order2(order_id, user_id)

# Using the tracer decorator
@trace_span(span_name="process_order")
def process_order(order_id, user_id):
    structured_logger.info("Processing order", order_id=order_id, user_id=user_id)
    # ... your processing logic here ...
    return process_order1(order_id, user_id)

#Call the traced function
result = process_order("ORDER123", "USER456")