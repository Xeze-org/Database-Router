# API Reference

Base URL: `http://localhost:8080` (or your domain behind a reverse proxy)

All routes under `/api/v1` require authentication via your reverse proxy (e.g. `X-API-Key` header enforced by Caddy). `/health` is always open.

---

## Health

| Method | Path | Response |
|---|---|---|
| GET | `/health` | `{"status":"healthy","service":"go-db-manager"}` |

---

## Connection Tests

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/test/all` | Test all configured databases |
| GET | `/api/v1/test/postgres` | Test PostgreSQL only |
| GET | `/api/v1/test/mongo` | Test MongoDB only |
| GET | `/api/v1/test/redis` | Test Redis only |

**Response example**
```json
{
  "postgres": {"status": "connected"},
  "mongo":    {"status": "connected"},
  "redis":    {"status": "connected"}
}
```

---

## PostgreSQL

### List databases
```
GET /api/v1/postgres/databases
```

### List tables
```
GET /api/v1/postgres/tables/:database
```

### Select rows
```
GET /api/v1/postgres/select/:database/:table
```
Query params:
| Param | Description |
|---|---|
| `limit` | Max rows to return (default: 100) |
| `offset` | Skip N rows |
| `where` | SQL WHERE clause fragment, e.g. `id=5` |

### Insert a row
```
POST /api/v1/postgres/insert/:database/:table
```
Body:
```json
{"name": "Alice", "email": "alice@example.com"}
```
Returns `201 Created` on success.

### Update a row
```
PUT /api/v1/postgres/update/:database/:table/:id
```
Body: fields to update.

### Delete a row
```
DELETE /api/v1/postgres/delete/:database/:table/:id
```

### Raw SQL query
```
POST /api/v1/postgres/query
```
Body:
```json
{"query": "SELECT count(*) FROM users WHERE active = true"}
```

---

## MongoDB

### List databases
```
GET /api/v1/mongo/databases
```

### List collections
```
GET /api/v1/mongo/collections/:database
```

### Insert a document
```
POST /api/v1/mongo/insert/:database/:collection
```
Body: any JSON object.

### Find documents
```
GET /api/v1/mongo/find/:database/:collection
```
Query params:
| Param | Description |
|---|---|
| `filter` | JSON filter object, e.g. `{"active":true}` |
| `limit` | Max documents (default: 100) |

### Update a document
```
PUT /api/v1/mongo/update/:database/:collection/:id
```
Body: fields to update.

### Delete a document
```
DELETE /api/v1/mongo/delete/:database/:collection/:id
```

---

## Redis

### List keys
```
GET /api/v1/redis/keys?pattern=*
```

### Get a value
```
GET /api/v1/redis/get/:key
```

### Set a value
```
POST /api/v1/redis/set
```
Body:
```json
{"key": "session:abc", "value": "user:42", "ttl": 3600}
```
`ttl` is optional (seconds). Omit for no expiry.

### Delete a key
```
DELETE /api/v1/redis/delete/:key
```

### Server info
```
GET /api/v1/redis/info
```

---

## Error responses

All errors return a JSON object:
```json
{"error": "description of what went wrong"}
```

Common status codes:
| Code | Meaning |
|---|---|
| 200 | OK |
| 201 | Created (insert succeeded) |
| 400 | Bad request (missing/invalid body) |
| 500 | Internal error (DB query failed) |
