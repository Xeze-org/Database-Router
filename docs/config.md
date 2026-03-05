# Configuration

Config is loaded from `config.json` in the working directory at startup.

**Never commit `config.json`** -- it holds your database passwords. Use `config.example.json` as the starting template.

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

---

## Environment variable overrides

Every config field can be overridden with an environment variable. This is useful when running in Kubernetes or passing secrets via Docker env rather than a mounted file.

| Env var | Overrides |
|---|---|
| `PG_ENABLED` | `postgres.enabled` |
| `PG_HOST` | `postgres.host` |
| `PG_PORT` | `postgres.port` |
| `PG_USER` | `postgres.user` |
| `PG_PASSWORD` | `postgres.password` |
| `PG_DATABASE` | `postgres.database` |
| `PG_SSLMODE` | `postgres.sslmode` |
| `MONGO_ENABLED` | `mongo.enabled` |
| `MONGO_URI` | `mongo.uri` |
| `MONGO_DATABASE` | `mongo.database` |
| `REDIS_ENABLED` | `redis.enabled` |
| `REDIS_HOST` | `redis.host` |
| `REDIS_PORT` | `redis.port` |
| `REDIS_PASSWORD` | `redis.password` |
| `REDIS_DB` | `redis.db` |
| `PORT` | HTTP server port (default `8080`) |

If an env var is set, it takes priority over the value in `config.json`.
