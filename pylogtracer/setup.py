from setuptools import setup, find_packages

setup(
    name="pylogtracer",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "loguru",
        "opentelemetry-sdk",
        "opentelemetry-api",
        "opentelemetry-exporter-otlp-proto-http",
        "opentelemetry-propagator-tracecontext",
    ],
    author="platform@nullify.ai",
    description="Internal logging library with structured logs and tracing",
)