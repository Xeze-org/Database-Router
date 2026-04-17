# xeze-dbr (Rust SDK)

Rust gRPC client for the [Xeze Database Router](https://code.xeze.org/xeze/Database-Router).

[![crates.io](https://img.shields.io/crates/v/xeze-dbr.svg)](https://crates.io/crates/xeze-dbr)

## Installation

```bash
cargo add xeze-dbr
```

## Quick Start

```rust
use xeze_dbr::{connect, Options, pb};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = connect(Options {
        host: "db.0.xeze.org:443".into(),
        cert_path: Some("client.crt".into()),
        key_path: Some("client.key".into()),
        ..Default::default()
    }).await?;

    // Health check
    let health = client.health.clone().check(pb::HealthCheckRequest {}).await?;
    println!("Healthy: {}", health.into_inner().overall_healthy);

    // PostgreSQL
    let dbs = client.postgres.clone()
        .list_databases(pb::ListDatabasesRequest {}).await?;
    println!("Databases: {:?}", dbs.into_inner().databases);

    // MongoDB
    let mongo_dbs = client.mongo.clone()
        .list_databases(pb::ListMongoDatabasesRequest {}).await?;
    println!("Mongo DBs: {:?}", mongo_dbs.into_inner().databases);

    // Redis
    let info = client.redis.clone()
        .info(pb::RedisInfoRequest {}).await?;
    println!("Redis DB size: {}", info.into_inner().db_size);

    Ok(())
}
```

## Connection Options

```rust
// File-based mTLS
let client = connect(Options {
    host: "db.0.xeze.org:443".into(),
    cert_path: Some("client.crt".into()),
    key_path: Some("client.key".into()),
    ca_path: Some("ca.crt".into()),  // optional
    ..Default::default()
}).await?;

// Raw bytes (Vault-friendly)
let client = connect(Options {
    host: "db.0.xeze.org:443".into(),
    cert_data: Some(cert_pem),
    key_data: Some(key_pem),
    ..Default::default()
}).await?;

// Local development (plaintext)
let client = connect(Options {
    host: "localhost:50051".into(),
    insecure: true,
    ..Default::default()
}).await?;
```
