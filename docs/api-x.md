# DB Manager API — Integration Guide

A unified HTTP API for PostgreSQL, MongoDB, and Redis — exposed through Caddy with TLS and API key auth.  
No direct database connections needed. Works from any language.

---

## Base URLs

| Service    | Base URL                     |
|------------|------------------------------|
| PostgreSQL | `https://pg.0.xeze.org`      |
| MongoDB    | `https://mongo.0.xeze.org`   |
| Redis      | `https://redis.0.xeze.org`   |

---

## Authentication

Every request must include the API key header:

```
X-API-Key: X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx
```

Requests missing this header receive `401 Unauthorized`.

---

## Error Responses

All errors return JSON:

```json
{ "error": "description of what went wrong" }
```

| Status | Meaning                          |
|--------|----------------------------------|
| 400    | Bad request / missing fields     |
| 401    | Missing or wrong API key         |
| 404    | Key / record not found           |
| 500    | Internal server / query error    |
| 503    | Database not enabled / down      |

---

## Health Check

```
GET https://pg.0.xeze.org/health
GET https://mongo.0.xeze.org/health
GET https://redis.0.xeze.org/health
```

No API key required. Returns `200 { "status": "healthy", "service": "go-db-manager" }`.

---

## PostgreSQL API (`pg.0.xeze.org`)

### Test connection
```
GET /test
```
```json
{ "status": "connected", "database": "unified_db", "host": "localhost" }
```

---

### List databases
```
GET /databases
```
```json
{ "databases": ["unified_db", "postgres"] }
```

---

### List tables
```
GET /tables/{database}
```
```json
{ "database": "unified_db", "tables": ["users", "orders"] }
```

---

### Execute raw SQL
```
POST /query
Content-Type: application/json

{ "query": "SELECT * FROM users WHERE active = true" }
```
```json
{
  "columns": ["id", "name", "email", "created_at"],
  "rows": [
    { "id": 1, "name": "Alice", "email": "alice@example.com", "created_at": "2026-03-01T10:00:00Z" }
  ],
  "count": 1
}
```
> Use this for any SELECT, CREATE TABLE, ALTER TABLE, etc.  
> For INSERT/UPDATE/DELETE use the dedicated endpoints below (they're safer).

---

### Select rows from table
```
GET /select/{database}/{table}?limit=50
```
```json
{
  "database": "unified_db",
  "table": "users",
  "data": [ { "id": 1, "name": "Alice" } ],
  "count": 1
}
```
Query param `limit` defaults to `100`.

---

### Insert row
```
POST /insert/{database}/{table}
Content-Type: application/json

{ "name": "Bob", "email": "bob@example.com" }
```
```json
{ "database": "unified_db", "table": "users", "inserted_id": 5 }
```
Keys in the JSON body map directly to column names. The table must have an `id` column with `RETURNING id`.

---

### Update row by ID
```
PUT /update/{database}/{table}/{id}
Content-Type: application/json

{ "name": "Bob Updated", "email": "new@example.com" }
```
```json
{ "database": "unified_db", "table": "users", "id": "5", "rows_affected": 1 }
```
Only the fields you send are updated. Uses `WHERE id = $n`.

---

### Delete row by ID
```
DELETE /delete/{database}/{table}/{id}
```
```json
{ "database": "unified_db", "table": "users", "id": "5", "rows_affected": 1 }
```

---

## MongoDB API (`mongo.0.xeze.org`)

### Test connection
```
GET /test
```
```json
{ "status": "connected", "database": "unified_db" }
```

---

### List databases
```
GET /databases
```
```json
{ "databases": ["unified_db", "admin", "local"] }
```

---

### List collections
```
GET /collections/{database}
```
```json
{ "database": "unified_db", "collections": ["users", "logs"] }
```

---

### Insert document
```
POST /insert/{database}/{collection}
Content-Type: application/json

{ "name": "Alice", "email": "alice@example.com", "age": 30 }
```
```json
{ "database": "unified_db", "collection": "users", "inserted_id": "65f1a2b3c4d5e6f7a8b9c0d1" }
```

---

### Find all documents
```
GET /find/{database}/{collection}
```
```json
{
  "database": "unified_db",
  "collection": "users",
  "documents": [ { "_id": "65f1...", "name": "Alice" } ],
  "count": 1
}
```

---

### Update document by ID
```
PUT /update/{database}/{collection}/{id}
Content-Type: application/json

{ "age": 31 }
```
`{id}` must be a valid MongoDB ObjectID hex string.
```json
{ "database": "unified_db", "collection": "users", "matched_count": 1, "modified_count": 1 }
```

---

### Delete document by ID
```
DELETE /delete/{database}/{collection}/{id}
```
```json
{ "database": "unified_db", "collection": "users", "deleted_count": 1 }
```

---

## Redis API (`redis.0.xeze.org`)

### Test connection
```
GET /test
```
```json
{ "status": "connected", "host": "127.0.0.1", "port": 6379 }
```

---

### List keys
```
GET /keys
GET /keys?pattern=session:*
```
```json
{ "keys": ["session:abc", "cache:users"], "count": 2 }
```
`pattern` defaults to `*` (all keys).

---

### Set a key
```
POST /set
Content-Type: application/json

{ "key": "session:abc", "value": "user123", "ttl": 3600 }
```
```json
{ "key": "session:abc", "value": "user123", "ttl": 3600 }
```
`ttl` is in seconds. Set `0` or omit for no expiry.  
`value` must be a string — JSON-encode objects before storing.

---

### Get a key
```
GET /get/{key}
```
```json
{ "key": "session:abc", "value": "user123", "ttl": 3580 }
```
Returns `404` if key does not exist.

---

### Delete a key
```
DELETE /delete/{key}
```
```json
{ "key": "session:abc", "deleted": true }
```

---

### Redis server info
```
GET /info
```
Returns raw Redis `INFO` output as a string in `{ "info": "..." }`.

---

## Code Examples

### Python (requests)

```python
import requests, json

API_KEY = "X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx"
PG      = "https://pg.0.xeze.org"
MONGO   = "https://mongo.0.xeze.org"
REDIS   = "https://redis.0.xeze.org"
HDR     = {"X-API-Key": API_KEY, "Content-Type": "application/json"}

# --- PostgreSQL ---
# Raw query
rows = requests.post(f"{PG}/query", headers=HDR,
    json={"query": "SELECT * FROM users"}).json()["rows"]

# Insert
new = requests.post(f"{PG}/insert/unified_db/users", headers=HDR,
    json={"name": "Alice", "email": "alice@example.com"}).json()
print(new["inserted_id"])  # → 1

# Update
requests.put(f"{PG}/update/unified_db/users/1", headers=HDR,
    json={"name": "Alice Updated"})

# Delete
requests.delete(f"{PG}/delete/unified_db/users/1", headers=HDR)

# --- Redis (cache helper pattern) ---
def cache_get(key):
    r = requests.get(f"{REDIS}/get/{key}", headers=HDR, timeout=3)
    return r.json().get("value") if r.status_code == 200 else None

def cache_set(key, value, ttl=300):
    requests.post(f"{REDIS}/set", headers=HDR,
        json={"key": key, "value": json.dumps(value), "ttl": ttl})

def cache_del(key):
    requests.delete(f"{REDIS}/delete/{key}", headers=HDR)

# Usage
cached = cache_get("users:list")
if not cached:
    users = requests.post(f"{PG}/query", headers=HDR,
        json={"query": "SELECT * FROM users"}).json()["rows"]
    cache_set("users:list", users, ttl=300)

# --- MongoDB ---
# Insert document
doc = requests.post(f"{MONGO}/insert/unified_db/events", headers=HDR,
    json={"type": "login", "user_id": 42}).json()
print(doc["inserted_id"])  # MongoDB ObjectID hex

# Find all
docs = requests.get(f"{MONGO}/find/unified_db/events", headers=HDR).json()["documents"]
```

---

### JavaScript / Node.js (fetch)

```js
const API_KEY = 'X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx';
const PG    = 'https://pg.0.xeze.org';
const MONGO = 'https://mongo.0.xeze.org';
const REDIS = 'https://redis.0.xeze.org';
const HDR   = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

// PostgreSQL — raw query
const { rows } = await fetch(`${PG}/query`, {
  method: 'POST', headers: HDR,
  body: JSON.stringify({ query: 'SELECT * FROM users' })
}).then(r => r.json());

// Insert
const { inserted_id } = await fetch(`${PG}/insert/unified_db/users`, {
  method: 'POST', headers: HDR,
  body: JSON.stringify({ name: 'Bob', email: 'bob@example.com' })
}).then(r => r.json());

// Redis — set with TTL
await fetch(`${REDIS}/set`, {
  method: 'POST', headers: HDR,
  body: JSON.stringify({ key: 'session:xyz', value: 'user99', ttl: 1800 })
});

// Redis — get
const res = await fetch(`${REDIS}/get/session:xyz`, { headers: HDR });
if (res.ok) {
  const { value } = await res.json();
}

// MongoDB — insert
const { inserted_id: mongoId } = await fetch(`${MONGO}/insert/unified_db/logs`, {
  method: 'POST', headers: HDR,
  body: JSON.stringify({ action: 'page_view', path: '/home' })
}).then(r => r.json());
```

---

### PHP (curl)

```php
<?php
$API_KEY = 'X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx';
$PG = 'https://pg.0.xeze.org';

function db_request(string $method, string $url, array $body = [], string $apiKey = ''): array {
    $ch = curl_init($url);
    curl_setopt_array($ch, [
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_CUSTOMREQUEST  => $method,
        CURLOPT_HTTPHEADER     => [
            "X-API-Key: $apiKey",
            'Content-Type: application/json',
        ],
        CURLOPT_POSTFIELDS     => $body ? json_encode($body) : null,
    ]);
    $response = curl_exec($ch);
    curl_close($ch);
    return json_decode($response, true);
}

// List tables
$tables = db_request('GET', "$PG/tables/unified_db", [], $API_KEY);

// Insert row
$result = db_request('POST', "$PG/insert/unified_db/users",
    ['name' => 'Carlos', 'email' => 'carlos@example.com'], $API_KEY);
echo $result['inserted_id'];
```

---

### Go (net/http)

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

const (
    apiKey = "X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx"
    pgBase = "https://pg.0.xeze.org"
)

func dbRequest(method, url string, body any) (*http.Response, error) {
    var buf bytes.Buffer
    if body != nil {
        json.NewEncoder(&buf).Encode(body)
    }
    req, _ := http.NewRequest(method, url, &buf)
    req.Header.Set("X-API-Key", apiKey)
    req.Header.Set("Content-Type", "application/json")
    return http.DefaultClient.Do(req)
}

func main() {
    // Insert user
    resp, _ := dbRequest("POST", pgBase+"/insert/unified_db/users",
        map[string]any{"name": "Diana", "email": "diana@example.com"})
    defer resp.Body.Close()

    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Println("Inserted ID:", result["inserted_id"])
}
```

---

## Redis — Caching Pattern

Store JSON objects in Redis by serializing them to strings:

```python
import json, requests

HDR   = {"X-API-Key": "X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx"}
REDIS = "https://redis.0.xeze.org"

def cache_get(key):
    r = requests.get(f"{REDIS}/get/{key}", headers=HDR, timeout=3)
    if r.status_code == 200:
        try: return json.loads(r.json()["value"])
        except: return r.json()["value"]
    return None

def cache_set(key, value, ttl=300):
    payload = json.dumps(value) if not isinstance(value, str) else value
    requests.post(f"{REDIS}/set", headers=HDR,
        json={"key": key, "value": payload, "ttl": ttl}, timeout=3)

def cache_delete_pattern(pattern):
    """Delete all keys matching a pattern (e.g. 'myapp:*')"""
    r = requests.get(f"{REDIS}/keys?pattern={pattern}", headers=HDR, timeout=3)
    for key in r.json().get("keys", []):
        requests.delete(f"{REDIS}/delete/{key}", headers=HDR, timeout=3)
```

**Key naming convention:**  
Use `{app}:{resource}:{id}` format, e.g. `myapp:users:list`, `myapp:user:42`.  
This lets you bulk-invalidate with `cache_delete_pattern("myapp:*")`.

---

## Adding This to a New App — Checklist

1. Copy these constants into your app config:
   ```
   PG_API    = https://pg.0.xeze.org
   MONGO_API = https://mongo.0.xeze.org
   REDIS_API = https://redis.0.xeze.org
   API_KEY   = X9kR2mT7vWcD4s8pLqBn6hJe3yFu5gAx
   DB_NAME   = unified_db
   ```

2. Add the `X-API-Key` header to **every** request — no exceptions.

3. Use `POST /query` for any complex SQL (JOINs, aggregations, migrations).

4. Use `/insert`, `/update`, `/delete` endpoints for simple CRUD — they use parameterized queries internally.

5. **Never store integers as Redis values directly** — the API expects strings. Use `str(value)` or `json.dumps(value)`.

6. A `404` from Redis GET means the key does not exist (cache miss), not an error.

7. Check connection health before your app starts:
   ```python
   r = requests.get("https://pg.0.xeze.org/test", headers=HDR, timeout=5)
   assert r.json()["status"] == "connected"
   ```

---

## All Endpoints — Quick Reference

### PostgreSQL (`pg.0.xeze.org`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Service health (no auth) |
| GET | `/test` | Test PG connection |
| GET | `/databases` | List all databases |
| GET | `/tables/{db}` | List tables in database |
| POST | `/query` | Execute raw SQL |
| GET | `/select/{db}/{table}?limit=N` | Select rows |
| POST | `/insert/{db}/{table}` | Insert row (JSON body = columns) |
| PUT | `/update/{db}/{table}/{id}` | Update row by id |
| DELETE | `/delete/{db}/{table}/{id}` | Delete row by id |

### MongoDB (`mongo.0.xeze.org`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Service health (no auth) |
| GET | `/test` | Test Mongo connection |
| GET | `/databases` | List databases |
| GET | `/collections/{db}` | List collections |
| POST | `/insert/{db}/{coll}` | Insert document |
| GET | `/find/{db}/{coll}` | Find all documents |
| PUT | `/update/{db}/{coll}/{id}` | Update doc by ObjectID |
| DELETE | `/delete/{db}/{coll}/{id}` | Delete doc by ObjectID |

### Redis (`redis.0.xeze.org`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Service health (no auth) |
| GET | `/test` | Test Redis connection |
| GET | `/keys?pattern=*` | List keys by pattern |
| POST | `/set` | Set key + value + TTL |
| GET | `/get/{key}` | Get value by key |
| DELETE | `/delete/{key}` | Delete key |
| GET | `/info` | Redis server info |
