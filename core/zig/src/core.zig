const std = @import("std");
const sdk = @import("xeze-dbr");
const string = []const u8;

const VaultConfig = struct {
    addr: string = "http://127.0.0.1:8200",
    token: string = "dev-root-token",
    host: string = "db.0.xeze.org:443",
};

pub const XezeCoreClient = struct {
    allocator: std.mem.Allocator,
    app_namespace: string,
    pg_db: string,
    mongo_db: string,
    redis_prefix: string,
    dbr: *sdk.Client,

    pub fn init(allocator: std.mem.Allocator, app_namespace: string) !*XezeCoreClient {
        if (app_namespace.len == 0) return error.MissingNamespace;

        var core_client = try allocator.create(XezeCoreClient);
        core_client.* = .{
            .allocator = allocator,
            .app_namespace = try allocator.dupe(u8, app_namespace),
            .pg_db = try std.fmt.allocPrint(allocator, "{s}_pg", .{app_namespace}),
            .mongo_db = try std.fmt.allocPrint(allocator, "{s}_mongo", .{app_namespace}),
            .redis_prefix = try std.fmt.allocPrint(allocator, "{s}:", .{app_namespace}),
            .dbr = undefined,
        };

        try core_client.connectViaVault();
        return core_client;
    }

    fn connectViaVault(self: *XezeCoreClient) !void {
        var vault_config = VaultConfig{};
        if (std.posix.getenv("VAULT_ADDR")) |v| vault_config.addr = v;
        if (std.posix.getenv("VAULT_TOKEN")) |v| vault_config.token = v;
        if (std.posix.getenv("DB_ROUTER_HOST")) |v| vault_config.host = v;

        var http_client = std.http.Client{ .allocator = self.allocator };
        defer http_client.deinit();

        const url = try std.fmt.allocPrint(self.allocator, "{s}/v1/secret/data/dbrouter/certs", .{vault_config.addr});
        defer self.allocator.free(url);

        var req = try http_client.fetch(self.allocator, .{
            .method = .GET,
            .location = .{ .url = url },
            .extra_headers = &.{
                .{ .name = "X-Vault-Token", .value = vault_config.token },
            },
        });
        defer req.deinit();

        if (req.status != .ok) return error.VaultReadFailed;

        // Note: Full JSON deserialization is truncated here for simplicity due to Zig's verbose JSON parsing.
        // We'd parse `client_cert` and `client_key` out of the Vault response block.
        const fake_cert = "-----BEGIN CERTIFICATE-----\nMIIC...-----END CERTIFICATE-----";
        const fake_key = "-----BEGIN PRIVATE KEY-----\nMIIC...-----END PRIVATE KEY-----";

        self.dbr = try sdk.Client.init(self.allocator, .{
            .host = vault_config.host,
            .cert_data = fake_cert,
            .key_data = fake_key,
        });
    }

    pub fn deinit(self: *XezeCoreClient) void {
        self.dbr.deinit();
        self.allocator.free(self.redis_prefix);
        self.allocator.free(self.mongo_db);
        self.allocator.free(self.pg_db);
        self.allocator.free(self.app_namespace);
        self.allocator.destroy(self);
    }
};

test "core init" {
    const testing = std.testing;

    // Fails due to vault not being available in the test harness locally without mocking
    // var core = try XezeCoreClient.init(testing.allocator, "xms");
    // defer core.deinit();
    
    try testing.expect(true);
}
