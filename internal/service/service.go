// Package service defines the business-logic interfaces and their concrete
// implementations for each database backend. The gRPC server layer delegates
// all work here, keeping transport concerns separate from domain logic.
package service

import "context"

// Row is a generic key-value representation of a database row/document.
type Row map[string]interface{}

// --- PostgresService ---

type PostgresService interface {
	ListDatabases(ctx context.Context) ([]string, error)
	CreateDatabase(ctx context.Context, name string) error
	ListTables(ctx context.Context, database string) ([]string, error)
	ExecuteQuery(ctx context.Context, query, database string) (columns []string, rows []Row, rowsAffected int64, isSelect bool, err error)
	SelectData(ctx context.Context, database, table string, limit int) ([]Row, error)
	InsertData(ctx context.Context, database, table string, data Row) (insertedID string, err error)
	UpdateData(ctx context.Context, database, table, id string, data Row) (int64, error)
	DeleteData(ctx context.Context, database, table, id string) (int64, error)
	TestConnection(ctx context.Context) (host, database string, err error)
}

// --- MongoService ---

type MongoService interface {
	ListDatabases(ctx context.Context) ([]string, error)
	ListCollections(ctx context.Context, database string) ([]string, error)
	InsertDocument(ctx context.Context, database, collection string, document Row) (string, error)
	FindDocuments(ctx context.Context, database, collection string) ([]Row, error)
	UpdateDocument(ctx context.Context, database, collection, id string, update Row) (matched, modified int64, err error)
	DeleteDocument(ctx context.Context, database, collection, id string) (int64, error)
	TestConnection(ctx context.Context) (database string, err error)
}

// --- RedisService ---

type RedisService interface {
	ListKeys(ctx context.Context, pattern string) ([]string, error)
	SetValue(ctx context.Context, key, value string, ttl int) error
	GetValue(ctx context.Context, key string) (value string, ttl int, err error)
	DeleteKey(ctx context.Context, key string) (bool, error)
	Info(ctx context.Context) (dbSize int64, info string, err error)
	TestConnection(ctx context.Context) (host, port string, err error)
}

// --- HealthService ---

type HealthService interface {
	CheckAll(ctx context.Context) (postgresOK, mongoOK, redisOK bool)
}
