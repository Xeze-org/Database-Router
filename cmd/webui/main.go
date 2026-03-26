// cmd/webui runs an HTTP test-panel that talks to the db-router gRPC
// server over the network.  It has ZERO direct database dependencies —
// the router stays fast and independent; this binary is purely a
// browser-friendly frontend for it.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"db-router/internal/webhandler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed web
var webFS embed.FS

func main() {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	log.Printf("connecting to gRPC router at %s ...", grpcAddr)

	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	h := webhandler.New(conn)

	mux := http.NewServeMux()
	h.Register(mux)

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to create sub-filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))

	port := os.Getenv("WEBUI_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received %v — shutting down", sig)
		srv.Close()
	}()

	log.Printf("db-router Web UI  →  http://localhost:%s", port)
	log.Printf("gRPC backend      →  %s", grpcAddr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
