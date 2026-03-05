package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestRedisConnection returns the live connection status.
func (h *Handler) TestRedisConnection(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.DB.RedisClient.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "disconnected", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "connected",
		"host":   h.DB.Config.Redis.Host,
		"port":   h.DB.Config.Redis.Port,
	})
}

// ListRedisKeys lists keys matching an optional ?pattern= query param (default *).
func (h *Handler) ListRedisKeys(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	pattern := c.DefaultQuery("pattern", "*")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keys, err := h.DB.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys, "count": len(keys)})
}

// SetRedisValue sets a key/value with an optional TTL in seconds.
func (h *Handler) SetRedisValue(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	var req struct {
		Key   string `json:"key"   binding:"required"`
		Value string `json:"value" binding:"required"`
		TTL   int    `json:"ttl"` // 0 = no expiry
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	if req.TTL > 0 {
		err = h.DB.RedisClient.Set(ctx, req.Key, req.Value, time.Duration(req.TTL)*time.Second).Err()
	} else {
		err = h.DB.RedisClient.Set(ctx, req.Key, req.Value, 0).Err()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": req.Key, "value": req.Value, "ttl": req.TTL})
}

// GetRedisValue retrieves a value by key name.
func (h *Handler) GetRedisValue(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	key := c.Param("key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value, err := h.DB.RedisClient.Get(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}
	ttl, _ := h.DB.RedisClient.TTL(ctx, key).Result()
	c.JSON(http.StatusOK, gin.H{"key": key, "value": value, "ttl": int(ttl.Seconds())})
}

// DeleteRedisKey removes a key.
func (h *Handler) DeleteRedisKey(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	key := c.Param("key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := h.DB.RedisClient.Del(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "deleted": result > 0})
}

// GetRedisInfo returns raw Redis server info and the current key count.
func (h *Handler) GetRedisInfo(c *gin.Context) {
	if h.DB.RedisClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not enabled"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	info, err := h.DB.RedisClient.Info(ctx).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dbSize, _ := h.DB.RedisClient.DBSize(ctx).Result()
	c.JSON(http.StatusOK, gin.H{"db_size": dbSize, "info": info})
}
