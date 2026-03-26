# gen-certs.ps1 — Generate a CA + server cert + client cert for mTLS on Windows.
#
# Prerequisites:
#   - OpenSSL must be installed and in PATH.
#     Download from: https://slproweb.com/products/Win32OpenSSL.html
#     Or via winget: winget install ShiningLight.OpenSSL
#
# Usage (run from the repo root):
#   .\certs\gen-certs.ps1 -Domain db.yourdomain.com
#
# Output files (written to certs\):
#   ca.key         CA private key     — KEEP SECRET, never deploy
#   ca.crt         CA certificate     — upload to Cloudflare "Client Certificates" CA
#   server.key     Server private key — deploy on your server
#   server.crt     Server certificate — deploy on your server
#   client.key     Client private key — distribute to trusted callers
#   client.crt     Client certificate — distribute to trusted callers

param(
    [string]$Domain = "localhost",
    [int]$Days = 3650
)

$ErrorActionPreference = "Stop"
$Out = Join-Path $PSScriptRoot "."   # writes into certs\

Write-Host "==> Generating mTLS certificates for domain: $Domain" -ForegroundColor Cyan
Write-Host "    Output directory: $Out"
Write-Host ""

# ── 1. Certificate Authority ──────────────────────────────────────────────────
Write-Host "--> [1/3] Generating CA key and self-signed certificate..." -ForegroundColor Yellow
openssl genrsa -out "$Out\ca.key" 4096
openssl req -new -x509 -days $Days `
    -key "$Out\ca.key" `
    -out "$Out\ca.crt" `
    -subj "/CN=db-router-CA/O=db-router/C=US"

# ── 2. Server certificate ─────────────────────────────────────────────────────
Write-Host "--> [2/3] Generating server key and certificate (SAN: $Domain)..." -ForegroundColor Yellow
openssl genrsa -out "$Out\server.key" 4096
openssl req -new `
    -key "$Out\server.key" `
    -out "$Out\server.csr" `
    -subj "/CN=$Domain/O=db-router/C=US"

# Write SAN extension config
$sanCfg = @"
[req]
req_extensions = v3_req
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = $Domain
DNS.2 = localhost
IP.1  = 127.0.0.1
"@
$sanCfg | Out-File -Encoding ascii "$Out\server-ext.cnf"

openssl x509 -req -days $Days `
    -in "$Out\server.csr" `
    -CA "$Out\ca.crt" `
    -CAkey "$Out\ca.key" `
    -CAcreateserial `
    -extfile "$Out\server-ext.cnf" `
    -extensions v3_req `
    -out "$Out\server.crt"

# ── 3. Client certificate ─────────────────────────────────────────────────────
Write-Host "--> [3/3] Generating client key and certificate..." -ForegroundColor Yellow
openssl genrsa -out "$Out\client.key" 4096
openssl req -new `
    -key "$Out\client.key" `
    -out "$Out\client.csr" `
    -subj "/CN=db-router-client/O=db-router/C=US"

openssl x509 -req -days $Days `
    -in "$Out\client.csr" `
    -CA "$Out\ca.crt" `
    -CAkey "$Out\ca.key" `
    -CAcreateserial `
    -out "$Out\client.crt"

# ── Cleanup temp files ────────────────────────────────────────────────────────
Remove-Item -ErrorAction SilentlyContinue "$Out\server.csr", "$Out\client.csr", "$Out\server-ext.cnf", "$Out\ca.srl"

Write-Host ""
$msg = @"

==> Done! Files written to $Out

    ca.crt      - upload to Cloudflare as your client certificate CA
    server.crt  - deploy on your server (with server.key)
    client.crt  - give to callers that need to connect (with client.key)

==> Verify server cert:
    openssl x509 -in $Out/server.crt -noout -text

==> Test mTLS (after starting db-router with TLS enabled):
    grpcurl -cacert $Out/ca.crt -cert $Out/client.crt -key $Out/client.key ${Domain}:50051 dbrouter.HealthService/Check
"@
Write-Host $msg -ForegroundColor Green
