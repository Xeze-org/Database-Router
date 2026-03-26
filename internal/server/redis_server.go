package server

import (
	"context"

	pb "db-router/proto/dbrouter"
	"db-router/internal/service"
)

// RedisServer implements the gRPC RedisServiceServer interface by
// delegating to a service.RedisService.
type RedisServer struct {
	pb.UnimplementedRedisServiceServer
	svc service.RedisService
}

func NewRedisServer(svc service.RedisService) *RedisServer {
	return &RedisServer{svc: svc}
}

func (s *RedisServer) ListKeys(ctx context.Context, req *pb.ListKeysRequest) (*pb.ListKeysResponse, error) {
	pattern := req.GetPattern()
	if pattern == "" {
		pattern = "*"
	}
	keys, err := s.svc.ListKeys(ctx, pattern)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ListKeysResponse{Keys: keys, Count: int64(len(keys))}, nil
}

func (s *RedisServer) SetValue(ctx context.Context, req *pb.SetValueRequest) (*pb.SetValueResponse, error) {
	err := s.svc.SetValue(ctx, req.GetKey(), req.GetValue(), int(req.GetTtl()))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SetValueResponse{
		Key:   req.GetKey(),
		Value: req.GetValue(),
		Ttl:   req.GetTtl(),
	}, nil
}

func (s *RedisServer) GetValue(ctx context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	value, ttl, err := s.svc.GetValue(ctx, req.GetKey())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetValueResponse{
		Key:   req.GetKey(),
		Value: value,
		Ttl:   int32(ttl),
	}, nil
}

func (s *RedisServer) DeleteKey(ctx context.Context, req *pb.DeleteKeyRequest) (*pb.DeleteKeyResponse, error) {
	deleted, err := s.svc.DeleteKey(ctx, req.GetKey())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteKeyResponse{Key: req.GetKey(), Deleted: deleted}, nil
}

func (s *RedisServer) Info(ctx context.Context, _ *pb.RedisInfoRequest) (*pb.RedisInfoResponse, error) {
	dbSize, info, err := s.svc.Info(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RedisInfoResponse{DbSize: dbSize, Info: info}, nil
}
