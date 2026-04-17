// Package xezecore provides a high-level, Vault-integrated client for the
// Xeze Database Router. It enforces database-per-service isolation via
// the app_namespace parameter and handles mTLS certificate loading from
// HashiCorp Vault automatically.
//
// Usage:
//
//	client, err := xezecore.New("xms")
//	defer client.Close()
//
//	client.InitWorkspace(ctx)
//	client.PgQuery(ctx, "SELECT * FROM users")
//	client.PgInsert(ctx, "users", map[string]interface{}{"name": "Ayush"})
//	client.MongoInsert(ctx, "logs", map[string]interface{}{"action": "login"})
//	client.RedisSet(ctx, "session:abc", "user_123", 3600)
//	val, _ := client.RedisGet(ctx, "session:abc")
package xezecore

import (
	"context"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	dbr "code.xeze.org/xeze/Database-Router/sdk/go"
	pb "code.xeze.org/xeze/Database-Router/sdk/go/proto/dbrouter"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client is the unified, namespace-isolated database client.
type Client struct {
	AppNamespace string
	PgDB         string
	MongoDB      string
	RedisPrefix  string
	dbr          *dbr.Client
}

// New creates a XezeCoreClient connected via Vault mTLS.
func New(appNamespace string) (*Client, error) {
	if appNamespace == "" {
		return nil, fmt.Errorf("app_namespace is required")
	}

	c := &Client{
		AppNamespace: appNamespace,
		PgDB:         appNamespace + "_pg",
		MongoDB:      appNamespace + "_mongo",
		RedisPrefix:  appNamespace + ":",
	}

	if err := c.connectViaVault(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connectViaVault() error {
	vaultAddr := getEnv("VAULT_ADDR", "http://127.0.0.1:8200")
	vaultToken := getEnv("VAULT_TOKEN", "dev-root-token")
	host := getEnv("DB_ROUTER_HOST", "db.0.xeze.org:443")

	config := vault.DefaultConfig()
	config.Address = vaultAddr
	vc, err := vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}
	vc.SetToken(vaultToken)

	secret, err := vc.KVv2("secret").Get(context.Background(), "dbrouter/certs")
	if err != nil {
		return fmt.Errorf("vault read: %w", err)
	}

	certPEM := []byte(secret.Data["client_cert"].(string))
	keyPEM := []byte(secret.Data["client_key"].(string))

	client, err := dbr.Connect(dbr.Options{
		Host:     host,
		CertData: certPEM,
		KeyData:  keyPEM,
	})
	if err != nil {
		return fmt.Errorf("dbr connect: %w", err)
	}

	c.dbr = client
	return nil
}

// InitWorkspace creates the isolated Postgres database. Safe for startup.
func (c *Client) InitWorkspace(ctx context.Context) {
	_, err := c.dbr.Postgres.CreateDatabase(ctx, &pb.CreateDatabaseRequest{Name: c.PgDB})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && strings.Contains(strings.ToLower(st.Message()), "already exists") {
			return
		}
		fmt.Printf("[WARN] Workspace check failed: %v\n", err)
		return
	}
	fmt.Printf("[OK] Provisioned workspace: %s\n", c.PgDB)
}

// --- PostgreSQL API ---

// PgQuery executes raw SQL and returns results as a slice of maps.
func (c *Client) PgQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	resp, err := c.dbr.Postgres.ExecuteQuery(ctx, &pb.ExecuteQueryRequest{
		Database: c.PgDB,
		Query:    query,
	})
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, row := range resp.Rows {
		rowData := make(map[string]interface{})
		for key, val := range row.Fields {
			switch val.Kind.(type) {
			case *structpb.Value_StringValue:
				rowData[key] = val.GetStringValue()
			case *structpb.Value_NumberValue:
				rowData[key] = val.GetNumberValue()
			case *structpb.Value_BoolValue:
				rowData[key] = val.GetBoolValue()
			}
		}
		results = append(results, rowData)
	}
	return results, nil
}

// PgInsert inserts a native Go map into PostgreSQL.
func (c *Client) PgInsert(ctx context.Context, table string, data map[string]interface{}) (string, error) {
	packed := make(map[string]*structpb.Value)
	for k, v := range data {
		switch val := v.(type) {
		case string:
			packed[k] = structpb.NewStringValue(val)
		case bool:
			packed[k] = structpb.NewBoolValue(val)
		case float64:
			packed[k] = structpb.NewNumberValue(val)
		case int:
			packed[k] = structpb.NewNumberValue(float64(val))
		default:
			packed[k] = structpb.NewStringValue(fmt.Sprintf("%v", v))
		}
	}

	resp, err := c.dbr.Postgres.InsertData(ctx, &pb.InsertDataRequest{
		Database: c.PgDB,
		Table:    table,
		Data:     packed,
	})
	if err != nil {
		return "", err
	}
	return resp.InsertedId, nil
}

// --- MongoDB API ---

// MongoInsert inserts a native Go map into MongoDB.
func (c *Client) MongoInsert(ctx context.Context, collection string, doc map[string]interface{}) (string, error) {
	s, err := structpb.NewStruct(doc)
	if err != nil {
		return "", fmt.Errorf("packing document: %w", err)
	}

	resp, err := c.dbr.Mongo.InsertDocument(ctx, &pb.InsertDocumentRequest{
		Database:   c.MongoDB,
		Collection: collection,
		Document:   s,
	})
	if err != nil {
		return "", err
	}
	return resp.InsertedId, nil
}

// --- Redis API ---

// RedisSet sets a namespaced key with a TTL in seconds.
func (c *Client) RedisSet(ctx context.Context, key, value string, ttl int32) error {
	nsKey := c.RedisPrefix + key
	_, err := c.dbr.Redis.SetValue(ctx, &pb.SetValueRequest{
		Key:   nsKey,
		Value: value,
		Ttl:   ttl,
	})
	return err
}

// RedisGet fetches a namespaced key. Returns ("", nil) if not found.
func (c *Client) RedisGet(ctx context.Context, key string) (string, error) {
	nsKey := c.RedisPrefix + key
	resp, err := c.dbr.Redis.GetValue(ctx, &pb.GetValueRequest{Key: nsKey})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && strings.Contains(strings.ToLower(st.Message()), "not found") {
			return "", nil
		}
		return "", err
	}
	return resp.Value, nil
}

// Close tears down the underlying gRPC connection.
func (c *Client) Close() error {
	return c.dbr.Close()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
