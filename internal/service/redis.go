package service

import (
	"context"
	"fmt"
	"time"

	"db-router/internal/db"
)

type redisService struct {
	db *db.Manager
}

// NewRedisService constructs a RedisService backed by the given db.Manager.
func NewRedisService(m *db.Manager) RedisService {
	return &redisService{db: m}
}

func (s *redisService) ListKeys(ctx context.Context, pattern string) ([]string, error) {
	if s.db.RedisClient == nil {
		return nil, ErrNotEnabled("Redis")
	}
	if pattern == "" {
		pattern = "*"
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.db.RedisClient.Keys(ctx, pattern).Result()
}

func (s *redisService) SetValue(ctx context.Context, key, value string, ttl int) error {
	if s.db.RedisClient == nil {
		return ErrNotEnabled("Redis")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var expiration time.Duration
	if ttl > 0 {
		expiration = time.Duration(ttl) * time.Second
	}
	return s.db.RedisClient.Set(ctx, key, value, expiration).Err()
}

func (s *redisService) GetValue(ctx context.Context, key string) (string, int, error) {
	if s.db.RedisClient == nil {
		return "", 0, ErrNotEnabled("Redis")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	value, err := s.db.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", 0, fmt.Errorf("key not found")
	}
	ttl, _ := s.db.RedisClient.TTL(ctx, key).Result()
	return value, int(ttl.Seconds()), nil
}

func (s *redisService) DeleteKey(ctx context.Context, key string) (bool, error) {
	if s.db.RedisClient == nil {
		return false, ErrNotEnabled("Redis")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.db.RedisClient.Del(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (s *redisService) Info(ctx context.Context) (int64, string, error) {
	if s.db.RedisClient == nil {
		return 0, "", ErrNotEnabled("Redis")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	info, err := s.db.RedisClient.Info(ctx).Result()
	if err != nil {
		return 0, "", err
	}
	dbSize, _ := s.db.RedisClient.DBSize(ctx).Result()
	return dbSize, info, nil
}

func (s *redisService) TestConnection(ctx context.Context) (string, string, error) {
	if s.db.RedisClient == nil {
		return "", "", ErrNotEnabled("Redis")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.db.RedisClient.Ping(ctx).Err(); err != nil {
		return "", "", err
	}
	return s.db.Config.Redis.Host, fmt.Sprintf("%d", s.db.Config.Redis.Port), nil
}
