/**
 * @xeze/dbr-core — High-level Vault-integrated database client for Node.js.
 *
 * Enforces database-per-service isolation via app_namespace.
 * Handles Vault mTLS certificate loading automatically.
 *
 * @example
 * const { XezeCoreClient } = require("@xeze/dbr-core");
 *
 * const db = await XezeCoreClient.create("xms");
 * await db.initWorkspace();
 *
 * await db.pgInsert("students", { name: "Ayush", grade: "A" });
 * await db.mongoInsert("logs", { action: "student_added" });
 * await db.redisSet("cache:student", "Ayush", 300);
 * const val = await db.redisGet("cache:student");
 */

const { connect } = require("@xeze/dbr");
const vault = require("node-vault");

class XezeCoreClient {
  /**
   * @param {string} appNamespace
   * @param {object} dbrClient — connected @xeze/dbr client
   */
  constructor(appNamespace, dbrClient) {
    this.appNamespace = appNamespace;
    this.pgDb = `${appNamespace}_pg`;
    this.mongoDb = `${appNamespace}_mongo`;
    this.redisPrefix = `${appNamespace}:`;
    this._client = dbrClient;
  }

  /**
   * Factory method — connects via Vault and returns a ready client.
   * @param {string} appNamespace
   * @returns {Promise<XezeCoreClient>}
   */
  static async create(appNamespace) {
    if (!appNamespace || typeof appNamespace !== "string") {
      throw new Error("A strict appNamespace (e.g., 'xms', 'selfnote') is required.");
    }

    const vaultAddr = process.env.VAULT_ADDR || "http://127.0.0.1:8200";
    const vaultToken = process.env.VAULT_TOKEN || "dev-root-token";
    const host = process.env.DB_ROUTER_HOST || "db.0.xeze.org:443";

    // Fetch certs from Vault KV v2
    const vc = vault({ apiVersion: "v1", endpoint: vaultAddr, token: vaultToken });
    const secret = await vc.read("secret/data/dbrouter/certs");
    const certData = Buffer.from(secret.data.data.client_cert);
    const keyData = Buffer.from(secret.data.data.client_key);

    const client = connect({ host, certData, keyData });
    return new XezeCoreClient(appNamespace, client);
  }

  /**
   * Creates the isolated Postgres database. Safe for startup.
   */
  async initWorkspace() {
    try {
      await this._client.postgres.CreateDatabase({ name: this.pgDb });
      console.log(`[OK] Provisioned workspace: ${this.pgDb}`);
    } catch (err) {
      if (err.details && err.details.toLowerCase().includes("already exists")) {
        // Expected
      } else {
        console.log(`[WARN] Workspace check failed: ${err.details || err.message}`);
      }
    }
  }

  // --- PostgreSQL API ---

  /**
   * Execute raw SQL and return results as plain JS objects.
   * @param {string} query
   * @returns {Promise<object[]>}
   */
  async pgQuery(query) {
    const resp = await this._client.postgres.ExecuteQuery({
      database: this.pgDb,
      query,
    });
    return (resp.rows || []).map((row) => {
      const obj = {};
      for (const [key, val] of Object.entries(row.fields || {})) {
        if (val.stringValue !== undefined) obj[key] = val.stringValue;
        else if (val.numberValue !== undefined) obj[key] = val.numberValue;
        else if (val.boolValue !== undefined) obj[key] = val.boolValue;
      }
      return obj;
    });
  }

  /**
   * Insert a plain JS object into PostgreSQL.
   * @param {string} table
   * @param {object} data
   * @returns {Promise<string>} inserted_id
   */
  async pgInsert(table, data) {
    const packed = {};
    for (const [k, v] of Object.entries(data)) {
      if (typeof v === "string") packed[k] = { stringValue: v };
      else if (typeof v === "boolean") packed[k] = { boolValue: v };
      else if (typeof v === "number") packed[k] = { numberValue: v };
      else packed[k] = { stringValue: String(v) };
    }
    const resp = await this._client.postgres.InsertData({
      database: this.pgDb,
      table,
      data: packed,
    });
    return resp.insertedId;
  }

  // --- MongoDB API ---

  /**
   * Insert a plain JS object into MongoDB.
   * @param {string} collection
   * @param {object} doc
   * @returns {Promise<string>} inserted_id
   */
  async mongoInsert(collection, doc) {
    const resp = await this._client.mongo.InsertDocument({
      database: this.mongoDb,
      collection,
      document: doc,
    });
    return resp.insertedId;
  }

  // --- Redis API ---

  /**
   * Set a namespaced key with TTL.
   * @param {string} key
   * @param {string} value
   * @param {number} [ttl=3600]
   */
  async redisSet(key, value, ttl = 3600) {
    const nsKey = `${this.redisPrefix}${key}`;
    await this._client.redis.SetValue({ key: nsKey, value: String(value), ttl });
  }

  /**
   * Get a namespaced key. Returns null if not found.
   * @param {string} key
   * @returns {Promise<string|null>}
   */
  async redisGet(key) {
    const nsKey = `${this.redisPrefix}${key}`;
    try {
      const resp = await this._client.redis.GetValue({ key: nsKey });
      return resp.value;
    } catch (err) {
      if (err.details && err.details.toLowerCase().includes("not found")) {
        return null;
      }
      throw err;
    }
  }

  /** Close the underlying gRPC connections. */
  close() {
    this._client.close();
  }
}

module.exports = { XezeCoreClient };
