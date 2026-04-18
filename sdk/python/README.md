# xeze-dbr

> 🚀 **Automatic Package Release on changes**

Python gRPC client for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router) — a unified interface for **PostgreSQL**, **MongoDB**, and **Redis** over **mTLS**.

## Install

```bash
pip install xeze-dbr
```

## Quick Start

```python
from xeze_dbr import connect

# Connect with mTLS (production)
client = connect(
    host="db.0.xeze.org:443",
    cert_path="client.crt",
    key_path="client.key",
)

# Health check
health = client.health.Check(client.pb2.HealthCheckRequest())
print(health.overall_healthy)  # True

# PostgreSQL — list databases
dbs = client.postgres.ListDatabases(client.pb2.ListDatabasesRequest())
print(dbs.databases)

# PostgreSQL — insert data
from google.protobuf import struct_pb2
resp = client.postgres.InsertData(client.pb2.InsertDataRequest(
    database="mydb",
    table="users",
    data={
        "name": struct_pb2.Value(string_value="Alice"),
        "age": struct_pb2.Value(number_value=28),
    }
))
print(resp.inserted_id)

# MongoDB — insert document
doc = struct_pb2.Struct()
doc.update({"event": "signup", "user": "Alice"})
resp = client.mongo.InsertDocument(client.pb2.InsertDocumentRequest(
    database="mydb",
    collection="events",
    document=doc,
))
print(resp.inserted_id)

# Redis — set & get
client.redis.SetValue(client.pb2.SetValueRequest(
    key="session:abc", value="user:42", ttl=3600
))
val = client.redis.GetValue(client.pb2.GetValueRequest(key="session:abc"))
print(val.value)

# Clean up
client.close()
```

## Vault / Secret Manager Integration (No Files)

If you use HashiCorp Vault, AWS Secrets Manager, or Doppler, you can pass raw certificate bytes instead of file paths using `cert_data` and `key_data` parameters:

```python
import hvac
from xeze_dbr import connect

# 1. Get raw bytes from Vault
vault = hvac.Client(url="http://127.0.0.1:8200", token="dev-root-token")
secret = vault.secrets.kv.v2.read_secret_version(path="dbrouter/certs")

# 2. Connect directly (no files on disk!)
client = connect(
    host="db.0.xeze.org:443",
    cert_data=secret["data"]["data"]["client_cert"].encode(),
    key_data=secret["data"]["data"]["client_key"].encode(),
)
```

## Local Development (no TLS)

```python
client = connect(host="localhost:50051", insecure=True)
```

## Available Services

| Stub | Methods |
|---|---|
| `client.health` | `Check`, `CheckPostgres`, `CheckMongo`, `CheckRedis` |
| `client.postgres` | `ListDatabases`, `CreateDatabase`, `ListTables`, `ExecuteQuery`, `SelectData`, `InsertData`, `UpdateData`, `DeleteData` |
| `client.mongo` | `ListDatabases`, `ListCollections`, `InsertDocument`, `FindDocuments`, `UpdateDocument`, `DeleteDocument` |
| `client.redis` | `ListKeys`, `SetValue`, `GetValue`, `DeleteKey`, `Info` |

## License

Apache 2.0
