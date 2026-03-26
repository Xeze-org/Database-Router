package server

import (
	"context"

	pb "db-router/proto/dbrouter"
	"db-router/internal/service"
)

// MongoServer implements the gRPC MongoServiceServer interface by
// delegating to a service.MongoService.
type MongoServer struct {
	pb.UnimplementedMongoServiceServer
	svc service.MongoService
}

func NewMongoServer(svc service.MongoService) *MongoServer {
	return &MongoServer{svc: svc}
}

func (s *MongoServer) ListDatabases(ctx context.Context, _ *pb.ListMongoDatabasesRequest) (*pb.ListMongoDatabasesResponse, error) {
	dbs, err := s.svc.ListDatabases(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ListMongoDatabasesResponse{Databases: dbs}, nil
}

func (s *MongoServer) ListCollections(ctx context.Context, req *pb.ListCollectionsRequest) (*pb.ListCollectionsResponse, error) {
	cols, err := s.svc.ListCollections(ctx, req.GetDatabase())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ListCollectionsResponse{Database: req.GetDatabase(), Collections: cols}, nil
}

func (s *MongoServer) InsertDocument(ctx context.Context, req *pb.InsertDocumentRequest) (*pb.InsertDocumentResponse, error) {
	doc := make(service.Row)
	if req.GetDocument() != nil {
		for k, v := range req.GetDocument().GetFields() {
			doc[k] = fromProtoValue(v)
		}
	}

	id, err := s.svc.InsertDocument(ctx, req.GetDatabase(), req.GetCollection(), doc)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.InsertDocumentResponse{
		Database:   req.GetDatabase(),
		Collection: req.GetCollection(),
		InsertedId: id,
	}, nil
}

func (s *MongoServer) FindDocuments(ctx context.Context, req *pb.FindDocumentsRequest) (*pb.FindDocumentsResponse, error) {
	docs, err := s.svc.FindDocuments(ctx, req.GetDatabase(), req.GetCollection())
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoDocs := make([]*pb.MongoDocument, len(docs))
	for i, d := range docs {
		protoDocs[i] = &pb.MongoDocument{Fields: rowToProtoFields(d)}
	}
	return &pb.FindDocumentsResponse{
		Database:   req.GetDatabase(),
		Collection: req.GetCollection(),
		Documents:  protoDocs,
		Count:      int64(len(docs)),
	}, nil
}

func (s *MongoServer) UpdateDocument(ctx context.Context, req *pb.UpdateDocumentRequest) (*pb.UpdateDocumentResponse, error) {
	update := make(service.Row)
	if req.GetUpdate() != nil {
		for k, v := range req.GetUpdate().GetFields() {
			update[k] = fromProtoValue(v)
		}
	}

	matched, modified, err := s.svc.UpdateDocument(ctx, req.GetDatabase(), req.GetCollection(), req.GetId(), update)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateDocumentResponse{
		Database:      req.GetDatabase(),
		Collection:    req.GetCollection(),
		MatchedCount:  matched,
		ModifiedCount: modified,
	}, nil
}

func (s *MongoServer) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	deleted, err := s.svc.DeleteDocument(ctx, req.GetDatabase(), req.GetCollection(), req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteDocumentResponse{
		Database:     req.GetDatabase(),
		Collection:   req.GetCollection(),
		DeletedCount: deleted,
	}, nil
}
