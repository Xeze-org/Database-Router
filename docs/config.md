# Configuration

Config is loaded from `config.json` in the working directory at startup.

**Never commit `config.json`** — it holds your database passwords. Use `config.example.json` as the starting template.

---

## Full example

```json
{
  "postgres": {
    "enabled": "true",
    "host": "localhost",
    "port": 5432,
    "user": "admin",
    "password": "your-password",
    "database": "mydb",
    "sslmode": "disable"
  },
  "mongo": {
    "enabled": "true",
    "uri": "mongodb://user:password@host:27017/dbname?authSource=admin",
    "database": "mydb"
  },
  "redis": {
    "enabled": "true",
    "host": "localhost",
    "port": 6379,
    "password": "your-password",
    "db": 0
  }
}
```

---

## PostgreSQL fields

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | `"true"/"false"` | yes | Skip connection if `"false"` |
| `host` | string | yes | Hostname or IP |
| `port` | number | yes | Usually `5432` |
| `user` | string | yes | Database user |
| `password` | string | yes | Database password |
| `database` | string | yes | Default database name |
| `sslmode` | string | no | `disable`, `require`, or `verify-full` |

## MongoDB fields

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | `"true"/"false"` | yes | Skip connection if `"false"` |
| `uri` | string | yes | Full connection URI |
| `database` | string | yes | Default database name |

## Redis fields

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | `"true"/"false"` | yes | Skip connection if `"false"` |
| `host` | string | yes | Hostname or IP |
| `port` | number | yes | Usually `6379` |
| `password` | string | no | Leave empty string if no auth |
| `db` | number | no | Redis DB index (default `0`) |

## TLS / mTLS fields

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | yes | `false` = plain-text, `true` = enable TLS |
| `cert_file` | string | if enabled | Path to server certificate (PEM) |
| `key_file` | string | if enabled | Path to server private key (PEM) |
| `ca_file` | string | for mTLS | Path to CA certificate used to verify client certs |
| `client_auth` | string | for mTLS | `"require"` enforces mutual TLS; omit for server-only TLS |

```json
"tls": {
  "enabled": true,
  "cert_file": "certs/server.crt",
  "key_file":  "certs/server.key",
  "ca_file":   "certs/ca.crt",
  "client_auth": "require"
}
```

Generate dev certificates:
```powershell
# Windows
.\certs\gen-certs.ps1 -Domain localhost

# Linux / macOS
bash certs/gen-certs.sh localhost
```

---

## Environment variable overrides

Every config field can be overridden with an environment variable. Useful for Kubernetes secrets or Docker env injection without a mounted file.

| Env var | Overrides |
|---|---|
| `POSTGRES_ENABLED` | `postgres.enabled` |
| `POSTGRES_HOST` | `postgres.host` |
| `POSTGRES_PORT` | `postgres.port` |
| `POSTGRES_USER` | `postgres.user` |
| `POSTGRES_PASSWORD` | `postgres.password` |
| `POSTGRES_DB` | `postgres.database` |
| `MONGO_ENABLED` | `mongo.enabled` |
| `MONGO_URI` | `mongo.uri` |
| `MONGO_DATABASE` | `mongo.database` |
| `REDIS_ENABLED` | `redis.enabled` |
| `REDIS_HOST` | `redis.host` |
| `REDIS_PORT` | `redis.port` |
| `REDIS_PASSWORD` | `redis.password` |
| `PORT` | gRPC server port (default `50051`) |
| `WEBUI_PORT` | Web UI HTTP port (default `8080`) |
| `GRPC_ADDR` | Web UI → router address (default `localhost:50051`) |
| `TLS_ENABLED` | `"true"` / `"false"` |
| `TLS_CERT_FILE` | Path to server cert |
| `TLS_KEY_FILE` | Path to server key |
| `TLS_CA_FILE` | Path to CA cert |
| `TLS_CLIENT_AUTH` | `"require"` for mTLS |

If an env var is set it takes priority over the value in `config.json`.
