import os
import boto3
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import ConsoleSpanExporter
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from .logger import structured_logger
from .tracer import trace_span

__all__ = ['structured_logger', 'trace_span']

def get_secret_from_param_store(param_name_env):
    """Fetch secret from AWS Parameter Store"""
    param_name = os.getenv(param_name_env)
    if not param_name:
        return None
    
    try:
        ssm = boto3.client('ssm')
        response = ssm.get_parameter(
            Name=param_name,
            WithDecryption=True
        )
        return response['Parameter']['Value']
    except Exception as e:
        print(f"Failed to fetch parameter {param_name}: {e}")
        return None

def create_exporter():
    """Create appropriate span exporter based on environment configuration"""
    if endpoint := os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT"):
        # Check for headers in Parameter Store
        headers = {}
        if headers_param := get_secret_from_param_store("OTEL_EXPORTER_OTLP_HEADERS_NAME"):
            try:
                # Parse headers string "key1=value1,key2=value2" format
                headers = dict(
                    pair.split('=', 1) 
                    for pair in headers_param.split(',')
                    if '=' in pair
                )
            except Exception as e:
                print(f"Failed to parse headers: {e}")
        
        try:
            return OTLPSpanExporter(
                endpoint=endpoint,
                headers=headers
            )
        except Exception as e:
            print(f"Failed to create OTLP exporter: {e}")
    
    # Fall back to console exporter if TRACE_OUTPUT_DEBUG is set
    if os.getenv("TRACE_OUTPUT_DEBUG"):
        try:
            return ConsoleSpanExporter()
        except Exception as e:
            print(f"Failed to create console exporter: {e}")
    
    return None

def initialize_tracer():
    """Initialize the OpenTelemetry tracer"""
    # Create resource with service information
    resource = Resource.create({
        "service.name": os.getenv("OTEL_SERVICE_NAME", "pylogtrace-service"),
        "service.namespace": os.getenv("OTEL_SERVICE_NAMESPACE", "default"),
        "deployment.environment": os.getenv("DEPLOYMENT_ENVIRONMENT", "development"),
        "service.version": os.getenv("SERVICE_VERSION", "0.0.0")
    })

    # Set up tracer provider with resource
    provider = TracerProvider(resource=resource)
    trace.set_tracer_provider(provider)

    # # Configure exporter
    # if exporter := create_exporter():
    #     provider.add_span_processor(BatchSpanProcessor(exporter))
    
    return trace.get_tracer(__name__)

# Initialize the tracer
tracer = initialize_tracer()
