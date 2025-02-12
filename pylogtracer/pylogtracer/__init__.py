import os
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

# Load environment variables for tracing
OTEL_EXPORTER_OTLP_ENDPOINT = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318/v1/traces")
SERVICE_NAME = os.getenv("OTEL_SERVICE_NAME", "pylogtrace-service")

# Set up OpenTelemetry tracing globally
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(SERVICE_NAME)

# Configure exporter to send traces to Grafana Tempo
otlp_exporter = OTLPSpanExporter(endpoint=OTEL_EXPORTER_OTLP_ENDPOINT)
trace.get_tracer_provider().add_span_processor(BatchSpanProcessor(otlp_exporter))

# Import logging and tracing functionalities
from .logger import structured_logger
from .tracer import trace_span
