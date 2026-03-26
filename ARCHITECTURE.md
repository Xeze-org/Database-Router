# Architecture

## Overview

`database-router` is a **gRPC server** written in Go that provides a single, schema-defined interface for PostgreSQL, MongoDB, and Redis. It is structured in three distinct layers so that transport concerns (gRPC), business logic (services), and infrastructure (database connections) never mix.

```
┌──────────────────────────────────────────────────────┐
│                  gRPC Clients                        │
│         (your app, grpcurl, grpcui, …)               │
└──────────────────────────┬───────────────────────────┘
                           │  protobuf over TCP :50051
┌──────────────────────────▼───────────────────────────┐
│              cmd/main.go  —  Server bootstrap        │
│  • Loads config                                      │
│  • Wires service layer                               │
│  • Registers gRPC servers                            │
│  • Enables server reflection                         │
│  • Graceful shutdown on SIGINT / SIGTERM             │
└──────────┬───────────────┬───────────────┬───────────┘
           │               │               │
┌──────────▼──┐   ┌────────▼──┐   ┌────────▼──────────┐
│ Postgres    │   │ Mongo     │   │ Redis              │
│ Server      │   │ Server    │   │ Server             │
│             │   │           │   │                    │
│ Health      │   │           │   │                    │
│ Server      │   │           │   │                    │
└──────────┬──┘   └────────┬──┘   └────────┬───────────┘
           │     internal/server            │
           │ (gRPC ↔ service translation)  │
           │               │               │
┌──────────▼──┐   ┌────────▼──┐   ┌────────▼───────────┐
│ Postgres    │   │ Mongo     │   │ Redis               │
│ Service     │   │ Service   │   │ Service             │
│ (interface) │   │(interface)│   │  (interface)        │
│             │   │           │   │                     │
│ Health      │   │           │   │                     │
│ Service     │   │           │   │                     │
└──────────┬──┘   └────────┬──┘   └────────┬────────────┘
           │    internal/service             │
           │     (business logic)           │
           │               │               │
┌──────────▼───────────────▼───────────────▼────────────┐
│                  db.Manager                            │
│  internal/db/database.go                              │
│  • PostgresDB  *sql.DB                                │
│  • MongoDB     *mongo.Client                          │
│  • RedisClient *redis.Client                          │
└──────────┬───────────────┬───────────────┬────────────┘
           │               │               │
      PostgreSQL        MongoDB          Redis
```

---

## Layer responsibilities

### `proto/` — Contract

`dbrouter.proto` is the single source of truth for every operation the router exposes. It defines four gRPC services and all their request/response message types. The generated Go files (`dbrouter.pb.go`, `dbrouter_grpc.pb.go`) are committed so callers can import them without running `protoc`.

### `cmd/main.go` — Bootstrap

Wires all layers together, registers each gRPC server, enables server reflection (so `grpcurl`/`grpcui` work without a `.proto` file at the client), and installs a signal handler for graceful shutdown.

### `internal/server/` — Transport layer

One file per gRPC service (`postgres_server.go`, `mongo_server.go`, `redis_server.go`, `health_server.go`). Each struct embeds the generated `Unimplemented*Server` for forward compatibility and holds a reference to the matching service interface. Their only job is to:

1. Extract values from protobuf request messages
2. Call the service layer
3. Pack results back into protobuf response messages

`convert.go` contains the `service.Row ↔ *structpb.Value` helpers shared by all four servers. `errors.go` maps service-layer errors to gRPC status codes.

### `internal/service/` — Business logic layer

Defines Go interfaces (`PostgresService`, `MongoService`, `RedisService`, `HealthService`) and their concrete implementations. All SQL construction, BSON building, Redis command logic, and validation live here — completely decoupled from gRPC.

The interface separation means implementations can be replaced or mocked in tests without touching gRPC or database code.

### `internal/db/` — Infrastructure layer

`Manager` holds the three live database connections (a `*sql.DB` pool for Postgres, a `*mongo.Client`, and a `*redis.Client`). It handles connection lifecycle (open, ping, pool configuration, close). Services receive the manager via constructor injection.

### `internal/config/` — Configuration

Loads `config.json`, then applies environment variable overrides. Keeps all configuration reading in one place.

---

## Data flow — example: `PostgresService.InsertData`

```
gRPC client
  │  InsertDataRequest { database, table, data: map<string, Value> }
  ▼
postgres_server.go  InsertData()
  │  protoFieldsToRow(req.GetData())   → service.Row
  ▼
postgres.go  InsertData()
  │  validates table name
  │  GetPostgresConnection(database)
  │  builds INSERT … RETURNING id
  │  conn.QueryRowContext / ExecContext
  │  returns insertedID string
  ▼
postgres_server.go
  │  builds InsertDataResponse
  ▼
gRPC client
     InsertDataResponse { database, table, inserted_id }
```

---

## gRPC Services

### HealthService

| RPC | Request | Response |
|-----|---------|----------|
| `Check` | `HealthCheckRequest` | `HealthCheckResponse` (all three statuses) |
| `CheckPostgres` | `HealthCheckRequest` | `ConnectionStatus` |
| `CheckMongo` | `HealthCheckRequest` | `ConnectionStatus` |
| `CheckRedis` | `HealthCheckRequest` | `ConnectionStatus` |

### PostgresService

| RPC | Notes |
|-----|-------|
| `ListDatabases` | `SELECT datname FROM pg_database WHERE datistemplate = false` |
| `ListTables` | tables in `public` schema of the requested database |
| `ExecuteQuery` | arbitrary SQL; SELECT returns columns+rows, DML returns `rows_affected` |
| `SelectData` | `SELECT * FROM <table> LIMIT $1` with identifier validation |
| `InsertData` | parameterised INSERT, tries `RETURNING id` first |
| `UpdateData` | `UPDATE … SET … WHERE id = $n` |
| `DeleteData` | `DELETE … WHERE id = $1` |

### MongoService

| RPC | Notes |
|-----|-------|
| `ListDatabases` | `ListDatabaseNames` |
| `ListCollections` | `ListCollectionNames` for the given database |
| `InsertDocument` | `InsertOne` with `google.protobuf.Struct` body |
| `FindDocuments` | `Find` with empty filter |
| `UpdateDocument` | `UpdateOne` with `$set` and `ObjectIDFromHex` |
| `DeleteDocument` | `DeleteOne` by ObjectID |

### RedisService

| RPC | Notes |
|-----|-------|
| `ListKeys` | `KEYS pattern` (default `*`) |
| `SetValue` | `SET key value [EX ttl]` |
| `GetValue` | `GET` + `TTL` |
| `DeleteKey` | `DEL` |
| `Info` | raw Redis `INFO` + `DBSIZE` |

---

## Object-oriented design

The service layer uses **interface-based polymorphism**:

```go
type PostgresService interface {
    ListDatabases(ctx context.Context) ([]string, error)
    InsertData(ctx context.Context, database, table string, data Row) (string, error)
    // …
}

// concrete implementation
type postgresService struct { db *db.Manager }

func NewPostgresService(m *db.Manager) PostgresService {
    return &postgresService{db: m}
}
```

Benefits:
- The gRPC server layer depends only on the interface, never on `*postgresService` directly
- Each service is independently testable with a mock
- Swapping a backend (e.g. a read replica) requires only a new implementation of the interface

---

## Error handling

Service errors are classified in `internal/server/errors.go`:

| Condition | gRPC status code |
|-----------|-----------------|
| Backend not enabled in config | `codes.Unavailable` |
| Invalid table name / bad input | `codes.InvalidArgument` |
| Key/document not found | `codes.NotFound` |
| Database / driver error | `codes.Internal` |

---

## Key design decisions

| Decision | Rationale |
|----------|-----------|
| gRPC instead of REST | Strongly typed contract, code-generated clients, binary efficiency |
| Server reflection enabled | `grpcurl`/`grpcui` discovery without distributing the `.proto` |
| Interface-based services | Decouples business logic from transport and infrastructure |
| `structpb.Value` for dynamic rows | Avoids per-table code generation while staying protobuf-native |
| `sql.DB` connection pool for Postgres | Built-in pooling, temp connections for cross-database queries |
| Graceful shutdown | `grpcServer.GracefulStop()` drains in-flight RPCs cleanly |
| Config from file + env overrides | Secrets stay out of the image; env vars work for containers |

---

## Directory reference

```
database-router/
├── cmd/
│   └── main.go                     gRPC server bootstrap
├── proto/
│   ├── dbrouter.proto              service + message definitions
│   └── dbrouter/
│       ├── dbrouter.pb.go          generated message types
│       └── dbrouter_grpc.pb.go     generated service interfaces
├── internal/
│   ├── config/
│   │   └── config.go               JSON + env config
│   ├── db/
│   │   └── database.go             connection manager (Manager)
│   ├── service/
│   │   ├── service.go              interfaces: Postgres/Mongo/Redis/HealthService
│   │   ├── postgres.go             PostgresService implementation
│   │   ├── mongo.go                MongoService implementation
│   │   ├── redis.go                RedisService implementation
│   │   ├── health.go               HealthService implementation
│   │   └── errors.go               NotEnabledError + helpers
│   └── server/
│       ├── postgres_server.go      gRPC PostgresServiceServer
│       ├── mongo_server.go         gRPC MongoServiceServer
│       ├── redis_server.go         gRPC RedisServiceServer
│       ├── health_server.go        gRPC HealthServiceServer
│       ├── convert.go              Row ↔ structpb.Value helpers
│       └── errors.go               service error → gRPC status
├── deploy/
│   └── docker-compose.yml          standalone compose (router only)
├── docs/                           reference documentation
├── Dockerfile
├── docker-compose.yml              full-stack compose (PG + Mongo + Redis + router)
├── config.example.json
├── go.mod
└── go.sum
```
