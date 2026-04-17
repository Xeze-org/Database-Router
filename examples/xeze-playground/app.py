"""
Xeze Playground — A fun shoutboard app using all 3 databases.

PostgreSQL  → Posts (author, message, created_at)
MongoDB     → Reactions per post (likes, fire, heart, etc.)
Redis       → View counter + trending score per post
"""
import os
import sys
import time
from flask import Flask, jsonify, request, send_from_directory
from xeze_core import XezeCoreClient

app = Flask(__name__, static_folder="static", static_url_path="")

# -- Init client ---------------------------------------------------------------
try:
    db = XezeCoreClient(app_namespace="playground")
    db.init_workspace()
    print("[OK] Connected — namespace: playground")
except Exception as e:
    print(f"[ERROR] {e}")
    sys.exit(1)

# -- Create posts table --------------------------------------------------------
db.pg_query("""
    CREATE TABLE IF NOT EXISTS posts (
        id SERIAL PRIMARY KEY,
        author VARCHAR(64) NOT NULL,
        message TEXT NOT NULL,
        avatar_color VARCHAR(16) NOT NULL DEFAULT '#6c5ce7',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
""")
print("[OK] Postgres table 'posts' ready")

REACTION_TYPES = ["like", "fire", "heart", "rocket", "clap"]

# -- Helpers -------------------------------------------------------------------

def _init_reactions(post_id):
    """Ensure a reactions doc exists for a post in Mongo."""
    key = f"reactions:{post_id}"
    if db.redis_get(f"reactions_init:{post_id}"):
        return
    db.mongo_insert("reactions", {
        "post_id": str(post_id),
        "like": 0, "fire": 0, "heart": 0, "rocket": 0, "clap": 0
    })
    db.redis_set(f"reactions_init:{post_id}", "1", ttl=60 * 60 * 24 * 30)

def _get_reactions(post_id):
    """Get reaction counts from Redis cache, fallback to 0s."""
    result = {}
    for rt in REACTION_TYPES:
        val = db.redis_get(f"rxn:{post_id}:{rt}")
        result[rt] = int(val) if val else 0
    return result

def _increment_reaction(post_id, reaction):
    key = f"rxn:{post_id}:{reaction}"
    current = int(db.redis_get(key) or 0)
    db.redis_set(key, current + 1, ttl=60 * 60 * 24 * 30)
    # Also bump trending score
    trending_key = f"trending:{post_id}"
    score = int(db.redis_get(trending_key) or 0)
    db.redis_set(trending_key, score + 2, ttl=60 * 60 * 24)

def _increment_views(post_id):
    key = f"views:{post_id}"
    current = int(db.redis_get(key) or 0)
    db.redis_set(key, current + 1, ttl=60 * 60 * 24 * 7)
    # Bump trending score
    trending_key = f"trending:{post_id}"
    score = int(db.redis_get(trending_key) or 0)
    db.redis_set(trending_key, score + 1, ttl=60 * 60 * 24)

def _get_views(post_id):
    val = db.redis_get(f"views:{post_id}")
    return int(val) if val else 0

def _get_trending_score(post_id):
    val = db.redis_get(f"trending:{post_id}")
    return int(val) if val else 0

AVATAR_COLORS = [
    "#6c5ce7", "#00b894", "#e17055", "#0984e3",
    "#fd79a8", "#fdcb6e", "#55efc4", "#a29bfe"
]

# -- Routes -------------------------------------------------------------------

@app.route("/")
def serve_ui():
    return send_from_directory("static", "index.html")

@app.route("/api/posts", methods=["GET"])
def get_posts():
    """Fetch all posts with views + reactions."""
    sort = request.args.get("sort", "recent")  # recent | trending
    try:
        rows = db.pg_query(
            "SELECT id, author, message, avatar_color, created_at::text FROM posts ORDER BY id DESC LIMIT 50;"
        )
        enriched = []
        for row in rows:
            pid = int(row["id"])
            views = _get_views(pid)
            reactions = _get_reactions(pid)
            trending = _get_trending_score(pid)
            enriched.append({
                **row,
                "id": pid,
                "views": views,
                "reactions": reactions,
                "trending_score": trending
            })

        if sort == "trending":
            enriched.sort(key=lambda x: x["trending_score"], reverse=True)

        return jsonify(enriched)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/posts", methods=["POST"])
def create_post():
    """Create a new post."""
    data = request.json
    author = (data.get("author") or "Anonymous").strip()[:64]
    message = (data.get("message") or "").strip()
    if not message:
        return jsonify({"error": "Message cannot be empty"}), 400
    if len(message) > 500:
        return jsonify({"error": "Message too long (max 500 chars)"}), 400

    import hashlib
    color_idx = int(hashlib.md5(author.encode()).hexdigest(), 16) % len(AVATAR_COLORS)
    avatar_color = AVATAR_COLORS[color_idx]

    try:
        inserted_id = db.pg_insert("posts", {
            "author": author,
            "message": message,
            "avatar_color": avatar_color
        })
        # Log to Mongo
        db.mongo_insert("activity", {
            "event": "post_created",
            "post_id": str(inserted_id),
            "author": author,
            "ts": str(int(time.time()))
        })
        return jsonify({"message": "Posted!", "id": inserted_id, "avatar_color": avatar_color})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/posts/<int:post_id>/view", methods=["POST"])
def view_post(post_id):
    """Increment view count."""
    _increment_views(post_id)
    return jsonify({"views": _get_views(post_id)})

@app.route("/api/posts/<int:post_id>/react", methods=["POST"])
def react_post(post_id):
    """Add a reaction to a post."""
    data = request.json
    reaction = data.get("reaction", "like")
    if reaction not in REACTION_TYPES:
        return jsonify({"error": "Invalid reaction type"}), 400
    _increment_reaction(post_id, reaction)
    # Log to Mongo
    db.mongo_insert("activity", {
        "event": "reaction_added",
        "post_id": str(post_id),
        "reaction": reaction,
        "ts": str(int(time.time()))
    })
    return jsonify({"reactions": _get_reactions(post_id)})

@app.route("/api/stats")
def get_stats():
    """Get board-level stats."""
    try:
        count_rows = db.pg_query("SELECT COUNT(*) as count FROM posts;")
        total_posts = int(count_rows[0]["count"]) if count_rows else 0
        return jsonify({
            "total_posts": total_posts,
            "namespace": db.app_namespace,
            "dbs": {
                "postgres": db.pg_db,
                "mongo": db.mongo_db,
                "redis_prefix": db.redis_prefix
            }
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5051, debug=True)
