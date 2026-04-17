import os
import sys
import grpc
import hvac
from flask import Flask, jsonify, request, send_from_directory
from xeze_dbr import connect
from google.protobuf import struct_pb2

app = Flask(__name__, static_folder="static", static_url_path="")

DB_NAME = "unified_db"
TABLE_NAME = "todos"

# ── Fetch certs from HashiCorp Vault ──────────────────────────────────────────
VAULT_ADDR = os.environ.get("VAULT_ADDR", "http://127.0.0.1:8200")
VAULT_TOKEN = os.environ.get("VAULT_TOKEN", "dev-root-token")
DB_ROUTER_HOST = os.environ.get("DB_ROUTER_HOST", "db.0.xeze.org:443")

print(f"🔐 Connecting to Vault at {VAULT_ADDR}...")
vault = hvac.Client(url=VAULT_ADDR, token=VAULT_TOKEN)

if not vault.is_authenticated():
    print("❌ Vault authentication failed!")
    sys.exit(1)

try:
    secret = vault.secrets.kv.v2.read_secret_version(path="dbrouter/certs", raise_on_deleted_version=True)
    cert_data = secret["data"]["data"]["client_cert"].encode()
    key_data = secret["data"]["data"]["client_key"].encode()
    print("✅ Certificates loaded from Vault!")
except Exception as e:
    print(f"❌ Failed to read certs from Vault: {e}")
    sys.exit(1)

# Connect to db-router using raw cert bytes from Vault — no files needed!
try:
    client = connect(host=DB_ROUTER_HOST, cert_data=cert_data, key_data=key_data)
except Exception as e:
    print(f"Error connecting to db-router: {e}")
    sys.exit(1)

def init_db():
    create_sql = f"""
    CREATE TABLE IF NOT EXISTS {TABLE_NAME} (
        id SERIAL PRIMARY KEY,
        task VARCHAR(255) NOT NULL,
        completed BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    """
    client.postgres.ExecuteQuery(client.pb2.ExecuteQueryRequest(database=DB_NAME, query=create_sql))

# Initialize table on startup
try:
    init_db()
    print("Database initialized successfully.")
except Exception as e:
    print(f"Failed to init db: {e}")

@app.route("/")
def serve_ui():
    return send_from_directory("static", "index.html")

@app.route("/api/todos", methods=["GET"])
def get_todos():
    try:
        resp = client.postgres.ExecuteQuery(client.pb2.ExecuteQueryRequest(
            database=DB_NAME,
            query=f"SELECT id, task, completed FROM {TABLE_NAME} ORDER BY id ASC;"
        ))
        
        todos = []
        for row in resp.rows:
            todos.append({
                "id": int(row.fields["id"].number_value),
                "task": row.fields["task"].string_value,
                "completed": row.fields["completed"].bool_value
            })
        return jsonify(todos)
    except grpc.RpcError as e:
        return jsonify({"error": e.details()}), 500

@app.route("/api/todos", methods=["POST"])
def add_todo():
    data = request.json
    task = data.get("task", "").strip()
    if not task:
        return jsonify({"error": "Task cannot be empty"}), 400
        
    try:
        req = client.pb2.InsertDataRequest(
            database=DB_NAME,
            table=TABLE_NAME,
            data={
                "task": struct_pb2.Value(string_value=task),
                "completed": struct_pb2.Value(bool_value=False)
            }
        )
        resp = client.postgres.InsertData(req)
        return jsonify({"message": resp.message, "id": resp.inserted_id})
    except grpc.RpcError as e:
        return jsonify({"error": e.details()}), 500

@app.route("/api/todos/<int:todo_id>", methods=["PUT"])
def toggle_todo(todo_id):
    data = request.json
    completed = data.get("completed", True)
    
    try:
        req = client.pb2.UpdateDataRequest(
            database=DB_NAME,
            table=TABLE_NAME,
            id=str(todo_id),
            data={
                "completed": struct_pb2.Value(bool_value=completed)
            }
        )
        resp = client.postgres.UpdateData(req)
        if resp.rows_affected > 0:
            return jsonify({"success": True})
        return jsonify({"error": "Todo not found"}), 404
    except grpc.RpcError as e:
        return jsonify({"error": e.details()}), 500

@app.route("/api/todos/<int:todo_id>", methods=["DELETE"])
def delete_todo(todo_id):
    try:
        req = client.pb2.DeleteDataRequest(
            database=DB_NAME,
            table=TABLE_NAME,
            id=str(todo_id)
        )
        resp = client.postgres.DeleteData(req)
        if resp.rows_affected > 0:
            return jsonify({"success": True})
        return jsonify({"error": "Todo not found"}), 404
    except grpc.RpcError as e:
        return jsonify({"error": e.details()}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
