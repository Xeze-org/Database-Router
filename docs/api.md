# gRPC API Reference

The router exposes four gRPC services on port `50051` (default). All service and message definitions live in [`proto/dbrouter.proto`](../proto/dbrouter.proto).

**Server reflection is always enabled** — you do not need the `.proto` file locally to call the API.

```bash
# discover services
grpcurl -plaintext localhost:50051 list

# discover RPCs inside a service
grpcurl -plaintext localhost:50051 list dbrouter.PostgresService

# inspect a message type
grpcurl -plaintext localhost:50051 describe dbrouter.InsertDataRequest
```

---

## HealthService

### `Check`
Returns the connection status of all three backends in one call.

```bash
grpcurl -plaintext localhost:50051 dbrouter.HealthService/Check
```

```json
{
  "overallHealthy": true,
  "postgres": { "status": "connected", "enabled": true, "host": "localhost", "database": "mydb" },
  "mongo":    { "status": "connected", "enabled": true, "database": "mydb" },
  "redis":    { "status": "connected", "enabled": true, "host": "localhost", "port": "6379" }
}
```

`ConnectionStatus` fields:

| Field | Description |
|-------|-------------|
| `status` | `"connected"`, `"disconnected"`, or `"disabled"` |
| `enabled` | `true` if the backend is configured, `false` if disabled in config |
| `host` | hostname (Postgres / Redis only) |
| `port` | port string (Redis only) |
| `database` | default database name (Postgres / Mongo) |
| `error` | error message if `status == "disconnected"` |

### `CheckPostgres` / `CheckMongo` / `CheckRedis`
Same as `Check` but returns a single `ConnectionStatus` for one backend.

```bash
grpcurl -plaintext localhost:50051 dbrouter.HealthService/CheckPostgres
```

---

## PostgresService

All RPCs that target a non-default database open a temporary connection for the duration of the call and close it afterwards.

### `ListDatabases`

Lists all non-template databases.

```bash
grpcurl -plaintext localhost:50051 dbrouter.PostgresService/ListDatabases
```

```json
{ "databases": ["postgres", "mydb", "analytics"] }
```

---

### `CreateDatabase`

Creates a new PostgreSQL database. The name must match `^[a-zA-Z_][a-zA-Z0-9_]*$`.

```bash
grpcurl -plaintext \
  -d '{"name":"analytics"}' \
  localhost:50051 dbrouter.PostgresService/CreateDatabase
```

```json
{ "name": "analytics", "message": "Database created successfully" }
```

> Runs outside any transaction (`CREATE DATABASE` cannot run inside one in PostgreSQL).

---

### `ListTables`

Lists all user tables in the `public` schema of the given database.

```bash
grpcurl -plaintext -d '{"database":"mydb"}' \
  localhost:50051 dbrouter.PostgresService/ListTables
```

```json
{ "database": "mydb", "tables": ["users", "orders", "products"] }
```

---

### `ExecuteQuery`

Executes arbitrary SQL. SELECT-like statements return `columns` + `rows`; DML/DDL returns `rows_affected`.

```bash
# SELECT
grpcurl -plaintext \
  -d '{"query":"SELECT id, name FROM users LIMIT 5","database":"mydb"}' \
  localhost:50051 dbrouter.PostgresService/ExecuteQuery
```

```json
{
  "columns": ["id", "name"],
  "rows": [
    {"fields": {"id": {"numberValue": 1}, "name": {"stringValue": "Alice"}}},
    {"fields": {"id": {"numberValue": 2}, "name": {"stringValue": "Bob"}}}
  ],
  "count": "2"
}
```

```bash
# DML
grpcurl -plaintext \
  -d '{"query":"UPDATE users SET active = true WHERE id = 1"}' \
  localhost:50051 dbrouter.PostgresService/ExecuteQuery
```

```json
{ "rowsAffected": "1", "message": "Command executed successfully" }
```

Request fields:

| Field | Required | Description |
|-------|----------|-------------|
| `query` | yes | SQL statement to execute |
| `database` | no | Target database; defaults to config `database` |

> **Warning:** This RPC executes arbitrary SQL. Never pass untrusted user input directly.

---

### `SelectData`

Runs `SELECT * FROM <table> LIMIT <limit>` with identifier validation and quoting.

```bash
grpcurl -plaintext \
  -d '{"database":"mydb","table":"users","limit":10}' \
  localhost:50051 dbrouter.PostgresService/SelectData
```

```json
{
  "database": "mydb",
  "table": "users",
  "data": [ {"fields": {"id": {"numberValue": 1}, "name": {"stringValue": "Alice"}}} ],
  "count": "1"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `database` | yes | Database name |
| `table` | yes | Table name (alphanumeric + underscore only) |
| `limit` | no | Max rows (default 100, max 10 000) |

---

### `InsertData`

Inserts a row. If the table has an `id` column the new value is returned.

```bash
grpcurl -plaintext \
  -d '{
    "database": "mydb",
    "table": "users",
    "data": {
      "name":  {"stringValue": "Alice"},
      "email": {"stringValue": "alice@example.com"}
    }
  }' \
  localhost:50051 dbrouter.PostgresService/InsertData
```

```json
{ "database": "mydb", "table": "users", "insertedId": "42" }
```

The `data` map uses `google.protobuf.Value` — use `stringValue`, `numberValue`, `boolValue`, or `nullValue` keys as appropriate.

---

### `UpdateData`

Updates a row by `id` (`WHERE id = $n`).

```bash
grpcurl -plaintext \
  -d '{
    "database": "mydb",
    "table": "users",
    "id": "42",
    "data": { "name": {"stringValue": "Alicia"} }
  }' \
  localhost:50051 dbrouter.PostgresService/UpdateData
```

```json
{ "database": "mydb", "table": "users", "id": "42", "rowsAffected": "1" }
```

---

### `DeleteData`

Deletes a row by `id`.

```bash
grpcurl -plaintext \
  -d '{"database":"mydb","table":"users","id":"42"}' \
  localhost:50051 dbrouter.PostgresService/DeleteData
```

```json
{ "database": "mydb", "table": "users", "id": "42", "rowsAffected": "1" }
```

---

## MongoService

### `ListDatabases`

```bash
grpcurl -plaintext localhost:50051 dbrouter.MongoService/ListDatabases
```

```json
{ "databases": ["admin", "mydb", "logs"] }
```

---

### `ListCollections`

```bash
grpcurl -plaintext -d '{"database":"mydb"}' \
  localhost:50051 dbrouter.MongoService/ListCollections
```

```json
{ "database": "mydb", "collections": ["users", "events"] }
```

---

### `InsertDocument`

Body is a `google.protobuf.Struct` (arbitrary JSON object).

```bash
grpcurl -plaintext \
  -d '{
    "database": "mydb",
    "collection": "events",
    "document": { "fields": { "type": {"stringValue": "login"}, "userId": {"numberValue": 1} } }
  }' \
  localhost:50051 dbrouter.MongoService/InsertDocument
```

```json
{ "database": "mydb", "collection": "events", "insertedId": "65f1a2b3c4d5e6f7a8b9c0d1" }
```

---

### `FindDocuments`

Returns all documents in a collection (no filter support yet — use `ExecuteQuery` via Postgres for filtered queries).

```bash
grpcurl -plaintext \
  -d '{"database":"mydb","collection":"events"}' \
  localhost:50051 dbrouter.MongoService/FindDocuments
```

```json
{
  "database": "mydb",
  "collection": "events",
  "documents": [
    { "fields": { "_id": {"stringValue": "65f1…"}, "type": {"stringValue": "login"} } }
  ],
  "count": "1"
}
```

---

### `UpdateDocument`

Updates a document by ObjectID using `$set`.

```bash
grpcurl -plaintext \
  -d '{
    "database": "mydb",
    "collection": "events",
    "id": "65f1a2b3c4d5e6f7a8b9c0d1",
    "update": { "fields": { "type": {"stringValue": "logout"} } }
  }' \
  localhost:50051 dbrouter.MongoService/UpdateDocument
```

```json
{ "database": "mydb", "collection": "events", "matchedCount": "1", "modifiedCount": "1" }
```

---

### `DeleteDocument`

Deletes a document by ObjectID.

```bash
grpcurl -plaintext \
  -d '{"database":"mydb","collection":"events","id":"65f1a2b3c4d5e6f7a8b9c0d1"}' \
  localhost:50051 dbrouter.MongoService/DeleteDocument
```

```json
{ "database": "mydb", "collection": "events", "deletedCount": "1" }
```

---

## RedisService

### `ListKeys`

```bash
grpcurl -plaintext -d '{"pattern":"session:*"}' \
  localhost:50051 dbrouter.RedisService/ListKeys
```

```json
{ "keys": ["session:abc", "session:xyz"], "count": "2" }
```

`pattern` defaults to `*` if omitted.

---

### `SetValue`

```bash
grpcurl -plaintext \
  -d '{"key":"session:abc","value":"user:42","ttl":3600}' \
  localhost:50051 dbrouter.RedisService/SetValue
```

```json
{ "key": "session:abc", "value": "user:42", "ttl": 3600 }
```

`ttl` is in seconds. Set to `0` (or omit) for no expiry.

---

### `GetValue`

```bash
grpcurl -plaintext -d '{"key":"session:abc"}' \
  localhost:50051 dbrouter.RedisService/GetValue
```

```json
{ "key": "session:abc", "value": "user:42", "ttl": 3540 }
```

Returns `codes.NotFound` if the key does not exist.

---

### `DeleteKey`

```bash
grpcurl -plaintext -d '{"key":"session:abc"}' \
  localhost:50051 dbrouter.RedisService/DeleteKey
```

```json
{ "key": "session:abc", "deleted": true }
```

---

### `Info`

Returns raw `INFO` output and the key count for the current database.

```bash
grpcurl -plaintext localhost:50051 dbrouter.RedisService/Info
```

```json
{ "dbSize": "42", "info": "# Server\nredis_version:7.2.4\n…" }
```

---

## Error codes

| gRPC code | Meaning |
|-----------|---------|
| `OK` | Success |
| `INVALID_ARGUMENT` | Bad input (invalid table name, missing field, etc.) |
| `NOT_FOUND` | Key / document / row not found |
| `UNAVAILABLE` | Backend not enabled in config, or connection refused |
| `INTERNAL` | Database driver error |
