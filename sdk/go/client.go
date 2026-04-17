// Package dbrclient provides a Go gRPC client for the Xeze Database Router.
//
// Usage:
//
//	import dbr "code.xeze.org/xeze/Database-Router/sdk/go"
//
//	// File-based mTLS
//	client, err := dbr.Connect(dbr.Options{
//	    Host:     "db.0.xeze.org:443",
//	    CertFile: "client.crt",
//	    KeyFile:  "client.key",
//	})
//
//	// Raw bytes (Vault-friendly)
//	client, err := dbr.Connect(dbr.Options{
//	    Host:     "db.0.xeze.org:443",
//	    CertData: certPEM,
//	    KeyData:  keyPEM,
//	})
//
//	// Use service stubs
//	resp, err := client.Postgres.ListDatabases(ctx, &pb.ListDatabasesRequest{})
//	resp, err := client.Mongo.InsertDocument(ctx, &pb.InsertDocumentRequest{...})
//	resp, err := client.Redis.SetValue(ctx, &pb.SetValueRequest{...})
package dbrclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "code.xeze.org/xeze/Database-Router/sdk/go/proto/dbrouter"
)

// Options configures the connection to a Database Router instance.
type Options struct {
	// Host is the gRPC target, e.g. "db.0.xeze.org:443".
	Host string

	// CertFile is the path to the client certificate (.crt) for mTLS.
	CertFile string
	// KeyFile is the path to the client private key (.key) for mTLS.
	KeyFile string
	// CAFile is the optional path to a custom CA certificate.
	CAFile string

	// CertData is raw PEM bytes of the client certificate (for Vault).
	CertData []byte
	// KeyData is raw PEM bytes of the client key (for Vault).
	KeyData []byte
	// CAData is raw PEM bytes of the CA certificate.
	CAData []byte

	// Insecure uses a plaintext channel (for local dev only).
	Insecure bool
}

// Client wraps all four gRPC service stubs for the Database Router.
type Client struct {
	conn     *grpc.ClientConn
	Health   pb.HealthServiceClient
	Postgres pb.PostgresServiceClient
	Mongo    pb.MongoServiceClient
	Redis    pb.RedisServiceClient
}

// Connect establishes a gRPC connection and returns a Client.
func Connect(opts Options) (*Client, error) {
	var dialOpts []grpc.DialOption

	if opts.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig, err := buildTLSConfig(opts)
		if err != nil {
			return nil, fmt.Errorf("tls config: %w", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}

	conn, err := grpc.NewClient(opts.Host, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return &Client{
		conn:     conn,
		Health:   pb.NewHealthServiceClient(conn),
		Postgres: pb.NewPostgresServiceClient(conn),
		Mongo:    pb.NewMongoServiceClient(conn),
		Redis:    pb.NewRedisServiceClient(conn),
	}, nil
}

// Close tears down the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

func buildTLSConfig(opts Options) (*tls.Config, error) {
	// Load client certificate
	var cert tls.Certificate
	var err error

	if len(opts.CertData) > 0 && len(opts.KeyData) > 0 {
		cert, err = tls.X509KeyPair(opts.CertData, opts.KeyData)
	} else if opts.CertFile != "" && opts.KeyFile != "" {
		cert, err = tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
	} else {
		return nil, fmt.Errorf("either CertFile/KeyFile or CertData/KeyData must be provided")
	}
	if err != nil {
		return nil, fmt.Errorf("loading client cert: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA certificate if provided
	var caData []byte
	if len(opts.CAData) > 0 {
		caData = opts.CAData
	} else if opts.CAFile != "" {
		caData, err = os.ReadFile(opts.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA file: %w", err)
		}
	}

	if len(caData) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = pool
	}

	return tlsConfig, nil
}
