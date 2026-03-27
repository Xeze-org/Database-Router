/// <reference types="node" />

interface ConnectOptions {
  /** gRPC target, e.g. "db.0.xeze.org:443" */
  host: string;
  /** Path to client certificate file (.crt) */
  certPath?: string;
  /** Path to client key file (.key) */
  keyPath?: string;
  /** Optional path to CA certificate */
  caPath?: string;
  /** Raw client certificate bytes (for Vault / secret managers) */
  certData?: Buffer;
  /** Raw client key bytes (for Vault / secret managers) */
  keyData?: Buffer;
  /** Raw CA certificate bytes */
  caData?: Buffer;
  /** Use plaintext channel (for local dev only) */
  insecure?: boolean;
}

interface HealthStub {
  Check(request?: {}): Promise<any>;
  CheckPostgres(request?: {}): Promise<any>;
  CheckMongo(request?: {}): Promise<any>;
  CheckRedis(request?: {}): Promise<any>;
  close(): void;
}

interface PostgresStub {
  ListDatabases(request?: {}): Promise<any>;
  CreateDatabase(request: { name: string }): Promise<any>;
  ListTables(request: { database: string }): Promise<any>;
  ExecuteQuery(request: { query: string; database: string }): Promise<any>;
  SelectData(request: { database: string; table: string; limit?: number }): Promise<any>;
  InsertData(request: { database: string; table: string; data: Record<string, any> }): Promise<any>;
  UpdateData(request: { database: string; table: string; id: string; data: Record<string, any> }): Promise<any>;
  DeleteData(request: { database: string; table: string; id: string }): Promise<any>;
  close(): void;
}

interface MongoStub {
  ListDatabases(request?: {}): Promise<any>;
  ListCollections(request: { database: string }): Promise<any>;
  InsertDocument(request: { database: string; collection: string; document: any }): Promise<any>;
  FindDocuments(request: { database: string; collection: string }): Promise<any>;
  UpdateDocument(request: { database: string; collection: string; id: string; update: any }): Promise<any>;
  DeleteDocument(request: { database: string; collection: string; id: string }): Promise<any>;
  close(): void;
}

interface RedisStub {
  ListKeys(request: { pattern: string }): Promise<any>;
  SetValue(request: { key: string; value: string; ttl?: number }): Promise<any>;
  GetValue(request: { key: string }): Promise<any>;
  DeleteKey(request: { key: string }): Promise<any>;
  Info(request?: {}): Promise<any>;
  close(): void;
}

interface DbRouterClient {
  health: HealthStub;
  postgres: PostgresStub;
  mongo: MongoStub;
  redis: RedisStub;
  proto: any;
  close(): void;
}

export function connect(options: ConnectOptions): DbRouterClient;
