// Package webhandler provides HTTP handlers for the web UI test panel.
// Every handler is a pure gRPC client — it calls the db-router gRPC
// server over the network and translates the responses to JSON for the
// browser.  The web UI binary has ZERO direct database dependencies.
package webhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	pb "db-router/proto/dbrouter"

	"google.golang.org/grpc"
)

// Handler holds the four gRPC client stubs.
type Handler struct {
	pg     pb.PostgresServiceClient
	mongo  pb.MongoServiceClient
	redis  pb.RedisServiceClient
	health pb.HealthServiceClient
}

// New constructs a Handler from a shared gRPC client connection.
func New(conn grpc.ClientConnInterface) *Handler {
	return &Handler{
		pg:     pb.NewPostgresServiceClient(conn),
		mongo:  pb.NewMongoServiceClient(conn),
		redis:  pb.NewRedisServiceClient(conn),
		health: pb.NewHealthServiceClient(conn),
	}
}

// Register wires all HTTP routes onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	// Health
	mux.HandleFunc("/api/health", h.withCORS(h.handleHealth))

	// PostgreSQL
	mux.HandleFunc("/api/pg/databases", h.withCORS(h.handlePGListDatabases))
	mux.HandleFunc("/api/pg/create-db", h.withCORS(h.handlePGCreateDatabase))
	mux.HandleFunc("/api/pg/tables", h.withCORS(h.handlePGListTables))
	mux.HandleFunc("/api/pg/query", h.withCORS(h.handlePGQuery))
	mux.HandleFunc("/api/pg/select", h.withCORS(h.handlePGSelect))
	mux.HandleFunc("/api/pg/insert", h.withCORS(h.handlePGInsert))
	mux.HandleFunc("/api/pg/update", h.withCORS(h.handlePGUpdate))
	mux.HandleFunc("/api/pg/delete", h.withCORS(h.handlePGDelete))

	// MongoDB
	mux.HandleFunc("/api/mongo/databases", h.withCORS(h.handleMongoListDatabases))
	mux.HandleFunc("/api/mongo/collections", h.withCORS(h.handleMongoListCollections))
	mux.HandleFunc("/api/mongo/find", h.withCORS(h.handleMongoFind))
	mux.HandleFunc("/api/mongo/insert", h.withCORS(h.handleMongoInsert))
	mux.HandleFunc("/api/mongo/update", h.withCORS(h.handleMongoUpdate))
	mux.HandleFunc("/api/mongo/delete", h.withCORS(h.handleMongoDelete))

	// Redis
	mux.HandleFunc("/api/redis/keys", h.withCORS(h.handleRedisListKeys))
	mux.HandleFunc("/api/redis/get", h.withCORS(h.handleRedisGet))
	mux.HandleFunc("/api/redis/set", h.withCORS(h.handleRedisSet))
	mux.HandleFunc("/api/redis/delete", h.withCORS(h.handleRedisDelete))
	mux.HandleFunc("/api/redis/info", h.withCORS(h.handleRedisInfo))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (h *Handler) writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeGRPCError(w http.ResponseWriter, err error) {
	h.writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
}

func (h *Handler) decodeBody(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}
	return json.NewDecoder(r.Body).Decode(dst)
}

func (h *Handler) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
