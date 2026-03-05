// Package config handles loading and overriding application configuration
// from config.json and environment variables.
package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// Config is the root configuration structure.
type Config struct {
	Postgres PostgresConfig `json:"postgres"`
	Mongo    MongoConfig    `json:"mongo"`
	Redis    RedisConfig    `json:"redis"`
}

// PostgresConfig holds PostgreSQL connection parameters.
type PostgresConfig struct {
	Enabled  string `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

// MongoConfig holds MongoDB connection parameters.
type MongoConfig struct {
	Enabled  string `json:"enabled"`
	URI      string `json:"uri"`
	Database string `json:"database"`
}

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Enabled  string `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// Load reads config.json (creating a default if absent) then applies any
// environment variable overrides. Returns a ready-to-use *Config.
func Load() *Config {
	const configFile = "config.json"

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Printf("config.json not found — creating defaults")
		createDefault(configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	overrideFromEnv(&cfg)
	return &cfg
}

// overrideFromEnv replaces config values with environment variables when set.
// This is used by Docker Compose to inject secrets without touching config.json.
func overrideFromEnv(cfg *Config) {
	// PostgreSQL
	if v := os.Getenv("POSTGRES_ENABLED"); v != "" {
		cfg.Postgres.Enabled = v
	}
	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		cfg.Postgres.Host = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Postgres.Port = p
		}
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		cfg.Postgres.User = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		cfg.Postgres.Password = v
	}
	if v := os.Getenv("POSTGRES_DB"); v != "" {
		cfg.Postgres.Database = v
	}

	// MongoDB
	if v := os.Getenv("MONGO_ENABLED"); v != "" {
		cfg.Mongo.Enabled = v
	}
	if v := os.Getenv("MONGO_URI"); v != "" {
		cfg.Mongo.URI = v
	}
	if v := os.Getenv("MONGO_DATABASE"); v != "" {
		cfg.Mongo.Database = v
	}

	// Redis
	if v := os.Getenv("REDIS_ENABLED"); v != "" {
		cfg.Redis.Enabled = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = p
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
}

// createDefault writes a config.json with sensible local development defaults.
func createDefault(filename string) {
	def := Config{
		Postgres: PostgresConfig{
			Enabled:  "true",
			Host:     "localhost",
			Port:     5432,
			User:     "admin",
			Password: "3rHb6NmA5jUc8Tg1",
			Database: "test",
			SSLMode:  "disable",
		},
		Mongo: MongoConfig{
			Enabled:  "true",
			URI:      "mongodb://admin:8fKx9Pq2LmZ4vW7y@mongo.0.xeze.org:27017/xeze_test?authSource=admin",
			Database: "xeze_test",
		},
		Redis: RedisConfig{
			Enabled:  "true",
			Host:     "localhost",
			Port:     6379,
			Password: "p9Kj2mT7vWcD4s8X",
			DB:       0,
		},
	}

	data, err := json.MarshalIndent(def, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal default config: %v", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Fatalf("failed to write default config: %v", err)
	}
	log.Printf("created default config: %s", filename)
}
