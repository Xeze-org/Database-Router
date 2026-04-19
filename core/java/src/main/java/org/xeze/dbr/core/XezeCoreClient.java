package org.xeze.dbr.core;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import dbrouter.Dbrouter;
import org.xeze.dbr.Options;
import org.xeze.dbr.XezeDbrClient;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class XezeCoreClient implements AutoCloseable {

    private final String appNamespace;
    private final String pgDb;
    private final String mongoDb;
    private final String redisPrefix;

    private XezeDbrClient dbr;
    private final ObjectMapper mapper = new ObjectMapper();

    public XezeCoreClient(String appNamespace) throws Exception {
        if (appNamespace == null || appNamespace.isEmpty()) {
            throw new IllegalArgumentException("app_namespace is required");
        }
        this.appNamespace = appNamespace;
        this.pgDb = appNamespace + "_pg";
        this.mongoDb = appNamespace + "_mongo";
        this.redisPrefix = appNamespace + ":";

        connectViaVault();
    }

    private void connectViaVault() throws Exception {
        String vaultAddr = getEnv("VAULT_ADDR", "http://127.0.0.1:8200");
        String vaultToken = getEnv("VAULT_TOKEN", "dev-root-token");
        String host = getEnv("DB_ROUTER_HOST", "db.0.xeze.org:443");

        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(vaultAddr + "/v1/secret/data/dbrouter/certs"))
                .header("X-Vault-Token", vaultToken)
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() != 200) {
            throw new RuntimeException("Failed to read from Vault. Status: " + response.statusCode() + " Body: " + response.body());
        }

        JsonNode rootNode = mapper.readTree(response.body());
        JsonNode dataNode = rootNode.path("data").path("data");

        String certPem = dataNode.path("client_cert").asText();
        String keyPem = dataNode.path("client_key").asText();

        if (certPem.isEmpty() || keyPem.isEmpty()) {
            throw new RuntimeException("Invalid or missing certificates in Vault response.");
        }

        Options opts = new Options(host);
        opts.certData = certPem.getBytes();
        opts.keyData = keyPem.getBytes();
        // Caddy's certificates are typically trusted if signed by a recognized CA,
        // or we usually assume standard system trust store. For mTLS, client certs are crucial.

        this.dbr = XezeDbrClient.connect(opts);
    }

    public void initWorkspace() {
        Dbrouter.CreateDatabaseRequest req = Dbrouter.CreateDatabaseRequest.newBuilder().setName(pgDb).build();
        try {
            dbr.postgres.createDatabase(req);
            System.out.println("[OK] Provisioned workspace: " + pgDb);
        } catch (Exception e) {
            if (e.getMessage() != null && e.getMessage().toLowerCase().contains("already exists")) {
                // Ignore
            } else {
                System.out.println("[WARN] Workspace check failed: " + e.getMessage());
            }
        }
    }

    // --- PostgreSQL API ---

    public List<Map<String, Object>> pgQuery(String query) {
        Dbrouter.ExecuteQueryRequest req = Dbrouter.ExecuteQueryRequest.newBuilder()
                .setDatabase(pgDb)
                .setQuery(query)
                .build();

        Dbrouter.ExecuteQueryResponse resp = dbr.postgres.executeQuery(req);

        List<Map<String, Object>> results = new ArrayList<>();
        for (Dbrouter.QueryResultRow row : resp.getRowsList()) {
            Map<String, Object> map = new HashMap<>();
            for (Map.Entry<String, Value> entry : row.getFieldsMap().entrySet()) {
                map.put(entry.getKey(), getValueObject(entry.getValue()));
            }
            results.add(map);
        }
        return results;
    }

    public String pgInsert(String table, Map<String, Object> data) {
        Map<String, Value> protoMap = new HashMap<>();
        for (Map.Entry<String, Object> entry : data.entrySet()) {
            protoMap.put(entry.getKey(), toProtoValue(entry.getValue()));
        }

        Dbrouter.InsertDataRequest req = Dbrouter.InsertDataRequest.newBuilder()
                .setDatabase(pgDb)
                .setTable(table)
                .putAllData(protoMap)
                .build();

        Dbrouter.InsertDataResponse resp = dbr.postgres.insertData(req);
        return resp.getInsertedId();
    }

    // --- MongoDB API ---

    public String mongoInsert(String collection, Map<String, Object> doc) {
        Map<String, Value> protoMap = new HashMap<>();
        for (Map.Entry<String, Object> entry : doc.entrySet()) {
            protoMap.put(entry.getKey(), toProtoValue(entry.getValue()));
        }

        Struct struct = Struct.newBuilder().putAllFields(protoMap).build();

        Dbrouter.InsertDocumentRequest req = Dbrouter.InsertDocumentRequest.newBuilder()
                .setDatabase(mongoDb)
                .setCollection(collection)
                .setDocument(struct)
                .build();

        Dbrouter.InsertDocumentResponse resp = dbr.mongo.insertDocument(req);
        return resp.getInsertedId();
    }

    // --- Redis API ---

    public void redisSet(String key, String value, int ttlSeconds) {
        String nsKey = redisPrefix + key;
        Dbrouter.SetValueRequest req = Dbrouter.SetValueRequest.newBuilder()
                .setKey(nsKey)
                .setValue(value)
                .setTtl(ttlSeconds)
                .build();
        dbr.redis.setValue(req);
    }

    public String redisGet(String key) {
        String nsKey = redisPrefix + key;
        Dbrouter.GetValueRequest req = Dbrouter.GetValueRequest.newBuilder()
                .setKey(nsKey)
                .build();
        try {
            Dbrouter.GetValueResponse resp = dbr.redis.getValue(req);
            return resp.getValue();
        } catch (Exception e) {
            if (e.getMessage() != null && e.getMessage().toLowerCase().contains("not found")) {
                return null;
            }
            throw e;
        }
    }

    @Override
    public void close() throws Exception {
        if (dbr != null) {
            dbr.close();
        }
    }

    // Helpers
    private String getEnv(String key, String fallback) {
        String val = System.getenv(key);
        return (val != null && !val.isEmpty()) ? val : fallback;
    }

    private Value toProtoValue(Object obj) {
        if (obj == null) return Value.newBuilder().setNullValue(com.google.protobuf.NullValue.NULL_VALUE).build();
        if (obj instanceof String) return Value.newBuilder().setStringValue((String) obj).build();
        if (obj instanceof Number) return Value.newBuilder().setNumberValue(((Number) obj).doubleValue()).build();
        if (obj instanceof Boolean) return Value.newBuilder().setBoolValue((Boolean) obj).build();

        // Fallback to string representation
        return Value.newBuilder().setStringValue(obj.toString()).build();
    }

    private Object getValueObject(Value v) {
        switch (v.getKindCase()) {
            case STRING_VALUE: return v.getStringValue();
            case NUMBER_VALUE: return v.getNumberValue();
            case BOOL_VALUE: return v.getBoolValue();
            default: return v.toString();
        }
    }
}
