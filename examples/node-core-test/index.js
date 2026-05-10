const { XezeCoreClient } = require("@xeze/dbr-core");
const path = require("path");

async function run() {
  // Setup environment for testing with the deployed db-router
  process.env.DB_ROUTER_HOST = "db.0.xeze.org:443";
  process.env.DB_CERTS_DIR = path.resolve(__dirname, "../../deployer/state/certs");

  console.log("Connecting to db-router at", process.env.DB_ROUTER_HOST);
  console.log("Using certs from", process.env.DB_CERTS_DIR);

  try {
    // We use a test namespace
    const db = await XezeCoreClient.create("testapp");
    console.log("Client created successfully.");

    console.log("Initializing workspace...");
    await db.initWorkspace();

    console.log("Testing PostgreSQL...");
    // Create a table for testing
    await db.pgQuery("CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name VARCHAR(255), value VARCHAR(255));");
    
    console.log("Inserting into Postgres...");
    await db.pgInsert("test_table", { name: "example", value: "working" });
    
    console.log("Querying Postgres...");
    const pgRes = await db.pgQuery("SELECT * FROM test_table ORDER BY id DESC LIMIT 1;");
    console.log("Postgres result:", pgRes);

    console.log("Testing MongoDB...");
    const mongoRes = await db.mongoInsert("test_logs", { action: "test_run", timestamp: Date.now() });
    console.log("MongoDB inserted ID:", mongoRes);

    console.log("Testing Redis...");
    await db.redisSet("test_cache", "hello_redis", 300);
    const redisVal = await db.redisGet("test_cache");
    console.log("Redis retrieved value:", redisVal);

    console.log("All tests passed! Cleaning up...");
    db.close();
  } catch (err) {
    console.error("Error during test:", err);
  }
}

run();
