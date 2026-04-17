"""
dbr-core-test: A test web app that exercises all three databases
(PostgreSQL, MongoDB, Redis) through the xeze-dbr-core unified client.
"""
import os
import sys
from flask import Flask, jsonify, request, send_from_directory
from xeze_core import XezeCoreClient

app = Flask(__name__, static_folder="static", static_url_path="")

# -- Initialize the unified client --------------------------------------------
try:
    db = XezeCoreClient(app_namespace="dbr_test")
    db.init_workspace()
    print("[OK] XezeCoreClient connected and workspace provisioned.")
except Exception as e:
    print(f"[ERROR] Failed to initialize XezeCoreClient: {e}")
    sys.exit(1)

# -- Create the Postgres table on startup -------------------------------------
try:
    db.pg_query("""
        CREATE TABLE IF NOT EXISTS contacts (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL,
            role VARCHAR(100) DEFAULT 'member',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    """)
    print("[OK] Postgres table 'contacts' ready.")
except Exception as e:
    print(f"[WARN] Table init: {e}")

# -- Serve UI -----------------------------------------------------------------

@app.route("/")
def serve_ui():
    return send_from_directory("static", "index.html")

# -- POSTGRESQL ENDPOINTS -----------------------------------------------------

@app.route("/api/pg/contacts", methods=["GET"])
def pg_list():
    """List all contacts from Postgres."""
    try:
        rows = db.pg_query("SELECT id, name, email, role FROM contacts ORDER BY id ASC;")
        return jsonify(rows)
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/pg/contacts", methods=["POST"])
def pg_add():
    """Insert a contact into Postgres."""
    data = request.json
    name = data.get("name", "").strip()
    email = data.get("email", "").strip()
    role = data.get("role", "member").strip()
    if not name or not email:
        return jsonify({"error": "name and email are required"}), 400
    try:
        inserted_id = db.pg_insert("contacts", {"name": name, "email": email, "role": role})
        return jsonify({"message": "Contact added", "id": inserted_id})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# -- MONGODB ENDPOINTS --------------------------------------------------------

@app.route("/api/mongo/logs", methods=["GET"])
def mongo_list():
    """Fetch recent audit logs from MongoDB (insert-only demo)."""
    return jsonify({"info": "MongoDB is insert-only in this demo. Check the insert endpoint."})

@app.route("/api/mongo/logs", methods=["POST"])
def mongo_add():
    """Insert an audit log into MongoDB."""
    data = request.json
    action = data.get("action", "").strip()
    detail = data.get("detail", "").strip()
    if not action:
        return jsonify({"error": "action is required"}), 400
    try:
        inserted_id = db.mongo_insert("audit_logs", {
            "action": action,
            "detail": detail,
            "source": "dbr-core-test-ui"
        })
        return jsonify({"message": "Log inserted", "id": inserted_id})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# -- REDIS ENDPOINTS ----------------------------------------------------------

@app.route("/api/redis/cache", methods=["POST"])
def redis_set():
    """Set a cache key in Redis."""
    data = request.json
    key = data.get("key", "").strip()
    value = data.get("value", "").strip()
    ttl = data.get("ttl", 3600)
    if not key or not value:
        return jsonify({"error": "key and value are required"}), 400
    try:
        db.redis_set(key, value, ttl=int(ttl))
        return jsonify({"message": f"Key '{key}' set with TTL {ttl}s"})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route("/api/redis/cache", methods=["GET"])
def redis_get():
    """Get a cache value from Redis."""
    key = request.args.get("key", "").strip()
    if not key:
        return jsonify({"error": "key query param is required"}), 400
    try:
        value = db.redis_get(key)
        if value is None:
            return jsonify({"key": key, "value": None, "found": False})
        return jsonify({"key": key, "value": value, "found": True})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# -- HEALTH --------------------------------------------------------------------

@app.route("/api/health")
def health():
    return jsonify({
        "status": "ok",
        "namespace": db.app_namespace,
        "pg_db": db.pg_db,
        "mongo_db": db.mongo_db,
        "redis_prefix": db.redis_prefix
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5050, debug=True)
