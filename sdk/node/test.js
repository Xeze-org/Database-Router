/**
 * Quick test: connect to db-router via Vault and run a health check.
 * Usage: node test.js
 */
const { connect } = require("./index");
const fs = require("fs");

async function main() {
  console.log("🔐 Connecting to db-router via mTLS...");

  // Use cert files from examples/secrets/certs/
  const certPath = "../../examples/secrets/certs/client.crt";
  const keyPath = "../../examples/secrets/certs/client.key";

  if (!fs.existsSync(certPath) || !fs.existsSync(keyPath)) {
    console.error("❌ Cert files not found. Copy them to examples/secrets/certs/");
    process.exit(1);
  }

  const client = connect({
    host: "db.0.xeze.org:443",
    certPath,
    keyPath,
  });

  try {
    // Health check
    console.log("\n📋 Health Check:");
    const health = await client.health.Check();
    console.log("   Overall healthy:", health.overallHealthy);
    console.log("   Postgres:", health.postgres.status);
    console.log("   Mongo:", health.mongo.status);
    console.log("   Redis:", health.redis.status);

    // PostgreSQL — list databases
    console.log("\n🐘 PostgreSQL Databases:");
    const dbs = await client.postgres.ListDatabases();
    console.log("  ", dbs.databases);

    // Redis — info
    console.log("\n🔴 Redis Info:");
    const info = await client.redis.Info();
    console.log("   DB size:", info.dbSize);

    console.log("\n✅ All tests passed!");
  } catch (err) {
    console.error("❌ Error:", err.message);
  } finally {
    client.close();
  }
}

main();
