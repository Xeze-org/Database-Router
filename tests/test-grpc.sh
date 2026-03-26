#!/usr/bin/env bash
# Quick gRPC connectivity test using grpcurl.
# Requires: grpcurl (https://github.com/fullstorydev/grpcurl)
#
# Usage:
#   ./test-grpc.sh <grpc-host>          # e.g. grpc.db.0.xeze.org
#   ./test-grpc.sh <grpc-host> --no-tls # skip mTLS (plain TLS only)
set -euo pipefail

HOST="${1:?Usage: $0 <grpc-host> [--no-tls]}"
PORT="${GRPC_PORT:-443}"
CERT_DIR="$(cd "$(dirname "$0")" && pwd)/certs"

if ! command -v grpcurl &>/dev/null; then
  echo "ERROR: grpcurl is not installed."
  echo "  Install: https://github.com/fullstorydev/grpcurl#installation"
  echo "  Or:  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
  exit 1
fi

TLS_ARGS=""
if [ "${2:-}" != "--no-tls" ] && [ -f "$CERT_DIR/client.crt" ]; then
  echo "Using mTLS certificates from $CERT_DIR"
  TLS_ARGS="-cacert $CERT_DIR/ca.crt -cert $CERT_DIR/client.crt -key $CERT_DIR/client.key"
elif [ "${2:-}" = "--no-tls" ]; then
  echo "Skipping client certificates (plain TLS)"
else
  echo "WARNING: No client certs found at $CERT_DIR. Run fetch-certs.sh first."
  echo "Trying without client certs..."
fi

echo ""
echo "=== List Services ==="
grpcurl $TLS_ARGS "$HOST:$PORT" list 2>&1 || true

echo ""
echo "=== Health Check ==="
grpcurl $TLS_ARGS "$HOST:$PORT" dbrouter.HealthService/CheckAll 2>&1 || true

echo ""
echo "=== Test PostgreSQL ==="
grpcurl $TLS_ARGS "$HOST:$PORT" dbrouter.HealthService/CheckPostgres 2>&1 || true

echo ""
echo "=== Test MongoDB ==="
grpcurl $TLS_ARGS "$HOST:$PORT" dbrouter.HealthService/CheckMongo 2>&1 || true

echo ""
echo "=== Test Redis ==="
grpcurl $TLS_ARGS "$HOST:$PORT" dbrouter.HealthService/CheckRedis 2>&1 || true

echo ""
echo "=== List Postgres Databases ==="
grpcurl $TLS_ARGS "$HOST:$PORT" dbrouter.PostgresService/ListDatabases 2>&1 || true

echo ""
echo "=== List Redis Keys ==="
grpcurl $TLS_ARGS -d '{"pattern":"*"}' "$HOST:$PORT" dbrouter.RedisService/ListKeys 2>&1 || true

echo ""
echo "Done."
