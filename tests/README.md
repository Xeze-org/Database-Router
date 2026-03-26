# Testing db-router

Quick CLI tools to download mTLS certificates and test the gRPC endpoint.

## Prerequisites

- **grpcurl**: [Install](https://github.com/fullstorydev/grpcurl#installation)
  ```bash
  # Go
  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
  # macOS
  brew install grpcurl
  # Windows (scoop)
  scoop install grpcurl
  ```
- **SSH access** to the db-router server (for downloading certificates)

## 1. Download Client Certificates

The mTLS client certificates are generated on the server during deployment.
Download them to your local machine:

```bash
# Linux/macOS
./fetch-certs.sh root@<droplet-ip>
./fetch-certs.sh root@<droplet-ip> path/to/ssh-key

# Windows PowerShell
.\fetch-certs.ps1 -RemoteHost root@<droplet-ip>
.\fetch-certs.ps1 -RemoteHost root@<droplet-ip> -Key path\to\ssh-key
```

Certificates are saved to `tests/certs/` (gitignored).

## 2. Test gRPC Endpoint

```bash
# Linux/macOS — with mTLS
./test-grpc.sh grpc.db.0.xeze.org

# Windows PowerShell — with mTLS
.\test-grpc.ps1 -GrpcHost grpc.db.0.xeze.org

# Without mTLS (if mTLS is disabled on server)
./test-grpc.sh grpc.db.0.xeze.org --no-tls
.\test-grpc.ps1 -GrpcHost grpc.db.0.xeze.org -NoTLS
```

## 3. Manual grpcurl Commands

```bash
CERTS="-cacert certs/ca.crt -cert certs/client.crt -key certs/client.key"

# List all gRPC services
grpcurl $CERTS grpc.db.0.xeze.org:443 list

# Health check
grpcurl $CERTS grpc.db.0.xeze.org:443 dbrouter.HealthService/CheckAll

# List PostgreSQL databases
grpcurl $CERTS grpc.db.0.xeze.org:443 dbrouter.PostgresService/ListDatabases

# Execute a SQL query
grpcurl $CERTS -d '{"database":"unified_db","query":"SELECT version()"}' \
  grpc.db.0.xeze.org:443 dbrouter.PostgresService/ExecuteQuery

# Set a Redis key
grpcurl $CERTS -d '{"key":"hello","value":"world"}' \
  grpc.db.0.xeze.org:443 dbrouter.RedisService/Set

# Get a Redis key
grpcurl $CERTS -d '{"key":"hello"}' \
  grpc.db.0.xeze.org:443 dbrouter.RedisService/Get

# List MongoDB databases
grpcurl $CERTS grpc.db.0.xeze.org:443 dbrouter.MongoService/ListDatabases
```

## WebUI

The web dashboard is available at `https://db.0.xeze.org` (no client certs needed).
