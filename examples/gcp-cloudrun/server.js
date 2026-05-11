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
app.use(express.json()); // Parse JSON bodies

// GET /api/posts - Fetch posts from Redis or MongoDB
app.get('/api/posts', async (req, res) => {
  try {
    const db = await initDb();
    
    // Check Redis cache first
    const cachedPosts = await db.redisGet("blog:posts");
    if (cachedPosts) {
      console.log("Serving posts from Redis cache");
      return res.json({ status: "success", source: "redis", data: JSON.parse(cachedPosts) });
    }

    // Fallback to MongoDB
    console.log("Serving posts from MongoDB");
    const posts = await db.mongoFind("posts");
    
    // Sort posts by date descending (assuming id contains timestamp or sorting in JS)
    posts.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

    // Cache in Redis for 10 minutes
    await db.redisSet("blog:posts", JSON.stringify(posts), 600);

    res.json({ status: "success", source: "mongodb", data: posts });
  } catch (error) {
    res.status(500).json({
      status: "error",
      message: "Failed to fetch posts",
      error: error.message
    });
  }
});

// POST /api/posts - Create a new post
app.post('/api/posts', async (req, res) => {
  try {
    const { title, content, author } = req.body;
    
    if (!title || !content) {
      return res.status(400).json({ status: "error", message: "Title and content are required." });
    }

    const db = await initDb();
    const newPost = {
      title,
      content,
      author: author || "Anonymous",
      createdAt: new Date().toISOString()
    };

    // Insert into MongoDB
    const insertedId = await db.mongoInsert("posts", newPost);
    newPost._id = insertedId;

    // Invalidate Redis cache
    await db._client.redis.DeleteKey({ key: `${db.redisPrefix}blog:posts` });

    res.status(201).json({ status: "success", data: newPost });
  } catch (error) {
    res.status(500).json({
      status: "error",
      message: "Failed to create post",
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
