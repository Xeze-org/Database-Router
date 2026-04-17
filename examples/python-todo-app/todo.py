import argparse
import sys
import grpc
from xeze_dbr import pb2 as dbrouter_pb2
from xeze_dbr import connect
from google.protobuf import struct_pb2

# Database and table to use for the Todo app
DB_NAME = "unified_db"
TABLE_NAME = "todos"

def get_channel():
    """Create an mTLS connected client."""
    try:
        return connect(host='db.0.xeze.org:443', cert_path='client.crt', key_path='client.key')
    except Exception as e:
        print(f"Error connecting to db-router: {e}")
        sys.exit(1)

def init_db(stub):
    """Create the todos table if it doesn't exist."""
    create_sql = f"""
    CREATE TABLE IF NOT EXISTS {TABLE_NAME} (
        id SERIAL PRIMARY KEY,
        task VARCHAR(255) NOT NULL,
        completed BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    """
    try:
        stub.ExecuteQuery(dbrouter_pb2.ExecuteQueryRequest(
            database=DB_NAME,
            query=create_sql
        ))
    except grpc.RpcError as e:
        print(f"Failed to initialize database: {e.details()}")
        sys.exit(1)

def add_todo(stub, task):
    """Add a new task to the database."""
    req = dbrouter_pb2.InsertDataRequest(
        database=DB_NAME,
        table=TABLE_NAME,
        data={
            "task": struct_pb2.Value(string_value=task),
            "completed": struct_pb2.Value(bool_value=False)
        }
    )
    try:
        resp = stub.InsertData(req)
        print(f"✅ Created task: '{task}' (ID: {resp.inserted_id})")
    except grpc.RpcError as e:
        print(f"Failed to add task: {e.details()}")

def list_todos(stub, show_all=True):
    """List tasks from the database."""
    # We use a custom query to order them by ID
    query = f"SELECT id, task, completed FROM {TABLE_NAME} ORDER BY id ASC;"
    try:
        resp = stub.ExecuteQuery(dbrouter_pb2.ExecuteQueryRequest(
            database=DB_NAME,
            query=query
        ))
        
        todos = resp.rows
        if not todos:
            print("📭 No tasks found! Enjoy your day.")
            return

        print("\n📋 Your Todo List:")
        print("-" * 40)
        for row in todos:
            fields = row.fields
            # Safely extract fields (protobuf struct mapping)
            t_id = int(fields["id"].number_value)
            task = fields["task"].string_value
            completed = fields["completed"].bool_value
            
            if not show_all and completed:
                continue
                
            status = "✅" if completed else "[]"
            print(f"{t_id:3d} | {status} | {task}")
        print("-" * 40 + "\n")
    except grpc.RpcError as e:
        print(f"Failed to list tasks: {e.details()}")

def complete_todo(stub, task_id):
    """Mark a task as completed."""
    req = dbrouter_pb2.UpdateDataRequest(
        database=DB_NAME,
        table=TABLE_NAME,
        id=str(task_id),
        data={
            "completed": struct_pb2.Value(bool_value=True)
        }
    )
    try:
        resp = stub.UpdateData(req)
        if resp.rows_affected > 0:
            print(f"✅ Marked task {task_id} as complete!")
        else:
            print(f"⚠️ Task {task_id} not found.")
    except grpc.RpcError as e:
        print(f"Failed to complete task: {e.details()}")

def delete_todo(stub, task_id):
    """Delete a task."""
    req = dbrouter_pb2.DeleteDataRequest(
        database=DB_NAME,
        table=TABLE_NAME,
        id=str(task_id)
    )
    try:
        resp = stub.DeleteData(req)
        if resp.rows_affected > 0:
            print(f"🗑️ Deleted task {task_id}.")
        else:
            print(f"⚠️ Task {task_id} not found.")
    except grpc.RpcError as e:
        print(f"Failed to delete task: {e.details()}")

def main():
    parser = argparse.ArgumentParser(description="A simple Todo CLI using db-router Postgres over mTLS")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Add command
    parser_add = subparsers.add_parser("add", help="Add a new task")
    parser_add.add_argument("task", type=str, help="The task description")

    # List command
    parser_list = subparsers.add_parser("list", help="List all tasks")
    parser_list.add_argument("--pending", action="store_true", help="Show only pending tasks")

    # Complete command
    parser_complete = subparsers.add_parser("complete", help="Mark a task as completed")
    parser_complete.add_argument("id", type=int, help="The task ID")

    # Delete command
    parser_delete = subparsers.add_parser("delete", help="Delete a task")
    parser_delete.add_argument("id", type=int, help="The task ID")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    # Setup connection and stub
    client = get_channel()
    pg_stub = client.postgres
    
    # Ensure table exists
    init_db(pg_stub)

    # Route command
    if args.command == "add":
        add_todo(pg_stub, args.task)
    elif args.command == "list":
        list_todos(pg_stub, show_all=not args.pending)
    elif args.command == "complete":
        complete_todo(pg_stub, args.id)
    elif args.command == "delete":
        delete_todo(pg_stub, args.id)

if __name__ == "__main__":
    main()
