# mTLS Certificate Guide

This guide explains what certificates you need, how to generate them for a
real domain, and exactly which files each backend app needs to connect.

---

## What is mTLS?

Normal TLS only proves the **server** is who it says it is (like HTTPS).
**Mutual TLS (mTLS)** means both sides prove their identity:

```
Your App ──── "here is my client.crt" ───▶ db-router
Your App ◀─── "here is my server.crt" ─── db-router
         both verified by the same CA
```

If a caller has no valid client certificate — a rogue app, a port scanner,
a misconfigured service — the TLS handshake is **rejected before a single
byte of application data is exchanged**.

---

## Certificate hierarchy

```
ca.key  +  ca.crt          ← Certificate Authority (you own this)
              │
    ┌─────────┴──────────┐
    │                    │
server.crt          client-X.crt     client-Y.crt …
(signed by CA)      (signed by CA)   (one per app)
```

| File | Who holds it | Purpose |
|---|---|---|
| `ca.key` | You only, offline | Signs new certs — **never deploy, never share** |
| `ca.crt` | Server + every client | Verifies the other side's cert was signed by your CA |
| `server.key` | Server only | Server's private key — keep secret |
| `server.crt` | Server only | Server's identity — given to clients so they can verify |
| `client.key` | Each app (secret) | App's private key |
| `client.crt` | Each app | App's identity — sent to server during handshake |

---

## Step 1 — Generate certificates for your domain

### Windows (PowerShell)

```powershell
# From the repo root:
.\certs\gen-certs.ps1 -Domain db.yourdomain.com

# For multiple SANs (e.g. also reachable as an internal hostname):
# Edit gen-certs.ps1 and add DNS.3 = db.internal to the [alt_names] block
```

### Linux / macOS

```bash
DOMAIN=db.yourdomain.com bash certs/gen-certs.sh
```

Both scripts write 6 files into `certs/`:

```
certs/
├── ca.key          ← lock in a safe, never leave this machine
├── ca.crt          ← distribute to every app that needs to verify connections
├── server.key      ← goes on the db-router server
├── server.crt      ← goes on the db-router server
├── client.key      ← goes in each backend app
└── client.crt      ← goes in each backend app
```

---

## Step 2 — What the server needs

In `config.json` on the machine running `db-router`:

```json
"tls": {
  "enabled":     true,
  "cert_file":   "certs/server.crt",
  "key_file":    "certs/server.key",
  "ca_file":     "certs/ca.crt",
  "client_auth": "require"
}
```

Or via environment variables:
```bash
TLS_ENABLED=true
TLS_CERT_FILE=certs/server.crt
TLS_KEY_FILE=certs/server.key
TLS_CA_FILE=certs/ca.crt
TLS_CLIENT_AUTH=require
```

The router will now **reject any connection that does not present a client
certificate signed by your CA**.

---

## Step 3 — What each backend app needs

Every app that connects to `db-router` needs **three files**:

| File | Why |
|---|---|
| `ca.crt` | To verify that the server's cert was signed by your CA |
| `client.crt` | The app's identity (shown to the server) |
| `client.key` | The app's private key (never leave the app's machine) |

> You can generate a **separate** client cert for each app so you can revoke
> them independently (see "Multiple clients" below).

---

## Step 4 — Connecting from each language / tool

### grpcurl (testing)

```bash
grpcurl \
  -cacert certs/ca.crt \
  -cert   certs/client.crt \
  -key    certs/client.key \
  db.yourdomain.com:50051 \
  dbrouter.HealthService/Check
```

---

### Go app (direct gRPC client)

```go
import (
    "google.golang.org/grpc/credentials"
    "crypto/tls"
    "crypto/x509"
    "os"
)

func loadCreds() credentials.TransportCredentials {
    ca, _   := os.ReadFile("certs/ca.crt")
    pool    := x509.NewCertPool()
    pool.AppendCertsFromPEM(ca)

    cert, _ := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")

    return credentials.NewTLS(&tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      pool,
        ServerName:   "db.yourdomain.com",
    })
}

conn, err := grpc.NewClient("db.yourdomain.com:50051",
    grpc.WithTransportCredentials(loadCreds()),
)
```

---

### Python app (`examples/python`)

The Python `DbRouterClient` calls the Caddy HTTPS proxy (`:443`).
To present a client certificate with requests, pass `verify=`:

```python
from client import DbRouterClient

c = DbRouterClient(
    "https://db.yourdomain.com:8080",
    tls_ca="certs/ca.crt",          # verify server cert
    tls_cert=("certs/client.crt", "certs/client.key"),  # present client cert
)
```

To add this support, extend `DbRouterClient.__init__` in `client.py`:

```python
def __init__(self, base_url, tls_ca=None, tls_cert=None, timeout=10):
    self.base    = base_url.rstrip("/")
    self.timeout = timeout
    self.session = requests.Session()
    if tls_ca:
        self.session.verify = tls_ca
    if tls_cert:
        self.session.cert = tls_cert   # (cert_path, key_path)
```

---

### Node.js / other languages

Every gRPC library follows the same three-file pattern:

```js
// Node.js (@grpc/grpc-js)
const grpc = require('@grpc/grpc-js');
const fs   = require('fs');

const creds = grpc.credentials.createSsl(
  fs.readFileSync('certs/ca.crt'),
  fs.readFileSync('certs/client.key'),
  fs.readFileSync('certs/client.crt'),
);
const client = new proto.PostgresService('db.yourdomain.com:50051', creds);
```

---

## Multiple client certificates (one per app)

Generate a new client cert for each backend app. Keep the same CA:

```bash
# Linux
openssl genrsa -out certs/backend-api.key 4096
openssl req -new -key certs/backend-api.key -out certs/backend-api.csr \
  -subj "/CN=backend-api/O=myorg/C=US"
openssl x509 -req -days 3650 \
  -in certs/backend-api.csr \
  -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
  -out certs/backend-api.crt
rm certs/backend-api.csr
```

```powershell
# Windows
openssl genrsa -out certs\backend-api.key 4096
openssl req -new -key certs\backend-api.key -out certs\backend-api.csr `
  -subj "/CN=backend-api/O=myorg/C=US"
openssl x509 -req -days 3650 `
  -in certs\backend-api.csr `
  -CA certs\ca.crt -CAkey certs\ca.key -CAcreateserial `
  -out certs\backend-api.crt
Remove-Item certs\backend-api.csr
```

Give `backend-api.crt` + `backend-api.key` + `ca.crt` to that app only.
If it is compromised you can revoke just that cert without touching others.

---

## Cloudflare setup

You have two independent options with Cloudflare — they can be used together.

### Option A — Cloudflare Authenticated Origin Pulls

Cloudflare proxies HTTPS traffic and presents **its own client certificate**
to your origin. Your server verifies that the request really came from
Cloudflare (not someone who bypassed the proxy).

1. Download Cloudflare's origin CA:
   ```bash
   curl -o certs/cloudflare-origin-ca.pem \
     https://developers.cloudflare.com/ssl/static/authenticated_origin_pull_ca.pem
   ```

2. Set `ca_file` in `config.json` to `certs/cloudflare-origin-ca.pem`

3. In Cloudflare dashboard: **SSL/TLS → Origin Server → Authenticated Origin Pulls → On**

Now your server only accepts connections from Cloudflare edges. Direct
connections (bypassing Cloudflare) are rejected.

---

### Option B — Cloudflare mTLS API Shield (client → Cloudflare)

Enforce that browsers / apps connecting to your Cloudflare-fronted domain
present a valid client certificate. Cloudflare checks the cert at its edge
before forwarding the request.

1. In Cloudflare: **SSL/TLS → Client Certificates → Create Certificate**
   (or upload your own CA's `ca.crt` under "Custom CA")

2. Create an **mTLS Rule** under **Security → WAF → Custom Rules**:
   ```
   (not cf.tls_client_auth.cert_verified) → Block
   ```

3. Distribute `client.crt` + `client.key` to each trusted caller.

---

### Combined (recommended for production)

```
Your App  ──[client.crt]──▶  Cloudflare edge  ──[CF origin cert]──▶  db-router
          mTLS at edge                           Authenticated Origin Pull
```

- Cloudflare verifies the caller has your `client.crt` (Option B)
- Your server verifies the connection came from Cloudflare (Option A)
- Nothing reaches your server without passing both checks

---

## Verify everything works

```bash
# 1. Check the server cert has the right SAN
openssl x509 -in certs/server.crt -noout -text | grep -A3 "Subject Alternative"

# 2. Verify the client cert was signed by your CA
openssl verify -CAfile certs/ca.crt certs/client.crt

# 3. Test mTLS end-to-end
grpcurl \
  -cacert certs/ca.crt \
  -cert   certs/client.crt \
  -key    certs/client.key \
  db.yourdomain.com:50051 \
  dbrouter.HealthService/Check

# 4. Confirm rejection without a cert (should fail)
grpcurl -cacert certs/ca.crt db.yourdomain.com:50051 dbrouter.HealthService/Check
# Expected: "code = Unavailable" or TLS handshake error
```

---

## File permissions checklist

```bash
chmod 600 certs/ca.key certs/server.key certs/client.key   # private keys — owner only
chmod 644 certs/ca.crt certs/server.crt certs/client.crt   # public certs — readable
```

On Windows: right-click → Properties → Security → remove all accounts except
your own service account from the `.key` files.

---

## Certificate expiry

The scripts default to **3650 days (10 years)**. For production consider:

| Cert | Recommended lifetime | Why |
|---|---|---|
| CA | 10 years | Hard to rotate; sign it once |
| Server cert | 1 year | Rotate annually |
| Client cert | 1 year | Shorter = smaller blast radius if leaked |

Check expiry at any time:
```bash
openssl x509 -in certs/server.crt -noout -dates
openssl x509 -in certs/client.crt -noout -dates
```
