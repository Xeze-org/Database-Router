# Database Router SDKs

Official client libraries for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router).

| SDK | Package | Install |
|-----|---------|---------|
| [Python](./python/) | [`xeze-dbr`](https://pypi.org/project/xeze-dbr/) | `pip install xeze-dbr` |
| [Node.js](./node/) | `@xeze/dbr` *(coming soon)* | `npm install @xeze/dbr` |

## Features

All SDKs provide:
- 🔐 **mTLS support** — connect with cert files or raw bytes (Vault-friendly)
- 🐘 **PostgreSQL** — CRUD, raw SQL, schema management
- 🍃 **MongoDB** — documents, collections, queries
- 🔴 **Redis** — key/value, TTL, info
- ❤️ **Health checks** — per-service status
