# @xeze/dbr

Node.js gRPC client for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router) — a unified interface for **PostgreSQL**, **MongoDB**, and **Redis** over **mTLS**.

## Install

```bash
npm install @xeze/dbr
```

## Quick Start

```javascript
const { connect } = require("@xeze/dbr");

// Connect with mTLS
const client = connect({
  host: "db.0.xeze.org:443",
  certPath: "client.crt",
  keyPath: "client.key",
});

// Health check
const health = await client.health.Check();
console.log(health.overallHealthy); // true

// PostgreSQL — execute query
const result = await client.postgres.ExecuteQuery({
  database: "mydb",
  query: "SELECT * FROM users LIMIT 10",
});
console.log(result.rows);

// PostgreSQL — insert data
const inserted = await client.postgres.InsertData({
  database: "mydb",
  table: "users",
  data: { name: { stringValue: "Alice" }, age: { numberValue: 28 } },
});
console.log(inserted.insertedId);

// MongoDB — insert document
const doc = await client.mongo.InsertDocument({
  database: "mydb",
  collection: "events",
  document: { fields: { event: { stringValue: "signup" } } },
});
console.log(doc.insertedId);

// Redis — set & get
await client.redis.SetValue({ key: "session:abc", value: "user:42", ttl: 3600 });
const val = await client.redis.GetValue({ key: "session:abc" });
console.log(val.value);

// Clean up
client.close();
```

## Vault / Secret Manager Integration (No Files)

```javascript
const { connect } = require("@xeze/dbr");
const vault = require("node-vault")({ endpoint: "http://127.0.0.1:8200", token: "dev-root-token" });

// Fetch certs from Vault
const secret = await vault.read("secret/data/dbrouter/certs");
const { client_cert, client_key } = secret.data.data;

// Connect directly with raw bytes — no files on disk!
const client = connect({
  host: "db.0.xeze.org:443",
  certData: Buffer.from(client_cert),
  keyData: Buffer.from(client_key),
});
```

## Local Development (no TLS)

```javascript
const client = connect({ host: "localhost:50051", insecure: true });
```

## Available Services

| Stub | Methods |
|---|---|
| `client.health` | `Check`, `CheckPostgres`, `CheckMongo`, `CheckRedis` |
| `client.postgres` | `ListDatabases`, `CreateDatabase`, `ListTables`, `ExecuteQuery`, `SelectData`, `InsertData`, `UpdateData`, `DeleteData` |
| `client.mongo` | `ListDatabases`, `ListCollections`, `InsertDocument`, `FindDocuments`, `UpdateDocument`, `DeleteDocument` |
| `client.redis` | `ListKeys`, `SetValue`, `GetValue`, `DeleteKey`, `Info` |

## License

Apache 2.0
