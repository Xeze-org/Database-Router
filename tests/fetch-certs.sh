#!/usr/bin/env bash
# Downloads mTLS client certificates from the db-router server.
# Usage: ./fetch-certs.sh <user@host> [ssh-key]
#
# Example:
#   ./fetch-certs.sh root@168.144.22.57
#   ./fetch-certs.sh root@168.144.22.57 ~/.ssh/id_ed25519
set -euo pipefail

HOST="${1:?Usage: $0 <user@host> [ssh-key]}"
KEY_OPTS=""
[ -n "${2:-}" ] && KEY_OPTS="-i $2"

DIR="$(cd "$(dirname "$0")" && pwd)/certs"
mkdir -p "$DIR"

echo "Downloading client certificates from $HOST ..."
scp -o StrictHostKeyChecking=no $KEY_OPTS \
  "$HOST:/opt/db-router/certs/ca.crt" \
  "$HOST:/opt/db-router/certs/client.crt" \
  "$HOST:/opt/db-router/certs/client.key" \
  "$DIR/"

chmod 600 "$DIR/client.key"
chmod 644 "$DIR/ca.crt" "$DIR/client.crt"

echo ""
echo "Certificates saved to: $DIR/"
ls -la "$DIR/"
echo ""
echo "Ready to test. Run:"
echo "  ./test-grpc.sh grpc.db.0.xeze.org"
