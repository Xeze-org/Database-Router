// Package tlsconfig builds gRPC transport credentials from the TLS section
// of the application config. It supports three modes:
//
//  1. Plain-text (TLS disabled) — development / trusted internal networks.
//  2. Server-side TLS — clients verify the server; no client cert required.
//  3. Mutual TLS (mTLS) — both sides present and verify certificates.
package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"db-router/internal/config"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Loader is an OOP wrapper around TLS credential construction.
// Instantiate it once at startup and call Build() to obtain the
// credentials that are passed directly to grpc.NewServer().
type Loader struct {
	cfg config.TLSConfig
}

// New creates a Loader for the given TLS configuration.
func New(cfg config.TLSConfig) *Loader {
	return &Loader{cfg: cfg}
}

// Build returns the appropriate gRPC server credentials:
//   - If TLS is disabled → insecure (plain-text).
//   - If CA file is absent → server-side TLS only.
//   - If CA file is present → mTLS with the configured ClientAuth policy.
func (l *Loader) Build() (credentials.TransportCredentials, error) {
	if !l.cfg.Enabled {
		return insecure.NewCredentials(), nil
	}

	if l.cfg.CertFile == "" || l.cfg.KeyFile == "" {
		return nil, fmt.Errorf("tls.cert_file and tls.key_file are required when TLS is enabled")
	}

	cert, err := tls.LoadX509KeyPair(l.cfg.CertFile, l.cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("loading server certificate: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	// mTLS: load the CA bundle that signed client certificates.
	if l.cfg.CAFile != "" {
		caCert, err := os.ReadFile(l.cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA file %q: %w", l.cfg.CAFile, err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("CA file %q contains no valid PEM certificates", l.cfg.CAFile)
		}

		tlsCfg.ClientCAs = caPool
		tlsCfg.ClientAuth = parseClientAuth(l.cfg.ClientAuth)
	}

	return credentials.NewTLS(tlsCfg), nil
}

// Mode returns a human-readable description of the current TLS mode.
func (l *Loader) Mode() string {
	if !l.cfg.Enabled {
		return "plain-text (TLS disabled)"
	}
	if l.cfg.CAFile == "" {
		return "server-side TLS"
	}
	return fmt.Sprintf("mTLS (client-auth=%s)", l.cfg.ClientAuth)
}

// parseClientAuth maps the config string to a crypto/tls constant.
// Defaults to RequireAndVerifyClientCert when unrecognised or empty,
// which is the safest choice once a CA file is provided.
func parseClientAuth(s string) tls.ClientAuthType {
	switch s {
	case "none":
		return tls.NoClientCert
	case "request":
		return tls.RequestClientCert
	case "require":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.RequireAndVerifyClientCert
	}
}
