# db-router Python Example

A minimal Flask web app that connects to the db-router and displays live data
from PostgreSQL, MongoDB, and Redis in a browser dashboard.

## Architecture

```
Browser :5000
    │
    ▼
Flask app.py              ← this example
    │  (HTTP REST calls)
    ▼
db-router webui :8080     ← HTTP proxy panel (cmd/webui)
    │  (gRPC calls)
    ▼
db-router :50051          ← gRPC server (cmd/main.go)
    │
    ├── PostgreSQL
    ├── MongoDB
    └── Redis
```

## Requirements

- Python 3.10+
- `db-router` gRPC server running on `:50051`
- `webui` HTTP proxy running on `:8080`

## Quick start

```bash
# 1. Install dependencies
pip install -r requirements.txt

# 2. (optional) point at a non-default router proxy
export ROUTER_URL=http://localhost:8080

# 3. Run
python app.py
# → http://localhost:5000
```

## Files

| File | Description |
|---|---|
| `client.py` | `DbRouterClient` class — clean Python wrapper around the REST proxy API |
| `app.py` | Flask web app — renders dashboard + proxy endpoints for browser JS |
| `requirements.txt` | Python dependencies |

## Using DbRouterClient standalone

```python
from client import DbRouterClient, DbRouterError

c = DbRouterClient("http://localhost:8080")

# Health
print(c.health())

# PostgreSQL
dbs = c.pg_databases()
rows = c.pg_select("shop", "products", limit=5)
result = c.pg_query("SELECT count(*) FROM products", "shop")
c.pg_insert("shop", "products", {"name": "Trackpad", "price": 59.99, "stock": 30})

# MongoDB
dbs = c.mongo_databases()
docs = c.mongo_find("xeze_test", "mycollection")
inserted = c.mongo_insert("xeze_test", "logs", {"event": "startup", "ok": True})

# Redis
c.redis_set("py:hello", "world", ttl=60)
print(c.redis_get("py:hello"))   # {"key": "py:hello", "value": "world", "ttl": 59}
keys = c.redis_keys("py:*")
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `ROUTER_URL` | `http://localhost:8080` | URL of the webui HTTP proxy |
| `APP_PORT` | `5000` | Port for this Flask app |
