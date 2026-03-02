"""
OpenTelemetry tracer initialization.

This module requires the ``[tracing]`` extra to be installed::

    pip install pylogtracer[tracing]
"""

import logging
import os

logger = logging.getLogger(__name__)


def _get_secret_from_param_store(param_name_env: str) -> str | None:
    """Fetch a secret from AWS Systems Manager Parameter Store.

    Args:
        param_name_env: Name of the env var that holds the SSM parameter path.

    Returns:
        The decrypted parameter value, or ``None`` on failure.
    """
    param_name = os.getenv(param_name_env)
    if not param_name:
        return None

    try:
        import boto3

        ssm = boto3.client("ssm")
        response = ssm.get_parameter(Name=param_name, WithDecryption=True)
        return response["Parameter"]["Value"]
    except Exception:
        logger.exception("Failed to fetch parameter %s", param_name)
        return None


def _create_exporter():
    """Create an OTLP or console span exporter based on environment variables."""
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
    from opentelemetry.sdk.trace.export import ConsoleSpanExporter

    if endpoint := os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT"):
        headers: dict[str, str] = {}
        if headers_param := _get_secret_from_param_store(
            "OTEL_EXPORTER_OTLP_HEADERS_NAME"
        ):
            try:
                headers = dict(
                    pair.split("=", 1)
                    for pair in headers_param.split(",")
                    if "=" in pair
                )
            except Exception:
                logger.exception("Failed to parse OTLP headers")
        try:
            return OTLPSpanExporter(endpoint=endpoint + "/v1/traces", headers=headers)
        except Exception:
            logger.exception("Failed to create OTLP exporter for %s", endpoint)

    if os.getenv("TRACE_OUTPUT_DEBUG"):
        try:
            return ConsoleSpanExporter()
        except Exception:
            logger.exception("Failed to create console exporter")

    return None


def initialize_tracer():
    """Initialize the OpenTelemetry tracer provider and return a tracer."""
    from opentelemetry import trace
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor

    resource = Resource.create(
        {
            "service.name": os.getenv("OTEL_SERVICE_NAME", "pylogtrace-service"),
            "service.namespace": os.getenv("OTEL_SERVICE_NAMESPACE", "default"),
            "deployment.environment": os.getenv(
                "DEPLOYMENT_ENVIRONMENT", "development"
            ),
            "service.version": os.getenv("SERVICE_VERSION", "0.0.0"),
        }
    )

    provider = TracerProvider(resource=resource)
    trace.set_tracer_provider(provider)

    if exporter := _create_exporter():
        provider.add_span_processor(BatchSpanProcessor(exporter))

    return trace.get_tracer(__name__)
