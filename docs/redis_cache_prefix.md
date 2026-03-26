# Why Do We Need a Cache Prefix?

## The Core Problem: One Redis, Many Apps

Redis is just a giant key-value store with **no built-in namespacing**. Every key lives in the same flat space. If two apps use the same key name, they will overwrite each other's data silently — and this is a very dangerous bug.

---

## Real Conflict Example (No Prefix)

```
App 1 (E-Commerce) stores:   user:1  →  { name: "Ayush", cart: [...] }
App 2 (Blog)       stores:   user:1  →  { name: "Ayush", posts: [...] }
```

App 2 just **silently overwrote** App 1's data. Now App 1 reads blog data thinking it's cart data. Your app breaks in unpredictable ways.

---

## The Fix: Cache Prefix

Add the app name as a prefix to every key. Now the keys are unique and never clash.

```
App 1 (E-Commerce):   ecommerce:user:1  →  { name: "Ayush", cart: [...] }
App 2 (Blog):         blog:user:1       →  { name: "Ayush", posts: [...] }
```

They coexist peacefully in the same Redis! ✅

---

## Standard Key Format

```
{CACHE_PREFIX}:{entity}:{id}
     ↑               ↑      ↑
  app name       what it is  which one
```

**Examples:**
```
testdb:user:1
testdb:user:42
testdb:product:99
testdb:session:abc123
```

---

## In Code

```python
CACHE_PREFIX = "testdb"
CACHE_TTL    = 300       # 5 minutes

def make_key(entity, id):
    return f"{CACHE_PREFIX}:{entity}:{id}"

# Usage
key = make_key("user", 1)        # → "testdb:user:1"
redis.set(key, data, ex=CACHE_TTL)
redis.get(key)
redis.delete(key)
```

---

## Summary

| Without Prefix | With Prefix |
|---|---|
| Keys can clash between apps | Every key is unique |
| Silent data corruption | Safe, isolated data |
| Hard to debug | Easy to identify which app owns a key |
| `user:1` | `testdb:user:1` |

> **Rule:** Always set a `CACHE_PREFIX` in your app config. It costs nothing and saves you from very painful bugs.
