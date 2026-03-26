# Deployment

---

## Option 1 — Docker (recommended)

The published image contains zero credentials. You always supply `config.json` at runtime.

```bash
docker run -d \
  -p 50051:50051 \
  -v $(pwd)/config.json:/app/config.json:ro \
  --restart unless-stopped \
  --name db-router \
  ghcr.io/youruser/db-router:latest
```

Verify it started:
```bash
docker logs db-router
grpcurl -plaintext localhost:50051 dbrouter.HealthService/Check
```

---

## Option 2 — Docker Compose (router only)

Use [`deploy/docker-compose.yml`](../deploy/docker-compose.yml) if your databases already run elsewhere:

```bash
cp config.example.json config.json
# fill in your credentials
docker compose -f deploy/docker-compose.yml up -d
```

The router is reachable at `localhost:50051`.

---

## Option 3 — Docker Compose (full stack)

Use [`docker-compose.yml`](../docker-compose.yml) at the repo root to spin up PostgreSQL, MongoDB, Redis, and db-router together:

```bash
cp config.example.json config.json
docker compose up -d
```

The databases are available on their standard ports (`5432`, `27017`, `6379`) and db-router on `50051`.

---

## Option 4 — Build from source

**Requirements:** Go 1.24+

```bash
git clone https://github.com/youruser/db-router
cd db-router
cp config.example.json config.json

# build
go build -o db-router ./cmd/

# run
./db-router
```

Windows:
```bat
start.bat
```

Default port: `50051`. Override with:
```bash
PORT=9000 ./db-router
```

---

## Verifying the server

Once running, use [grpcurl](https://github.com/fullstorydev/grpcurl) to confirm:

```bash
# List all services (server reflection)
grpcurl -plaintext localhost:50051 list

# Health check
grpcurl -plaintext localhost:50051 dbrouter.HealthService/Check

# List PostgreSQL databases
grpcurl -plaintext localhost:50051 dbrouter.PostgresService/ListDatabases
```

Or open an interactive browser with [grpcui](https://github.com/fullstorydev/grpcui):
```bash
grpcui -plaintext localhost:50051
```

---

## Protecting the gRPC port

`database-router` has no built-in authentication. Use one of the following approaches:

### Option A — Bind to localhost only

In `docker-compose.yml` bind to `127.0.0.1`:
```yaml
ports:
  - "127.0.0.1:50051:50051"
```

Your app then calls `localhost:50051` directly. Nothing external can reach it.

### Option B — mTLS with Envoy

Put [Envoy Proxy](https://www.envoyproxy.io/) in front with mutual TLS so only clients holding a trusted certificate can connect:

```yaml
# envoy.yaml (sketch)
static_resources:
  listeners:
  - address: { socket_address: { address: 0.0.0.0, port_value: 443 } }
    filter_chains:
    - transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
          require_client_certificate: true
          common_tls_context:
            tls_certificates: [ { certificate_chain: { filename: /certs/server.crt }, private_key: { filename: /certs/server.key } } ]
            validation_context: { trusted_ca: { filename: /certs/ca.crt } }
      filters:
      - name: envoy.filters.network.http_connection_manager
        # … proxy to localhost:50051
```

### Option C — Caddy with gRPC routing

Caddy 2.7+ supports gRPC reverse proxy natively:

```caddy
db.yourdomain.com {
    reverse_proxy h2c://localhost:50051
}
```

Add an auth plugin (`caddy-auth-portal`, `caddy-security`) or a token-checking `route` block to enforce an API key.

---

## Publishing a new Docker image

The GitHub Actions workflow is **manual only**. To publish a new image:

1. Push your changes to `main`
2. Go to **Actions** in your GitHub repo
3. Click **Build & Publish Docker Image**
4. Click **Run workflow**

The image will be pushed to GHCR as:
- `ghcr.io/youruser/db-router:latest`
- `ghcr.io/youruser/db-router:<short-sha>`

---

## Regenerating protobuf bindings

Only needed if you modify `proto/dbrouter.proto`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --proto_path=proto \
  --go_out=proto/dbrouter --go_opt=module=db-router/proto/dbrouter \
  --go-grpc_out=proto/dbrouter --go-grpc_opt=module=db-router/proto/dbrouter \
  proto/dbrouter.proto
```
