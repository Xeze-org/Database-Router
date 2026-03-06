// Package db manages all database connections (PostgreSQL, MongoDB, Redis)
// and exposes them via a single Manager struct.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"db-router/internal/config"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Manager holds all active database connections and the shared config.
type Manager struct {
	Config      *config.Config
	PostgresDB  *sql.DB
	MongoDB     *mongo.Client
	RedisClient *redis.Client
	ctx         context.Context
}

// Ctx returns the root context used when initialising connections.
func (m *Manager) Ctx() context.Context {
	return m.ctx
}

// New initialises all enabled database connections and returns a Manager.
// Connections that fail are logged but do not abort startup — the handler
// returns 503 when that database is unavailable.
func New(cfg *config.Config) *Manager {
	m := &Manager{
		Config: cfg,
		ctx:    context.Background(),
	}

	if cfg.Postgres.Enabled == "true" {
		if err := m.initPostgres(); err != nil {
			log.Printf("PostgreSQL init failed: %v", err)
		} else {
			log.Println("PostgreSQL connected")
		}
	}

	if cfg.Mongo.Enabled == "true" {
		if err := m.initMongo(); err != nil {
			log.Printf("MongoDB init failed: %v", err)
		} else {
			log.Println("MongoDB connected")
		}
	}

	if cfg.Redis.Enabled == "true" {
		if err := m.initRedis(); err != nil {
			log.Printf("Redis init failed: %v", err)
		} else {
			log.Println("Redis connected")
		}
	}

	return m
}

func (m *Manager) initPostgres() error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		m.Config.Postgres.Host,
		m.Config.Postgres.Port,
		m.Config.Postgres.User,
		m.Config.Postgres.Password,
		m.Config.Postgres.Database,
		m.Config.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Connection pool settings — tune for your workload.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	m.PostgresDB = db
	return nil
}

// GetPostgresConnection returns a connection to a specific PostgreSQL database.
// If dbname is empty or matches the default database, returns the existing connection.
// Otherwise, creates a temporary connection for the specified database.
// Caller is responsible for closing temporary connections.
func (m *Manager) GetPostgresConnection(dbname string) (*sql.DB, bool, error) {
	if m.PostgresDB == nil {
		return nil, false, fmt.Errorf("PostgreSQL not enabled")
	}

	// If no database specified or same as default, use existing connection
	if dbname == "" || dbname == m.Config.Postgres.Database {
		return m.PostgresDB, false, nil
	}

	// Create temporary connection to the specified database
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		m.Config.Postgres.Host,
		m.Config.Postgres.Port,
		m.Config.Postgres.User,
		m.Config.Postgres.Password,
		dbname,
		m.Config.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, false, err
	}

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, false, err
	}

	return db, true, nil
}

func (m *Manager) initMongo() error {
	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.Config.Mongo.URI))
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	m.MongoDB = client
	return nil
}

func (m *Manager) initRedis() error {
	m.RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", m.Config.Redis.Host, m.Config.Redis.Port),
		Password: m.Config.Redis.Password,
		DB:       m.Config.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	return m.RedisClient.Ping(ctx).Err()
}

// Close gracefully shuts down all active database connections.
func (m *Manager) Close() {
	if m.PostgresDB != nil {
		m.PostgresDB.Close()
		log.Println("PostgreSQL connection closed")
	}
	if m.MongoDB != nil {
		m.MongoDB.Disconnect(m.ctx)
		log.Println("MongoDB connection closed")
	}
	if m.RedisClient != nil {
		m.RedisClient.Close()
		log.Println("Redis connection closed")
	}
}
