# xeze-dbr-core (Rust)

High-level, Vault-integrated database client for Rust applications on Xeze infrastructure.

## Installation

```toml
[dependencies]
xeze-dbr-core = { git = "https://code.xeze.org/xeze/Database-Router", subdirectory = "core/rust" }
tokio = { version = "1", features = ["full"] }
```

## Quick Start

```rust
use xeze_dbr_core::XezeCoreClient;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let db = XezeCoreClient::new("xms").await?;
    db.init_workspace().await;

    // Postgres
    let rows = db.pg_query("SELECT * FROM users").await?;

    // Redis
    db.redis_set("cache:student", "Ayush", 300).await?;
    let val = db.redis_get("cache:student").await?;
    println!("{:?}", val); // Some("Ayush")

    Ok(())
}
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VAULT_ADDR` | `http://127.0.0.1:8200` | HashiCorp Vault address |
| `VAULT_TOKEN` | `dev-root-token` | Vault authentication token |
| `DB_ROUTER_HOST` | `db.0.xeze.org:443` | Database Router gRPC host |
