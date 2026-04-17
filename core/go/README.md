# xeze-dbr-core (Go)

High-level, Vault-integrated database client for Go applications on Xeze infrastructure.

## Installation

```bash
go get code.xeze.org/xeze/Database-Router/core/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    xeze "code.xeze.org/xeze/Database-Router/core/go"
)

func main() {
    db, err := xeze.New("xms")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx := context.Background()
    db.InitWorkspace(ctx)

    // Postgres
    db.PgInsert(ctx, "students", map[string]interface{}{
        "name": "Ayush", "grade": "A",
    })

    // MongoDB
    db.MongoInsert(ctx, "audit_logs", map[string]interface{}{
        "action": "student_added",
    })

    // Redis
    db.RedisSet(ctx, "cache:student:latest", "Ayush", 300)
    val, _ := db.RedisGet(ctx, "cache:student:latest")
    fmt.Println(val) // "Ayush"
}
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VAULT_ADDR` | `http://127.0.0.1:8200` | HashiCorp Vault address |
| `VAULT_TOKEN` | `dev-root-token` | Vault authentication token |
| `DB_ROUTER_HOST` | `db.0.xeze.org:443` | Database Router gRPC host |
