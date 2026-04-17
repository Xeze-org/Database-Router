# xeze-dbr-core

Official unified database wrapper for the Xeze infrastructure. Provides a single, heavily-abstracted client for **PostgreSQL**, **MongoDB**, and **Redis** over mTLS via HashiCorp Vault.

## Installation

Add the following to your `pom.xml`:

```xml
<dependency>
    <groupId>org.xeze.dbr.core</groupId>
    <artifactId>xeze-dbr-core</artifactId>
    <version>0.1.0</version>
</dependency>
```

## Quick Start

```java
import org.xeze.dbr.core.XezeCoreClient;

import java.util.Map;
import java.util.HashMap;

public class Main {
    public static void main(String[] args) throws Exception {
        try (XezeCoreClient db = new XezeCoreClient("xms")) {
            
            // Ensures the 'xms_pg' database exists
            db.initWorkspace(); 
            
            // Postgres - working directly with Java Maps
            Map<String, Object> pgData = new HashMap<>();
            pgData.put("name", "Ayush");
            pgData.put("grade", "A");
            db.pgInsert("students", pgData);
            
            // MongoDB
            Map<String, Object> mongoDoc = new HashMap<>();
            mongoDoc.put("action", "student_added");
            mongoDoc.put("timestamp", "2026-04-05");
            db.mongoInsert("audit_logs", mongoDoc);
            
            // Redis
            db.redisSet("cache:student:latest", "Ayush", 300);
        }
    }
}
```

## Environment Variables

| Variable           | Default                    | Description                |
| ------------------ | -------------------------- | -------------------------- |
| `VAULT_ADDR`       | `http://127.0.0.1:8200`   | HashiCorp Vault address    |
| `VAULT_TOKEN`      | `dev-root-token`           | Vault authentication token |
| `DB_ROUTER_HOST`   | `db.0.xeze.org:443`       | Database Router gRPC host  |

## Architecture

- **Database-per-Service isolation** — each `appNamespace` gets its own `{ns}_pg`, `{ns}_mongo`, and `{ns}:` Redis prefix.
- **Vault mTLS** — client certs are fetched from Vault KV at `dbrouter/certs` via standard Java `HttpClient` and loaded in-memory only.
- **gRPC/Protobuf abstraction** — all Protobuf packing is handled internally; developers work with native `java.util.Map`.

## License

Apache 2.0
