# @xeze/dbr-core (Node.js)

> 🚀 **Automatic Package Release on changes**

High-level, Vault-integrated database client for Node.js apps on Xeze infrastructure.

## Installation

```bash
npm install @xeze/dbr-core
```

## Quick Start

```javascript
const { XezeCoreClient } = require("@xeze/dbr-core");

async function main() {
  const db = await XezeCoreClient.create("xms");
  await db.initWorkspace();

  // Postgres
  await db.pgInsert("students", { name: "Ayush", grade: "A" });
  const rows = await db.pgQuery("SELECT * FROM students");
  console.log(rows);

  // MongoDB
  await db.mongoInsert("audit_logs", { action: "student_added" });

  // Redis
  await db.redisSet("cache:student:latest", "Ayush", 300);
  const val = await db.redisGet("cache:student:latest");
  console.log(val); // "Ayush"

  db.close();
}

main();
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VAULT_ADDR` | `http://127.0.0.1:8200` | HashiCorp Vault address |
| `VAULT_TOKEN` | `dev-root-token` | Vault authentication token |
| `DB_ROUTER_HOST` | `db.0.xeze.org:443` | Database Router gRPC host |
