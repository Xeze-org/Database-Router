//! # xeze-dbr
//!
//! Rust gRPC client for the Xeze Database Router.
//!
//! ```rust,no_run
//! use xeze_dbr::{connect, Options};
//!
//! #[tokio::main]
//! async fn main() -> Result<(), Box<dyn std::error::Error>> {
//!     let client = connect(Options {
//!         host: "db.0.xeze.org:443".into(),
//!         cert_path: Some("client.crt".into()),
//!         key_path: Some("client.key".into()),
//!         ..Default::default()
//!     }).await?;
//!
//!     // Health check
//!     let health = client.health.clone().check(pb::HealthCheckRequest {}).await?;
//!     println!("Healthy: {}", health.into_inner().overall_healthy);
//!
//!     Ok(())
//! }
//! ```

pub mod pb {
    tonic::include_proto!("dbrouter");
}

use std::path::PathBuf;
use tonic::transport::{Certificate, Channel, ClientTlsConfig, Identity};

/// Connection options for the Database Router.
#[derive(Debug, Clone, Default)]
pub struct Options {
    /// gRPC target, e.g. "db.0.xeze.org:443"
    pub host: String,

    /// Path to client certificate file (.crt)
    pub cert_path: Option<PathBuf>,
    /// Path to client key file (.key)
    pub key_path: Option<PathBuf>,
    /// Optional path to CA certificate
    pub ca_path: Option<PathBuf>,

    /// Raw PEM bytes of client certificate (for Vault)
    pub cert_data: Option<Vec<u8>>,
    /// Raw PEM bytes of client key (for Vault)
    pub key_data: Option<Vec<u8>>,
    /// Raw PEM bytes of CA certificate
    pub ca_data: Option<Vec<u8>>,

    /// Use plaintext channel (for local dev only)
    pub insecure: bool,
}

/// A connected Database Router client with all four service stubs.
pub struct Client {
    pub health: pb::health_service_client::HealthServiceClient<Channel>,
    pub postgres: pb::postgres_service_client::PostgresServiceClient<Channel>,
    pub mongo: pb::mongo_service_client::MongoServiceClient<Channel>,
    pub redis: pb::redis_service_client::RedisServiceClient<Channel>,
}

/// Connect to a Database Router instance.
pub async fn connect(opts: Options) -> Result<Client, Box<dyn std::error::Error>> {
    let endpoint = Channel::from_shared(format!("https://{}", opts.host))?;

    let channel = if opts.insecure {
        let endpoint = Channel::from_shared(format!("http://{}", opts.host))?;
        endpoint.connect().await?
    } else {
        let mut tls = ClientTlsConfig::new();

        // Load client identity
        let cert_pem = match (&opts.cert_data, &opts.cert_path) {
            (Some(data), _) => data.clone(),
            (_, Some(path)) => tokio::fs::read(path).await?,
            _ => return Err("cert_path or cert_data required".into()),
        };
        let key_pem = match (&opts.key_data, &opts.key_path) {
            (Some(data), _) => data.clone(),
            (_, Some(path)) => tokio::fs::read(path).await?,
            _ => return Err("key_path or key_data required".into()),
        };
        tls = tls.identity(Identity::from_pem(cert_pem, key_pem));

        // Load CA if provided
        let ca_pem = match (&opts.ca_data, &opts.ca_path) {
            (Some(data), _) => Some(data.clone()),
            (_, Some(path)) => Some(tokio::fs::read(path).await?),
            _ => None,
        };
        if let Some(ca) = ca_pem {
            tls = tls.ca_certificate(Certificate::from_pem(ca));
        }

        endpoint.tls_config(tls)?.connect().await?
    };

    Ok(Client {
        health: pb::health_service_client::HealthServiceClient::new(channel.clone()),
        postgres: pb::postgres_service_client::PostgresServiceClient::new(channel.clone()),
        mongo: pb::mongo_service_client::MongoServiceClient::new(channel.clone()),
        redis: pb::redis_service_client::RedisServiceClient::new(channel),
    })
}
