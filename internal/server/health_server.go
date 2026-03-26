package server

import (
	"context"
	"fmt"

	pb "db-router/proto/dbrouter"
	"db-router/internal/service"
)

// HealthServer implements the gRPC HealthServiceServer interface.
type HealthServer struct {
	pb.UnimplementedHealthServiceServer
	health service.HealthService
	pg     service.PostgresService
	mongo  service.MongoService
	redis  service.RedisService
}

func NewHealthServer(h service.HealthService, pg service.PostgresService, mg service.MongoService, rd service.RedisService) *HealthServer {
	return &HealthServer{health: h, pg: pg, mongo: mg, redis: rd}
}

func (s *HealthServer) Check(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	pgOK, mgOK, rdOK := s.health.CheckAll(ctx)

	return &pb.HealthCheckResponse{
		OverallHealthy: pgOK && mgOK && rdOK,
		Postgres:       buildStatus(s.pg, ctx),
		Mongo:          buildMongoStatus(s.mongo, ctx),
		Redis:          buildRedisStatus(s.redis, ctx),
	}, nil
}

func (s *HealthServer) CheckPostgres(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.ConnectionStatus, error) {
	return buildStatus(s.pg, ctx), nil
}

func (s *HealthServer) CheckMongo(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.ConnectionStatus, error) {
	return buildMongoStatus(s.mongo, ctx), nil
}

func (s *HealthServer) CheckRedis(ctx context.Context, _ *pb.HealthCheckRequest) (*pb.ConnectionStatus, error) {
	return buildRedisStatus(s.redis, ctx), nil
}

func buildStatus(pg service.PostgresService, ctx context.Context) *pb.ConnectionStatus {
	host, database, err := pg.TestConnection(ctx)
	if err != nil {
		if service.IsNotEnabled(err) {
			return &pb.ConnectionStatus{Status: "disabled", Enabled: false}
		}
		return &pb.ConnectionStatus{Status: "disconnected", Enabled: true, Error: err.Error()}
	}
	return &pb.ConnectionStatus{Status: "connected", Enabled: true, Host: host, Database: database}
}

func buildMongoStatus(mg service.MongoService, ctx context.Context) *pb.ConnectionStatus {
	database, err := mg.TestConnection(ctx)
	if err != nil {
		if service.IsNotEnabled(err) {
			return &pb.ConnectionStatus{Status: "disabled", Enabled: false}
		}
		return &pb.ConnectionStatus{Status: "disconnected", Enabled: true, Error: err.Error()}
	}
	return &pb.ConnectionStatus{Status: "connected", Enabled: true, Database: database}
}

func buildRedisStatus(rd service.RedisService, ctx context.Context) *pb.ConnectionStatus {
	host, port, err := rd.TestConnection(ctx)
	if err != nil {
		if service.IsNotEnabled(err) {
			return &pb.ConnectionStatus{Status: "disabled", Enabled: false}
		}
		return &pb.ConnectionStatus{Status: "disconnected", Enabled: true, Error: err.Error()}
	}
	return &pb.ConnectionStatus{Status: "connected", Enabled: true, Host: host, Port: fmt.Sprintf("%s", port)}
}
