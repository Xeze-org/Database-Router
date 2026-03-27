"""
xeze-dbr — Python gRPC client for the Xeze Database Router.

Usage:
    pip install xeze-dbr

    from xeze_dbr import connect

    client = connect(
        host="db.0.xeze.org:443",
        cert_path="client.crt",
        key_path="client.key",
    )

    # Health check
    health = client.health.Check(client.pb2.HealthCheckRequest())
    print(health.overall_healthy)

    # PostgreSQL
    client.postgres.ListDatabases(client.pb2.ListDatabasesRequest())

    # MongoDB
    client.mongo.ListDatabases(client.pb2.ListMongoDatabasesRequest())

    # Redis
    client.redis.Info(client.pb2.RedisInfoRequest())
"""

__version__ = "0.2.1"

import grpc

from . import dbrouter_pb2 as pb2
from . import dbrouter_pb2_grpc as pb2_grpc


class DbRouterClient:
    """A convenience wrapper around the raw gRPC stubs."""

    def __init__(self, channel: grpc.Channel):
        self.channel = channel
        self.pb2 = pb2
        self.health = pb2_grpc.HealthServiceStub(channel)
        self.postgres = pb2_grpc.PostgresServiceStub(channel)
        self.mongo = pb2_grpc.MongoServiceStub(channel)
        self.redis = pb2_grpc.RedisServiceStub(channel)

    def close(self):
        self.channel.close()

    def __enter__(self):
        return self

    def __exit__(self, *args):
        self.close()


def connect(
    host: str,
    cert_path: str | None = None,
    key_path: str | None = None,
    ca_path: str | None = None,
    cert_data: bytes | None = None,
    key_data: bytes | None = None,
    ca_data: bytes | None = None,
    insecure: bool = False,
) -> DbRouterClient:
    """
    Create a DbRouterClient connected to a db-router instance.

    Args:
        host: gRPC target, e.g. "db.0.xeze.org:443" or "localhost:50051".
        cert_path: Path to the client certificate (.crt) for mTLS.
        key_path: Path to the client private key (.key) for mTLS.
        ca_path: Optional path to a custom CA certificate.
        cert_data: Raw bytes of the client certificate (alternative to cert_path).
        key_data: Raw bytes of the client key (alternative to key_path).
        ca_data: Raw bytes of the CA certificate (alternative to ca_path).
        insecure: If True, use a plaintext channel (for local dev only).

    Returns:
        A DbRouterClient with .health, .postgres, .mongo, .redis stubs.
    """
    if insecure:
        channel = grpc.insecure_channel(host)
    else:
        root_ca = ca_data
        client_key = key_data
        client_cert = cert_data

        if ca_path and not root_ca:
            with open(ca_path, "rb") as f:
                root_ca = f.read()
        if key_path and not client_key:
            with open(key_path, "rb") as f:
                client_key = f.read()
        if cert_path and not client_cert:
            with open(cert_path, "rb") as f:
                client_cert = f.read()

        credentials = grpc.ssl_channel_credentials(
            root_certificates=root_ca,
            private_key=client_key,
            certificate_chain=client_cert,
        )
        channel = grpc.secure_channel(host, credentials)

    return DbRouterClient(channel)

