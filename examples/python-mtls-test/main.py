import grpc
import dbrouter_pb2
import dbrouter_pb2_grpc
from google.protobuf import struct_pb2

def run():
    # 1. Load mTLS certificates
    with open('client.key', 'rb') as f:
        client_key = f.read()
    with open('client.crt', 'rb') as f:
        client_cert = f.read()

    # 2. Create mTLS credentials
    credentials = grpc.ssl_channel_credentials(
        private_key=client_key,
        certificate_chain=client_cert
    )

    # 3. Connect to db-router via Caddy
    #    Target port 443; the domain must match the cert/Caddyfile
    print("Connecting to db.0.xeze.org:443...")
    channel = grpc.secure_channel('db.0.xeze.org:443', credentials)

    # 4. Create clients
    health_stub = dbrouter_pb2_grpc.HealthServiceStub(channel)
    pg_stub = dbrouter_pb2_grpc.PostgresServiceStub(channel)
    mongo_stub = dbrouter_pb2_grpc.MongoServiceStub(channel)
    redis_stub = dbrouter_pb2_grpc.RedisServiceStub(channel)

    # --- Test 1: Health Check ---
    print("\n[1] Testing Health Service...")
    resp = health_stub.Check(dbrouter_pb2.HealthCheckRequest())
    print(f"Overall Healthy: {resp.overall_healthy}")
    print(f"Postgres Status: {resp.postgres.status}")

    # --- Test 2: PostgreSQL Insert ---
    print("\n[2] Testing PostgreSQL Insert...")
    try:
        # We need to setup a table first (assuming public schema)
        create_sql = "CREATE TABLE IF NOT EXISTS test_users (id SERIAL PRIMARY KEY, name VARCHAR(50), age INT);"
        pg_stub.ExecuteQuery(dbrouter_pb2.ExecuteQueryRequest(
            database="unified_db",
            query=create_sql
        ))
        
        # Now insert some data
        insert_req = dbrouter_pb2.InsertDataRequest(
            database="unified_db",
            table="test_users",
            data={
                "name": struct_pb2.Value(string_value="Alice_mTLS"),
                "age": struct_pb2.Value(number_value=28)
            }
        )
        insert_resp = pg_stub.InsertData(insert_req)
        print(f"PostgreSQL Insert Result: {insert_resp.message} (ID: {insert_resp.inserted_id})")
        
        # Let's verify it's there
        select_resp = pg_stub.SelectData(dbrouter_pb2.SelectDataRequest(
            database="unified_db",
            table="test_users"
        ))
        print(f"Rows found in Postgres: {len(select_resp.data)}")
    except Exception as e:
        print(f"PostgreSQL Test Failed: {e}")

    # --- Test 3: MongoDB Insert ---
    print("\n[3] Testing MongoDB Insert...")
    try:
        doc_struct = struct_pb2.Struct()
        doc_struct.update({"event": "login", "user": "Alice_mTLS", "secure": True})
        
        mongo_req = dbrouter_pb2.InsertDocumentRequest(
            database="unified_db",
            collection="test_events",
            document=doc_struct
        )
        mongo_resp = mongo_stub.InsertDocument(mongo_req)
        print(f"MongoDB Inserted ID: {mongo_resp.inserted_id}")
    except Exception as e:
        print(f"MongoDB Test Failed: {e}")

    # --- Test 4: Redis Set & Get ---
    print("\n[4] Testing Redis...")
    try:
        redis_stub.SetValue(dbrouter_pb2.SetValueRequest(
            key="test:session:alice",
            value="active_mtls_session",
            ttl=3600
        ))
        redis_get = redis_stub.GetValue(dbrouter_pb2.GetValueRequest(key="test:session:alice"))
        print(f"Redis Retrieved Value: {redis_get.value}")
    except Exception as e:
        print(f"Redis Test Failed: {e}")

if __name__ == '__main__':
    run()
