from setuptools import find_packages, setup

setup(
    name="pylogtracer",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "opentelemetry-api",
        "opentelemetry-sdk",
        "opentelemetry-exporter-otlp",
        "loguru",
    ],
    description="Internal logging library with structured logs and tracing",
    extras_require={
        "dev": [
            "pytest>=8.3.4",
            "pytest-asyncio>=0.25.0",
            "pip-tools>=7.4.1",
            "ruff>=0.8.3"
        ],
    },
)
