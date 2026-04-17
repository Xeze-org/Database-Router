# xeze-dbr

Java gRPC client for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router) — a unified interface for **PostgreSQL**, **MongoDB**, and **Redis** over **mTLS**.

## Install

Add the following to your `pom.xml`:

```xml
<dependency>
    <groupId>org.xeze.dbr</groupId>
    <artifactId>xeze-dbr</artifactId>
    <version>0.1.0</version>
</dependency>
```

## Quick Start

```java
import org.xeze.dbr.Options;
import org.xeze.dbr.XezeDbrClient;
import dbrouter.Dbrouter.HealthCheckRequest;
import dbrouter.Dbrouter.HealthCheckResponse;

// 1. Configure and connect
Options opts = new Options("db.0.xeze.org:443");
opts.certFile = new java.io.File("client.crt");
opts.keyFile = new java.io.File("client.key");

try (XezeDbrClient client = XezeDbrClient.connect(opts)) {

    // 2. Health check
    HealthCheckResponse health = client.health.check(HealthCheckRequest.newBuilder().build());
    System.out.println("Healthy? " + health.getOverallHealthy());
    
    // 3. PostgreSQL
    var dbResp = client.postgres.listDatabases(dbrouter.Dbrouter.ListDatabasesRequest.newBuilder().build());
    System.out.println(dbResp.getDatabasesList());
}
```

## Vault / Secret Manager Integration (No Files)

If you use HashiCorp Vault, you can pass raw certificate bytes instead of file paths:

```java
Options opts = new Options("db.0.xeze.org:443");
opts.certData = rawCertBytes;
opts.keyData = rawKeyBytes;

XezeDbrClient client = XezeDbrClient.connect(opts);
```

## Available Services

You can access the generated blocking stubs directly from the client instance:

| Stub |
|---|
| `client.health` |
| `client.postgres` |
| `client.mongo` |
| `client.redis` |

## License

Apache 2.0
