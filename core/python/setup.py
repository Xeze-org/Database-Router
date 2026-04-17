from setuptools import setup, find_packages

setup(
    name="xeze-dbr-core",
    version="0.1.0",
    description="Official unified database wrapper for the Xeze infrastructure.",
    packages=find_packages(),
    install_requires=[
        "hvac>=1.1.0",
        "grpcio>=1.50.0",
        "protobuf>=4.21.0",
        "xeze-dbr"  # Your underlying gRPC router SDK
    ],
    python_requires=">=3.8",
)
