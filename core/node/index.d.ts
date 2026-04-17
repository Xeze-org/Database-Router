/// <reference types="node" />

interface XezeCoreClientInstance {
  readonly appNamespace: string;
  readonly pgDb: string;
  readonly mongoDb: string;
  readonly redisPrefix: string;

  /** Creates the isolated Postgres database. Safe for startup. */
  initWorkspace(): Promise<void>;

  /** Execute raw SQL returning plain JS objects. */
  pgQuery(query: string): Promise<Record<string, any>[]>;

  /** Insert a plain JS object into PostgreSQL. */
  pgInsert(table: string, data: Record<string, any>): Promise<string>;

  /** Insert a plain JS object into MongoDB. */
  mongoInsert(collection: string, doc: Record<string, any>): Promise<string>;

  /** Set a namespaced key with TTL (default 3600s). */
  redisSet(key: string, value: string, ttl?: number): Promise<void>;

  /** Get a namespaced key. Returns null if not found. */
  redisGet(key: string): Promise<string | null>;

  /** Close the underlying gRPC connections. */
  close(): void;
}

export declare class XezeCoreClient {
  private constructor(appNamespace: string, dbrClient: any);

  /**
   * Factory method — connects via Vault and returns a ready client.
   * @param appNamespace - Unique namespace (e.g., 'xms', 'selfnote')
   */
  static create(appNamespace: string): Promise<XezeCoreClientInstance>;
}
