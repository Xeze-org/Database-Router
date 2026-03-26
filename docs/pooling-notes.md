# Connection Pooling Notes

## What is Connection Pooling?

Opening a database connection is **expensive** (takes 20–21 seconds in Docker on first call). Connection pooling solves this by opening connections **once** at startup and reusing them for all requests.

---

## Mental Model: The Taxi Stand

```
WITHOUT Pooling                    WITH Pooling
─────────────────────              ─────────────────────
Request → Open (21s) → Query       App Start → Open 5 connections
Request → Open (21s) → Query       
Request → Open (21s) → Query       Request → Grab free ⚡ → Query → Return
                                   Request → Grab free ⚡ → Query → Return
                                   Request → Grab free ⚡ → Query → Return
```

---

## How to Implement

### 🐘 PostgreSQL (Python)

```bash
pip install psycopg2-binary
```

```python
from psycopg2 import pool

# Create once at startup
pg_pool = pool.SimpleConnectionPool(
    minconn=1,   # Keep at least 1 connection open always
    maxconn=10,  # Allow up to 10 simultaneous connections
    dsn="postgresql://admin:3rHb6NmA5jUc8Tg1@localhost:5432/unified_db"
)

# Use in every request
def get_user(user_id):
    conn = pg_pool.getconn()       # Grab a free connection ⚡
    try:
        cur = conn.cursor()
        cur.execute("SELECT * FROM users WHERE id = %s", (user_id,))
        return cur.fetchone()
    finally:
        pg_pool.putconn(conn)      # Return to pool when done
```

---

### 🐘 PostgreSQL (Node.js)

```bash
npm install pg
```

```javascript
const { Pool } = require("pg");

// Create once at startup
const pool = new Pool({
  host: "localhost",
  port: 5432,
  user: "admin",
  password: "3rHb6NmA5jUc8Tg1",
  database: "unified_db",
  max: 10,          // Max connections in pool
  idleTimeoutMillis: 30000,
});

// Use in every request
async function getUser(userId) {
  const { rows } = await pool.query(
    "SELECT * FROM users WHERE id = $1", [userId]
  );
  return rows[0]; // Pool manages connection automatically ⚡
}
```

---

### 🟥 Redis (Python)

Redis-py has a built-in connection pool enabled by default.

```python
import redis

# Create once at startup — pool is automatic
r = redis.Redis(
    host="localhost",
    port=6379,
    password="p9Kj2mT7vWcD4s8X",
    decode_responses=True,
    max_connections=10   # Pool size
)

# Use anywhere — connections are reused automatically ⚡
r.set("key", "value", ex=60)
r.get("key")
```

---

### 🟥 Redis (Node.js)

```bash
npm install ioredis
```

```javascript
const Redis = require("ioredis");

// ioredis manages a pool automatically
const redis = new Redis({
  host: "localhost",
  port: 6379,
  password: "p9Kj2mT7vWcD4s8X",
});

// Use anywhere ⚡
await redis.set("key", "value", "EX", 60);
await redis.get("key");
```

---

## Pool Sizing Guide

| App Type | Min Connections | Max Connections |
|---|---|---|
| Small personal project | 1 | 5 |
| Medium web app | 2 | 20 |
| High-traffic API | 5 | 50+ |

> **Rule of thumb:** `max_connections = (number of CPU cores × 2) + active disk spindles`

---

## Key Takeaways

| | Without Pooling | With Pooling |
|---|---|---|
| First request | 21,000ms | 21,000ms (startup only) |
| Every request after | 21,000ms | ~1ms ⚡ |
| DB connections used | 1 per request | Shared pool |
| Resource usage | High | Low |

- ✅ Always use a pool in production
- ✅ Create the pool **once** at app startup, not inside request handlers
- ✅ For Redis, pooling is built-in — just reuse the same client instance
