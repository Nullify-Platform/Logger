import os
from uuid import uuid4
from langfuse import Langfuse
from langfuse.decorators import observe

# Initialize Langfuse client
langfuse = Langfuse(
    secret_key=os.getenv("LANGFUSE_SECRET_KEY"),
    public_key=os.getenv("LANGFUSE_PUBLIC_KEY"),
    host=os.getenv("LANGFUSE_HOST", "https://cloud.langfuse.com")
)

# Generate a unique trace ID
trace_id = str(uuid4())
print("Trace ID:", trace_id)
 
@observe(name="test")
def process_user_request(user_id, request_data, **kwargs):
    # Function logic here
    print("Processing user request:", user_id, request_data, kwargs)
    pass
 
@observe()
def main(**kwargs):
    trace = langfuse.trace(
        name="main_trace",
        id=trace_id  # Use the same trace ID
    )
    process_user_request(
        "user_id",
        "request",
        langfuse_parent_trace_id=trace.id,
        **kwargs
    )
 
 
main(langfuse_parent_trace_id=trace_id)