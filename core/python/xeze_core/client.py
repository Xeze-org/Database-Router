import os
import hvac
import grpc
from xeze_dbr import connect
from google.protobuf import struct_pb2


class XezeCoreClient:
    """
    Unified client for Postgres, MongoDB, and Redis over mTLS via HashiCorp Vault.
    Enforces database isolation per application namespace.
    """

    def __init__(self, app_namespace):
        if not app_namespace or not isinstance(app_namespace, str):
            raise ValueError("A strict app_namespace (e.g., 'xms', 'selfnote') is required.")

        self.app_namespace = app_namespace
        # Automatic isolated routing paths
        self.pg_db = f"{app_namespace}_pg"
        self.mongo_db = f"{app_namespace}_mongo"
        self.redis_prefix = f"{app_namespace}:"

        self._connect_via_vault()

    def _connect_via_vault(self):
        """Silently handles Vault authentication and memory-only cert loading."""
        vault_addr = os.environ.get("VAULT_ADDR", "http://127.0.0.1:8200")
        vault_token = os.environ.get("VAULT_TOKEN", "dev-root-token")
        host = os.environ.get("DB_ROUTER_HOST", "db.0.xeze.org:443")

        vault = hvac.Client(url=vault_addr, token=vault_token)
        if not vault.is_authenticated():
            raise Exception("Critical: Vault authentication failed.")

        secret = vault.secrets.kv.v2.read_secret_version(path="dbrouter/certs")
        cert_data = secret["data"]["data"]["client_cert"].encode()
        key_data = secret["data"]["data"]["client_key"].encode()

        self.client = connect(host=host, cert_data=cert_data, key_data=key_data)

    def init_workspace(self):
        """
        Attempts to create the isolated Postgres database for this namespace.
        Safe to call on application startup.
        """
        try:
            req = self.client.pb2.CreateDatabaseRequest(name=self.pg_db)
            self.client.postgres.CreateDatabase(req)
            print(f"[OK] Provisioned workspace: {self.pg_db}")
        except grpc.RpcError as e:
            if "already exists" in e.details().lower():
                pass  # Expected behavior if already provisioned
            else:
                print(f"[WARN] Workspace check failed: {e.details()}")

    def _pack_pg_dict(self, data_dict):
        """Internal helper for PostgreSQL Protobuf typing."""
        packed = {}
        for k, v in data_dict.items():
            if isinstance(v, str):
                packed[k] = struct_pb2.Value(string_value=v)
            elif isinstance(v, bool):
                packed[k] = struct_pb2.Value(bool_value=v)
            elif isinstance(v, (int, float)):
                packed[k] = struct_pb2.Value(number_value=float(v))
            else:
                packed[k] = struct_pb2.Value(string_value=str(v))
        return packed

    # --- POSTGRESQL API -------------------------------------------------------

    def pg_query(self, query):
        """Executes raw SQL returning a list of native Python dictionaries."""
        req = self.client.pb2.ExecuteQueryRequest(database=self.pg_db, query=query)
        resp = self.client.postgres.ExecuteQuery(req)

        results = []
        for row in resp.rows:
            row_data = {}
            for key, val in row.fields.items():
                if val.HasField("string_value"):
                    row_data[key] = val.string_value
                elif val.HasField("number_value"):
                    row_data[key] = val.number_value
                elif val.HasField("bool_value"):
                    row_data[key] = val.bool_value
            results.append(row_data)
        return results

    def pg_insert(self, table, data_dict):
        """Inserts a native Python dict into PostgreSQL."""
        req = self.client.pb2.InsertDataRequest(
            database=self.pg_db,
            table=table,
            data=self._pack_pg_dict(data_dict)
        )
        resp = self.client.postgres.InsertData(req)
        return resp.inserted_id

    # --- MONGODB API ----------------------------------------------------------

    def mongo_insert(self, collection, doc_dict):
        """Inserts a native Python dictionary into MongoDB."""
        doc = struct_pb2.Struct()
        doc.update(doc_dict)

        req = self.client.pb2.InsertDocumentRequest(
            database=self.mongo_db,
            collection=collection,
            document=doc
        )
        resp = self.client.mongo.InsertDocument(req)
        return resp.inserted_id

    # --- REDIS API ------------------------------------------------------------

    def redis_set(self, key, value, ttl=3600):
        """Sets a prefixed key with a default 1-hour TTL."""
        namespaced_key = f"{self.redis_prefix}{key}"
        req = self.client.pb2.SetValueRequest(key=namespaced_key, value=str(value), ttl=ttl)
        self.client.redis.SetValue(req)

    def redis_get(self, key):
        """Fetches a prefixed key gracefully."""
        namespaced_key = f"{self.redis_prefix}{key}"
        try:
            req = self.client.pb2.GetValueRequest(key=namespaced_key)
            return self.client.redis.GetValue(req).value
        except grpc.RpcError as e:
            if "not found" in e.details().lower():
                return None
            raise e

    def close(self):
        """Closes the underlying gRPC connection."""
        self.client.close()
