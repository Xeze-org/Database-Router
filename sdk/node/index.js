/**
 * @xeze/dbr — Node.js gRPC client for the Xeze Database Router.
 *
 * @example
 * const { connect } = require("@xeze/dbr");
 *
 * // File-based certs
 * const client = connect({ host: "db.0.xeze.org:443", certPath: "client.crt", keyPath: "client.key" });
 *
 * // Vault / raw bytes
 * const client = connect({ host: "db.0.xeze.org:443", certData: Buffer.from(cert), keyData: Buffer.from(key) });
 */

const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const fs = require("fs");
const path = require("path");

const PROTO_PATH = path.join(__dirname, "proto", "dbrouter.proto");

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: false,
  longs: Number,
  enums: String,
  defaults: true,
  oneofs: true,
});

const proto = grpc.loadPackageDefinition(packageDefinition).dbrouter;

/**
 * Connect to a Xeze Database Router instance.
 *
 * @param {object} options
 * @param {string} options.host - gRPC target, e.g. "db.0.xeze.org:443"
 * @param {string} [options.certPath] - Path to client certificate file (.crt)
 * @param {string} [options.keyPath] - Path to client key file (.key)
 * @param {string} [options.caPath] - Optional path to CA certificate
 * @param {Buffer} [options.certData] - Raw client certificate bytes (for Vault)
 * @param {Buffer} [options.keyData] - Raw client key bytes (for Vault)
 * @param {Buffer} [options.caData] - Raw CA certificate bytes
 * @param {boolean} [options.insecure=false] - Use plaintext (for local dev)
 * @returns {DbRouterClient}
 */
function connect(options) {
  const { host, certPath, keyPath, caPath, certData, keyData, caData, insecure = false } = options;

  let channel;

  if (insecure) {
    channel = grpc.credentials.createInsecure();
  } else {
    const rootCa = caData || (caPath ? fs.readFileSync(caPath) : null);
    const clientKey = keyData || (keyPath ? fs.readFileSync(keyPath) : null);
    const clientCert = certData || (certPath ? fs.readFileSync(certPath) : null);

    channel = grpc.credentials.createSsl(rootCa, clientKey, clientCert);
  }

  const health = new proto.HealthService(host, channel);
  const postgres = new proto.PostgresService(host, channel);
  const mongo = new proto.MongoService(host, channel);
  const redis = new proto.RedisService(host, channel);

  return {
    health: promisifyStub(health),
    postgres: promisifyStub(postgres),
    mongo: promisifyStub(mongo),
    redis: promisifyStub(redis),

    /** Access raw proto definitions for building requests */
    proto,

    /** Close all underlying connections */
    close() {
      health.close();
      postgres.close();
      mongo.close();
      redis.close();
    },
  };
}

/**
 * Wraps a gRPC stub so every method returns a Promise instead of using callbacks.
 */
function promisifyStub(stub) {
  const wrapped = {};
  const serviceMethods = Object.keys(Object.getPrototypeOf(stub)).filter(
    (k) => k !== "constructor" && typeof stub[k] === "function"
  );

  for (const method of serviceMethods) {
    wrapped[method] = (request = {}) => {
      return new Promise((resolve, reject) => {
        stub[method](request, (err, response) => {
          if (err) reject(err);
          else resolve(response);
        });
      });
    };
  }

  // Keep the close method accessible
  wrapped.close = () => stub.close();
  return wrapped;
}

module.exports = { connect };
