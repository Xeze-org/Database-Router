# Core — xeze-dbr-core

High-level, Vault-integrated database client wrappers for the Xeze Database Router infrastructure.

Each implementation provides the same developer experience:
- **One client** for PostgreSQL, MongoDB, and Redis
- **Vault mTLS** — certificates fetched and loaded in-memory automatically
- **Namespace isolation** — `app_namespace` enforces database-per-service

## Implementations

| Language | Directory | Package | Install |
|----------|-----------|---------|---------|
| **Python** | [`python/`](./python/) | [`xeze-dbr-core`](https://pypi.org/project/xeze-dbr-core/) | `pip install xeze-dbr-core` |
| **Node.js** | [`node/`](./node/) | `@xeze/dbr-core` | `npm install @xeze/dbr-core` |
| **Go** | [`go/`](./go/) | `code.xeze.org/xeze/Database-Router/core/go` | `go get code.xeze.org/xeze/Database-Router/core/go` |
| **Rust** | [`rust/`](./rust/) | [`xeze-dbr-core`](https://crates.io/crates/xeze-dbr-core) | `cargo add xeze-dbr-core` |

## SDK vs Core

The **Core** libraries are high-level, opinionated wrappers built *directly on top* of the raw **SDK** libraries.

| Feature | SDK (`sdk/`) | Core (`core/`) |
|---|---|---|
| **Architecture** | Raw generated gRPC stubs | Opinionated wrapper utilizing the SDK internally |
| **mTLS Auth** | Manual (Files, raw bytes, or custom Vault integration) | Automatic (Pre-configured HashiCorp Vault client) |
| **Database Isolation** | Manual (explicitly target any DB or collection) | Vault enforced database-per-service via `app_namespace` |
| **Data Serialization** | Strict Protobuf structures (e.g. `structpb.Value`) | Ergonomic native language types (Dicts, Maps, Objects) |
| **Primary Use Case** | Custom automation, CLI tools, framework builders | Backend application service development |
