package server

import (
	"strings"

	"db-router/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}
	if service.IsNotEnabled(err) {
		return status.Error(codes.Unavailable, err.Error())
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "invalid table name"),
		strings.Contains(msg, "no data to update"),
		strings.Contains(msg, "limit must be"),
		strings.Contains(msg, "invalid ID format"):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	}
	return status.Error(codes.Internal, msg)
}
