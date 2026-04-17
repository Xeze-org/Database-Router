# xeze-dbr (Go SDK)

Go gRPC client for the [Xeze Database Router](https://code.xeze.org/xeze/Database-Router).

## Installation

```bash
go get code.xeze.org/xeze/Database-Router/sdk/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    dbr "code.xeze.org/xeze/Database-Router/sdk/go"
    pb "code.xeze.org/xeze/Database-Router/sdk/go/proto/dbrouter"
)

func main() {
    client, err := dbr.Connect(dbr.Options{
        Host:     "db.0.xeze.org:443",
        CertFile: "client.crt",
        KeyFile:  "client.key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Health check
    health, _ := client.Health.Check(ctx, &pb.HealthCheckRequest{})
    fmt.Println("Healthy:", health.OverallHealthy)

    // PostgreSQL
    dbs, _ := client.Postgres.ListDatabases(ctx, &pb.ListDatabasesRequest{})
    fmt.Println("Databases:", dbs.Databases)

    // MongoDB
    mongoDbs, _ := client.Mongo.ListDatabases(ctx, &pb.ListMongoDatabasesRequest{})
    fmt.Println("Mongo DBs:", mongoDbs.Databases)

    // Redis
    info, _ := client.Redis.Info(ctx, &pb.RedisInfoRequest{})
    fmt.Printf("Redis DB size: %d\n", info.DbSize)
}
```

## Connection Options

```go
// File-based mTLS
client, _ := dbr.Connect(dbr.Options{
    Host:     "db.0.xeze.org:443",
    CertFile: "client.crt",
    KeyFile:  "client.key",
    CAFile:   "ca.crt",        // optional
})

// Raw bytes (Vault-friendly)
client, _ := dbr.Connect(dbr.Options{
    Host:     "db.0.xeze.org:443",
    CertData: certPEM,
    KeyData:  keyPEM,
})

// Local development (plaintext)
client, _ := dbr.Connect(dbr.Options{
    Host:     "localhost:50051",
    Insecure: true,
})
```
