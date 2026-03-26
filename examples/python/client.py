"""
DbRouterClient — thin Python wrapper around the db-router HTTP proxy API.

The web UI panel (cmd/webui, default :8080) exposes a REST API that proxies
every call to the gRPC router (:50051).  This client talks to that REST layer
so no protobuf compilation is needed.

Usage:
    from client import DbRouterClient
    c = DbRouterClient("http://localhost:8080")
    print(c.health())
    print(c.pg_databases())
"""

import requests


class DbRouterError(Exception):
    pass


class DbRouterClient:
    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        timeout: int = 10,
        tls_ca: str | None = None,
        tls_cert: tuple[str, str] | None = None,
    ):
        """
        Args:
            base_url:  URL of the db-router HTTP proxy (cmd/webui).
            timeout:   Request timeout in seconds.
            tls_ca:    Path to CA certificate (PEM) for verifying the proxy's TLS cert.
            tls_cert:  (cert_path, key_path) tuple — client certificate for mTLS.

        Example (plain HTTP, default):
            c = DbRouterClient()

        Example (HTTPS webui with mTLS):
            c = DbRouterClient(
                "https://db.yourdomain.com:8080",
                tls_ca="certs/ca.crt",
                tls_cert=("certs/client.crt", "certs/client.key"),
            )
        """
        self.base    = base_url.rstrip("/")
        self.timeout = timeout
        self.session = requests.Session()
        if tls_ca:
            self.session.verify = tls_ca
        if tls_cert:
            self.session.cert = tls_cert  # (cert_path, key_path)

    # ── internal ──────────────────────────────────────────────────────────────

    def _get(self, path: str, **params) -> dict:
        r = self.session.get(f"{self.base}{path}", params=params, timeout=self.timeout)
        return self._unwrap(r)

    def _post(self, path: str, body: dict) -> dict:
        r = self.session.post(f"{self.base}{path}", json=body, timeout=self.timeout)
        return self._unwrap(r)

    def _delete(self, path: str, **params) -> dict:
        r = self.session.delete(f"{self.base}{path}", params=params, timeout=self.timeout)
        return self._unwrap(r)

    @staticmethod
    def _unwrap(r: requests.Response) -> dict:
        data = r.json()
        if "error" in data:
            raise DbRouterError(data["error"])
        return data

    # ── Health ────────────────────────────────────────────────────────────────

    def health(self) -> dict:
        """Returns connection status for all three backends."""
        return self._get("/api/health")

    # ── PostgreSQL ────────────────────────────────────────────────────────────

    def pg_databases(self) -> list[str]:
        return self._get("/api/pg/databases")["databases"]

    def pg_create_database(self, name: str) -> dict:
        return self._post("/api/pg/create-db", {"name": name})

    def pg_tables(self, database: str) -> list[str]:
        return self._get("/api/pg/tables", db=database)["tables"]

    def pg_query(self, query: str, database: str = "") -> dict:
        """Execute any SQL. SELECT returns {columns, rows, count}; DML returns {rows_affected}."""
        return self._post("/api/pg/query", {"query": query, "database": database})

    def pg_select(self, database: str, table: str, limit: int = 100) -> list[dict]:
        return self._get("/api/pg/select", db=database, table=table, limit=limit)["data"]

    def pg_insert(self, database: str, table: str, data: dict) -> dict:
        return self._post("/api/pg/insert", {"database": database, "table": table, "data": data})

    def pg_update(self, database: str, table: str, id: str, data: dict) -> dict:
        return self._post("/api/pg/update", {"database": database, "table": table, "id": id, "data": data})

    def pg_delete(self, database: str, table: str, id: str) -> dict:
        return self._delete("/api/pg/delete", db=database, table=table, id=id)

    # ── MongoDB ───────────────────────────────────────────────────────────────

    def mongo_databases(self) -> list[str]:
        return self._get("/api/mongo/databases")["databases"]

    def mongo_collections(self, database: str) -> list[str]:
        return self._get("/api/mongo/collections", db=database)["collections"]

    def mongo_find(self, database: str, collection: str) -> list[dict]:
        return self._get("/api/mongo/find", db=database, col=collection)["documents"]

    def mongo_insert(self, database: str, collection: str, document: dict) -> dict:
        return self._post("/api/mongo/insert", {"database": database, "collection": collection, "document": document})

    def mongo_update(self, database: str, collection: str, id: str, update: dict) -> dict:
        return self._post("/api/mongo/update", {"database": database, "collection": collection, "id": id, "update": update})

    def mongo_delete(self, database: str, collection: str, id: str) -> dict:
        return self._delete("/api/mongo/delete", db=database, col=collection, id=id)

    # ── Redis ─────────────────────────────────────────────────────────────────

    def redis_keys(self, pattern: str = "*") -> list[str]:
        return self._get("/api/redis/keys", pattern=pattern)["keys"]

    def redis_get(self, key: str) -> dict:
        return self._get("/api/redis/get", key=key)

    def redis_set(self, key: str, value: str, ttl: int = 0) -> dict:
        return self._post("/api/redis/set", {"key": key, "value": value, "ttl": ttl})

    def redis_delete(self, key: str) -> dict:
        return self._delete("/api/redis/delete", key=key)

    def redis_info(self) -> dict:
        return self._get("/api/redis/info")
