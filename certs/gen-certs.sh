#!/usr/bin/env bash
# gen-certs.sh — Generate a CA + server cert + client cert for mTLS.
#
# Usage:
#   chmod +x certs/gen-certs.sh
#   DOMAIN=db.yourdomain.com ./certs/gen-certs.sh
#
# Output files (all written to the certs/ directory):
#   ca.key         CA private key     — KEEP SECRET, never deploy
#   ca.crt         CA certificate     — upload to Cloudflare "Client Certificates" CA
#   server.key     Server private key — deploy on your server
#   server.crt     Server certificate — deploy on your server
#   client.key     Client private key — distribute to trusted callers
#   client.crt     Client certificate — distribute to trusted callers
#
# Cloudflare Authenticated Origin Pulls (optional):
#   Download Cloudflare's origin CA from:
#   https://developers.cloudflare.com/ssl/static/authenticated_origin_pull_ca.pem
#   Save it as certs/cloudflare-origin-pull-ca.pem and set ca_file to that path
#   so your server only accepts connections from Cloudflare.

set -euo pipefail

DOMAIN="${DOMAIN:-localhost}"
DAYS="${DAYS:-3650}"          # 10 years for the CA; adjust for certs as needed
OUT="$(dirname "$0")"         # writes into the certs/ directory

echo "==> Generating mTLS certificates for domain: $DOMAIN"
echo "    Output directory: $OUT"
echo ""

# ── 1. Certificate Authority ──────────────────────────────────────────────────
echo "--> [1/3] Generating CA key and self-signed certificate..."
openssl genrsa -out "$OUT/ca.key" 4096

openssl req -new -x509 -days "$DAYS" \
  -key "$OUT/ca.key" \
  -out "$OUT/ca.crt" \
  -subj "/CN=db-router-CA/O=db-router/C=US"

# ── 2. Server certificate ─────────────────────────────────────────────────────
echo "--> [2/3] Generating server key and certificate (SAN: $DOMAIN)..."
openssl genrsa -out "$OUT/server.key" 4096

openssl req -new \
  -key "$OUT/server.key" \
  -out "$OUT/server.csr" \
  -subj "/CN=$DOMAIN/O=db-router/C=US"

# SAN extension so gRPC hostname verification passes
cat > "$OUT/server-ext.cnf" <<EOF
[req]
req_extensions = v3_req
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = $DOMAIN
DNS.2 = localhost
IP.1  = 127.0.0.1
EOF

openssl x509 -req -days "$DAYS" \
  -in "$OUT/server.csr" \
  -CA "$OUT/ca.crt" \
  -CAkey "$OUT/ca.key" \
  -CAcreateserial \
  -extfile "$OUT/server-ext.cnf" \
  -extensions v3_req \
  -out "$OUT/server.crt"

# ── 3. Client certificate ─────────────────────────────────────────────────────
echo "--> [3/3] Generating client key and certificate..."
openssl genrsa -out "$OUT/client.key" 4096

openssl req -new \
  -key "$OUT/client.key" \
  -out "$OUT/client.csr" \
  -subj "/CN=db-router-client/O=db-router/C=US"

openssl x509 -req -days "$DAYS" \
  -in "$OUT/client.csr" \
  -CA "$OUT/ca.crt" \
  -CAkey "$OUT/ca.key" \
  -CAcreateserial \
  -out "$OUT/client.crt"

# ── Cleanup temp files ────────────────────────────────────────────────────────
rm -f "$OUT/server.csr" "$OUT/client.csr" "$OUT/server-ext.cnf"

# ── Permissions ───────────────────────────────────────────────────────────────
chmod 600 "$OUT/ca.key" "$OUT/server.key" "$OUT/client.key"
chmod 644 "$OUT/ca.crt" "$OUT/server.crt" "$OUT/client.crt"

echo ""
echo "==> Done! Files written to $OUT/"
echo ""
echo "    ca.crt      — upload to Cloudflare as your client certificate CA"
echo "    server.crt  — deploy on your server (with server.key)"
echo "    client.crt  — give to callers that need to connect (with client.key)"
echo ""
echo "==> Verify server cert:"
echo "    openssl x509 -in $OUT/server.crt -noout -text | grep -A2 'Subject Alternative'"
echo ""
echo "==> Test mTLS locally (after starting db-router with TLS enabled):"
echo "    grpcurl -cacert $OUT/ca.crt \\"
echo "            -cert $OUT/client.crt \\"
echo "            -key $OUT/client.key \\"
echo "            $DOMAIN:50051 dbrouter.HealthService/Check"
