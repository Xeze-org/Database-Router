package main

import (
	"log"
	"net/http"
	"os"

	"db-router/internal/config"
	"db-router/internal/db"
	"db-router/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	dbManager := db.New(cfg)
	defer dbManager.Close()

	h := handlers.New(dbManager)

	router := gin.Default()
	router.Use(handlers.CORS())

	// Basic health — no auth needed, used by load balancers / Caddy upstreams.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "go-db-manager"})
	})

	v1 := router.Group("/api/v1")
	{
		// ── PostgreSQL ────────────────────────────────────────────────────────
		pg := v1.Group("/postgres")
		{
			pg.GET("/databases", h.ListPostgresDatabases)
			pg.GET("/tables/:database", h.ListPostgresTables)
			pg.POST("/query", h.ExecutePostgresQuery)
			pg.GET("/select/:database/:table", h.SelectPostgresData)
			pg.POST("/insert/:database/:table", h.InsertPostgresData)
			pg.PUT("/update/:database/:table/:id", h.UpdatePostgresData)
			pg.DELETE("/delete/:database/:table/:id", h.DeletePostgresData)
		}

		// ── MongoDB ───────────────────────────────────────────────────────────
		mongo := v1.Group("/mongo")
		{
			mongo.GET("/databases", h.ListMongoDatabases)
			mongo.GET("/collections/:database", h.ListMongoCollections)
			mongo.POST("/insert/:database/:collection", h.InsertMongoDocument)
			mongo.GET("/find/:database/:collection", h.FindMongoDocuments)
			mongo.PUT("/update/:database/:collection/:id", h.UpdateMongoDocument)
			mongo.DELETE("/delete/:database/:collection/:id", h.DeleteMongoDocument)
		}

		// ── Redis ─────────────────────────────────────────────────────────────
		redis := v1.Group("/redis")
		{
			redis.GET("/keys", h.ListRedisKeys)
			redis.POST("/set", h.SetRedisValue)
			redis.GET("/get/:key", h.GetRedisValue)
			redis.DELETE("/delete/:key", h.DeleteRedisKey)
			redis.GET("/info", h.GetRedisInfo)
		}

		// ── Connection tests ──────────────────────────────────────────────────
		test := v1.Group("/test")
		{
			test.GET("/postgres", h.TestPostgresConnection)
			test.GET("/mongo", h.TestMongoConnection)
			test.GET("/redis", h.TestRedisConnection)
			test.GET("/all", h.TestAllConnections)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go Database Manager starting on :%s", port)
	log.Printf("PostgreSQL: %s | MongoDB: %s | Redis: %s",
		cfg.Postgres.Enabled, cfg.Mongo.Enabled, cfg.Redis.Enabled)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
