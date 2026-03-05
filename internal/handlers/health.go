package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestAllConnections checks every database and returns a combined health report.
func (h *Handler) TestAllConnections(c *gin.Context) {
	results := gin.H{
		"postgres": h.testPostgres(),
		"mongo":    h.testMongo(),
		"redis":    h.testRedis(),
	}

	allHealthy := true
	if pg, ok := results["postgres"].(gin.H); ok {
		if pg["status"] != "connected" {
			allHealthy = false
		}
	}
	if mg, ok := results["mongo"].(gin.H); ok {
		if mg["status"] != "connected" {
			allHealthy = false
		}
	}
	if rd, ok := results["redis"].(gin.H); ok {
		if rd["status"] != "connected" {
			allHealthy = false
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"overall_status": allHealthy,
		"connections":    results,
	})
}

// ── per-database test helpers ─────────────────────────────────────────────────

func (h *Handler) testPostgres() gin.H {
	if h.DB.PostgresDB == nil {
		return gin.H{"status": "disabled", "enabled": false}
	}
	if err := h.DB.PostgresDB.Ping(); err != nil {
		return gin.H{"status": "disconnected", "enabled": true, "error": err.Error()}
	}
	return gin.H{
		"status":   "connected",
		"enabled":  true,
		"database": h.DB.Config.Postgres.Database,
		"host":     h.DB.Config.Postgres.Host,
	}
}

func (h *Handler) testMongo() gin.H {
	if h.DB.MongoDB == nil {
		return gin.H{"status": "disabled", "enabled": false}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.DB.MongoDB.Ping(ctx, nil); err != nil {
		return gin.H{"status": "disconnected", "enabled": true, "error": err.Error()}
	}
	return gin.H{
		"status":   "connected",
		"enabled":  true,
		"database": h.DB.Config.Mongo.Database,
	}
}

func (h *Handler) testRedis() gin.H {
	if h.DB.RedisClient == nil {
		return gin.H{"status": "disabled", "enabled": false}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.DB.RedisClient.Ping(ctx).Err(); err != nil {
		return gin.H{"status": "disconnected", "enabled": true, "error": err.Error()}
	}
	return gin.H{
		"status":  "connected",
		"enabled": true,
		"host":    h.DB.Config.Redis.Host,
		"port":    h.DB.Config.Redis.Port,
	}
}
