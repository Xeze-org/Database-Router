package server

import (
	"context"

	pb "db-router/proto/dbrouter"
	"db-router/internal/service"
)

// PostgresServer implements the gRPC PostgresServiceServer interface by
// delegating to a service.PostgresService.
type PostgresServer struct {
	pb.UnimplementedPostgresServiceServer
	svc service.PostgresService
}

func NewPostgresServer(svc service.PostgresService) *PostgresServer {
	return &PostgresServer{svc: svc}
}

func (s *PostgresServer) ListDatabases(ctx context.Context, _ *pb.ListDatabasesRequest) (*pb.ListDatabasesResponse, error) {
	dbs, err := s.svc.ListDatabases(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ListDatabasesResponse{Databases: dbs}, nil
}

func (s *PostgresServer) CreateDatabase(ctx context.Context, req *pb.CreateDatabaseRequest) (*pb.CreateDatabaseResponse, error) {
	if err := s.svc.CreateDatabase(ctx, req.GetName()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateDatabaseResponse{
		Name:    req.GetName(),
		Message: "Database created successfully",
	}, nil
}

func (s *PostgresServer) ListTables(ctx context.Context, req *pb.ListTablesRequest) (*pb.ListTablesResponse, error) {
	tables, err := s.svc.ListTables(ctx, req.GetDatabase())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ListTablesResponse{Database: req.GetDatabase(), Tables: tables}, nil
}

func (s *PostgresServer) ExecuteQuery(ctx context.Context, req *pb.ExecuteQueryRequest) (*pb.ExecuteQueryResponse, error) {
	columns, rows, affected, isSelect, err := s.svc.ExecuteQuery(ctx, req.GetQuery(), req.GetDatabase())
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.ExecuteQueryResponse{}
	if isSelect {
		resp.Columns = columns
		resp.Count = int64(len(rows))
		resp.Rows = make([]*pb.QueryResultRow, len(rows))
		for i, r := range rows {
			resp.Rows[i] = &pb.QueryResultRow{Fields: rowToProtoFields(r)}
		}
	} else {
		resp.RowsAffected = affected
		resp.Message = "Command executed successfully"
	}
	return resp, nil
}

func (s *PostgresServer) SelectData(ctx context.Context, req *pb.SelectDataRequest) (*pb.SelectDataResponse, error) {
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 100
	}
	rows, err := s.svc.SelectData(ctx, req.GetDatabase(), req.GetTable(), limit)
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoRows := make([]*pb.QueryResultRow, len(rows))
	for i, r := range rows {
		protoRows[i] = &pb.QueryResultRow{Fields: rowToProtoFields(r)}
	}
	return &pb.SelectDataResponse{
		Database: req.GetDatabase(),
		Table:    req.GetTable(),
		Data:     protoRows,
		Count:    int64(len(rows)),
	}, nil
}

func (s *PostgresServer) InsertData(ctx context.Context, req *pb.InsertDataRequest) (*pb.InsertDataResponse, error) {
	data := protoFieldsToRow(req.GetData())
	id, err := s.svc.InsertData(ctx, req.GetDatabase(), req.GetTable(), data)
	if err != nil {
		return nil, toGRPCError(err)
	}

	resp := &pb.InsertDataResponse{
		Database: req.GetDatabase(),
		Table:    req.GetTable(),
	}
	if id != "" {
		resp.InsertedId = id
	} else {
		resp.Message = "Row inserted successfully"
	}
	return resp, nil
}

func (s *PostgresServer) UpdateData(ctx context.Context, req *pb.UpdateDataRequest) (*pb.UpdateDataResponse, error) {
	data := protoFieldsToRow(req.GetData())
	affected, err := s.svc.UpdateData(ctx, req.GetDatabase(), req.GetTable(), req.GetId(), data)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateDataResponse{
		Database:     req.GetDatabase(),
		Table:        req.GetTable(),
		Id:           req.GetId(),
		RowsAffected: affected,
	}, nil
}

func (s *PostgresServer) DeleteData(ctx context.Context, req *pb.DeleteDataRequest) (*pb.DeleteDataResponse, error) {
	affected, err := s.svc.DeleteData(ctx, req.GetDatabase(), req.GetTable(), req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteDataResponse{
		Database:     req.GetDatabase(),
		Table:        req.GetTable(),
		Id:           req.GetId(),
		RowsAffected: affected,
	}, nil
}
