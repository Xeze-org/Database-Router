//! # xeze-dbr-core
//!
//! High-level, Vault-integrated database client for Rust applications
//! on Xeze infrastructure. Enforces database-per-service isolation
//! via the `app_namespace` parameter.
//!
//! ```rust,no_run
//! use xeze_dbr_core::XezeCoreClient;
//!
//! #[tokio::main]
//! async fn main() -> Result<(), Box<dyn std::error::Error>> {
//!     let db = XezeCoreClient::new("xms").await?;
//!     db.init_workspace().await;
//!
//!     // Postgres
//!     db.pg_query("SELECT * FROM users").await?;
//!
//!     // Redis
//!     db.redis_set("key", "value", 3600).await?;
//!     let val = db.redis_get("key").await?;
//!
//!     Ok(())
//! }
//! ```

use std::collections::HashMap;
use std::env;

use xeze_dbr::pb;

/// Unified client for Postgres, MongoDB, and Redis over mTLS via Vault.
pub struct XezeCoreClient {
    pub app_namespace: String,
    pub pg_db: String,
    pub mongo_db: String,
    pub redis_prefix: String,
    client: xeze_dbr::Client,
}

impl XezeCoreClient {
    /// Create a new client connected via Vault mTLS.
    pub async fn new(app_namespace: &str) -> Result<Self, Box<dyn std::error::Error>> {
        if app_namespace.is_empty() {
            return Err("app_namespace is required".into());
        }

        let vault_addr = env::var("VAULT_ADDR").unwrap_or_else(|_| "http://127.0.0.1:8200".into());
        let vault_token = env::var("VAULT_TOKEN").unwrap_or_else(|_| "dev-root-token".into());
        let host = env::var("DB_ROUTER_HOST").unwrap_or_else(|_| "db.0.xeze.org:443".into());

        // Fetch certs from Vault KV v2
        let http = reqwest::Client::new();
        let url = format!("{}/v1/secret/data/dbrouter/certs", vault_addr);
        let resp: serde_json::Value = http
            .get(&url)
            .header("X-Vault-Token", &vault_token)
            .send()
            .await?
            .json()
            .await?;

        let data = &resp["data"]["data"];
        let cert_pem = data["client_cert"].as_str().ok_or("missing client_cert")?;
        let key_pem = data["client_key"].as_str().ok_or("missing client_key")?;

        let client = xeze_dbr::connect(xeze_dbr::Options {
            host,
            cert_data: Some(cert_pem.as_bytes().to_vec()),
            key_data: Some(key_pem.as_bytes().to_vec()),
            ..Default::default()
        })
        .await?;

        Ok(Self {
            app_namespace: app_namespace.to_string(),
            pg_db: format!("{}_pg", app_namespace),
            mongo_db: format!("{}_mongo", app_namespace),
            redis_prefix: format!("{}:", app_namespace),
            client,
        })
    }

    /// Create the isolated Postgres database. Safe for startup.
    pub async fn init_workspace(&self) {
        let req = pb::CreateDatabaseRequest {
            name: self.pg_db.clone(),
        };
        match self.client.postgres.clone().create_database(req).await {
            Ok(_) => println!("[OK] Provisioned workspace: {}", self.pg_db),
            Err(e) => {
                let msg = e.message().to_lowercase();
                if !msg.contains("already exists") {
                    println!("[WARN] Workspace check failed: {}", e.message());
                }
            }
        }
    }

    // --- PostgreSQL API ---

    /// Execute raw SQL returning results as a Vec of HashMaps.
    pub async fn pg_query(
        &self,
        query: &str,
    ) -> Result<Vec<HashMap<String, String>>, tonic::Status> {
        let req = pb::ExecuteQueryRequest {
            database: self.pg_db.clone(),
            query: query.to_string(),
        };
        let resp = self.client.postgres.clone().execute_query(req).await?;
        let inner = resp.into_inner();

        let mut results = Vec::new();
        for row in inner.rows {
            let mut row_data = HashMap::new();
            for (key, val) in row.fields {
                let s = match val.kind {
                    Some(prost_types::value::Kind::StringValue(s)) => s,
                    Some(prost_types::value::Kind::NumberValue(n)) => n.to_string(),
                    Some(prost_types::value::Kind::BoolValue(b)) => b.to_string(),
                    _ => String::new(),
                };
                row_data.insert(key, s);
            }
            results.push(row_data);
        }
        Ok(results)
    }

    // --- Redis API ---

    /// Set a namespaced key with TTL.
    pub async fn redis_set(
        &self,
        key: &str,
        value: &str,
        ttl: i32,
    ) -> Result<(), tonic::Status> {
        let ns_key = format!("{}{}", self.redis_prefix, key);
        let req = pb::SetValueRequest {
            key: ns_key,
            value: value.to_string(),
            ttl,
        };
        self.client.redis.clone().set_value(req).await?;
        Ok(())
    }

    /// Get a namespaced key. Returns None if not found.
    pub async fn redis_get(&self, key: &str) -> Result<Option<String>, tonic::Status> {
        let ns_key = format!("{}{}", self.redis_prefix, key);
        let req = pb::GetValueRequest { key: ns_key };
        match self.client.redis.clone().get_value(req).await {
            Ok(resp) => Ok(Some(resp.into_inner().value)),
            Err(e) => {
                if e.message().to_lowercase().contains("not found") {
                    Ok(None)
                } else {
                    Err(e)
                }
            }
        }
    }
}
