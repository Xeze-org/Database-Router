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
| **Rust** | [`rust/`](./rust/) | [`xeze-dbr-core`](https://crates.io/crates/xeze-dbr) | via git dependency |

## SDK vs Core

| | SDK (`sdk/`) | Core (`core/`) |
|---|---|---|
| **Level** | Low-level gRPC stubs | High-level abstraction |
| **Auth** | Manual cert loading | Automatic via Vault |
| **Isolation** | None — raw access | Enforced `app_namespace` |
| **Use case** | Custom tooling, infra scripts | Application development |
