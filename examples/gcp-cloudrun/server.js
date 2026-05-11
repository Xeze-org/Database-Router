const express = require('express');
const { XezeCoreClient } = require('@xeze/dbr-core');

const app = express();
const port = process.env.PORT || 8080;

let dbClient = null;

async function initDb() {
  if (dbClient) return dbClient;
  
  console.log("Initializing database connection...");
  console.log(`Using DB_ROUTER_HOST: ${process.env.DB_ROUTER_HOST}`);
  console.log(`Using DB_CERTS_DIR: ${process.env.DB_CERTS_DIR}`);
  
  try {
    // "gcp_demo" is our application namespace
    dbClient = await XezeCoreClient.create("gcp_demo");
    await dbClient.initWorkspace();
    console.log("Database connection established.");
    return dbClient;
  } catch (error) {
    console.error("Failed to connect to database router:", error);
    throw error;
  }
}

// Serve static files from the 'public' directory
const path = require('path');
app.use(express.static(path.join(__dirname, 'public')));

// API endpoint to run database tests
app.get('/api/test', async (req, res) => {
  try {
    const db = await initDb();
    const timestamp = new Date().toISOString();
    
    const results = {
      status: "success",
      message: "GCP Cloud Run Example App is alive!",
      tests: {}
    };

    // 1. Test PostgreSQL
    try {
      await db.pgQuery("CREATE TABLE IF NOT EXISTS page_views (id SERIAL PRIMARY KEY, view_time VARCHAR(255));");
      await db.pgInsert("page_views", { view_time: timestamp });
      const pgData = await db.pgQuery("SELECT * FROM page_views ORDER BY id DESC LIMIT 1;");
      results.tests.postgres = { status: "OK", latest_row: pgData[0] };
    } catch (e) {
      results.tests.postgres = { status: "ERROR", error: e.message };
    }

    // 2. Test MongoDB
    try {
      const mongoId = await db.mongoInsert("access_logs", { endpoint: "/", accessed_at: timestamp });
      results.tests.mongodb = { status: "OK", inserted_id: mongoId };
    } catch (e) {
      results.tests.mongodb = { status: "ERROR", error: e.message };
    }

    // 3. Test Redis
    try {
      await db.redisSet("latest_visitor", timestamp, 300);
      const redisVal = await db.redisGet("latest_visitor");
      results.tests.redis = { status: "OK", cached_value: redisVal };
    } catch (e) {
      results.tests.redis = { status: "ERROR", error: e.message };
    }

    res.json(results);
    
  } catch (error) {
    res.status(500).json({
      status: "error",
      message: "Failed to process request",
      error: error.message
    });
  }
});

// Cloud Run gracefully shuts down instances
process.on('SIGTERM', () => {
  console.log('SIGTERM signal received: closing HTTP server');
  if (dbClient) {
    dbClient.close();
  }
  process.exit(0);
});

app.listen(port, () => {
  console.log(`GCP Cloud Run Example app listening on port ${port}`);
});
