from setuptools import setup, find_packages

setup(
    name="pylogtracer",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "opentelemetry-api",
        "opentelemetry-sdk",
        "opentelemetry-exporter-otlp",
        "loguru"
    ],
    description="Internal logging library with structured logs and tracing",
)