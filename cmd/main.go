package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"db-router/internal/config"
	"db-router/internal/db"
	"db-router/internal/server"
	"db-router/internal/service"
	"db-router/internal/tlsconfig"
	pb "db-router/proto/dbrouter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	dbManager := db.New(cfg)
	defer dbManager.Close()

	// ── TLS / mTLS credentials ───────────────────────────────────────────
	tlsLoader := tlsconfig.New(cfg.TLS)
	creds, err := tlsLoader.Build()
	if err != nil {
		log.Fatalf("TLS configuration error: %v", err)
	}

	// ── Service layer (OOP interfaces) ───────────────────────────────────
	pgSvc := service.NewPostgresService(dbManager)
	mgSvc := service.NewMongoService(dbManager)
	rdSvc := service.NewRedisService(dbManager)
	healthSvc := service.NewHealthService(pgSvc, mgSvc, rdSvc)

	// ── gRPC server wiring ───────────────────────────────────────────────
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	pb.RegisterPostgresServiceServer(grpcServer, server.NewPostgresServer(pgSvc))
	pb.RegisterMongoServiceServer(grpcServer, server.NewMongoServer(mgSvc))
	pb.RegisterRedisServiceServer(grpcServer, server.NewRedisServer(rdSvc))
	pb.RegisterHealthServiceServer(grpcServer, server.NewHealthServer(healthSvc, pgSvc, mgSvc, rdSvc))

	// Server reflection lets tools like grpcurl discover services at runtime.
	reflection.Register(grpcServer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on :%s: %v", port, err)
	}

	log.Printf("gRPC Database Router starting on :%s  [%s]", port, tlsLoader.Mode())
	log.Printf("PostgreSQL: %s | MongoDB: %s | Redis: %s",
		cfg.Postgres.Enabled, cfg.Mongo.Enabled, cfg.Redis.Enabled)

	// Graceful shutdown on SIGINT / SIGTERM.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received %v — shutting down gracefully", sig)
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
