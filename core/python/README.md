# xeze-dbr-core

> 🚀 **Automatic Package Release on changes**

Official unified database wrapper for the Xeze infrastructure. Provides a single, heavily-abstracted client for **PostgreSQL**, **MongoDB**, and **Redis** over mTLS via HashiCorp Vault.

## Installation

```bash
pip install git+https://code.xeze.org/xeze/xeze-dbr-core.git
```

Or install locally for development:

```bash
pip install .
```

## Quick Start

```python
from xeze_core import XezeCoreClient

db = XezeCoreClient(app_namespace="xms")
db.init_workspace()  # Ensures the 'xms_pg' database exists

# Postgres
db.pg_insert("students", {"name": "Ayush", "grade": "A"})

# MongoDB
db.mongo_insert("audit_logs", {"action": "student_added", "timestamp": "2026-04-05"})

# Redis
db.redis_set("cache:student:latest", "Ayush", ttl=300)
```

## Environment Variables

| Variable           | Default                    | Description                |
| ------------------ | -------------------------- | -------------------------- |
| `VAULT_ADDR`       | `http://127.0.0.1:8200`   | HashiCorp Vault address    |
| `VAULT_TOKEN`      | `dev-root-token`           | Vault authentication token |
| `DB_ROUTER_HOST`   | `db.0.xeze.org:443`       | Database Router gRPC host  |

## Architecture

- **Database-per-Service isolation** — each `app_namespace` gets its own `{ns}_pg`, `{ns}_mongo`, and `{ns}:` Redis prefix.
- **Vault mTLS** — client certs are fetched from Vault KV at `dbrouter/certs` and loaded in-memory only.
- **gRPC/Protobuf abstraction** — all Protobuf packing is handled internally; developers work with native Python dicts.
