from pylogtracer import structured_logger, trace_span

@trace_span
def nested_function():
    structured_logger.info("Inside nested function")

@trace_span
def my_function():
    structured_logger.info("Executing function", user_id="12345", session_id="abcde")
    nested_function()

my_function()
