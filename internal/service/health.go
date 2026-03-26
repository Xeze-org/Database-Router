package service

import "context"

type healthService struct {
	pg    PostgresService
	mongo MongoService
	redis RedisService
}

// NewHealthService constructs a HealthService that aggregates the connection
// status of all three database backends.
func NewHealthService(pg PostgresService, mongo MongoService, redis RedisService) HealthService {
	return &healthService{pg: pg, mongo: mongo, redis: redis}
}

func (s *healthService) CheckAll(ctx context.Context) (bool, bool, bool) {
	_, _, pgErr := s.pg.TestConnection(ctx)
	_, mgErr := s.mongo.TestConnection(ctx)
	_, _, rdErr := s.redis.TestConnection(ctx)
	return pgErr == nil, mgErr == nil, rdErr == nil
}
