# Database Router SDKs

Official low-level client libraries for the [Xeze Database Router](https://code.xeze.org/xeze/Database-Router).

| SDK | Package | Install |
|-----|---------|---------|
| [Python](./python/) | [`xeze-dbr`](https://pypi.org/project/xeze-dbr/) | `pip install xeze-dbr` |
| [Node.js](./node/) | [`@xeze/dbr`](https://www.npmjs.com/package/@xeze/dbr) | `npm install @xeze/dbr` |
| [Go](./go/) | `code.xeze.org/xeze/Database-Router/sdk/go` | `go get code.xeze.org/xeze/Database-Router/sdk/go` |
| [Rust](./rust/) | [`xeze-dbr`](https://crates.io/crates/xeze-dbr) | `cargo add xeze-dbr` |

## Features

All SDKs provide:
- **mTLS support** — connect with cert files or raw bytes (Vault-friendly)
- **PostgreSQL** — CRUD, raw SQL, schema management
- **MongoDB** — documents, collections, queries
- **Redis** — key/value, TTL, info
- **Health checks** — per-service status

> **Note:** For a higher-level developer experience with automatic Vault auth and namespace isolation, use the [Core wrappers](../core/) instead.
