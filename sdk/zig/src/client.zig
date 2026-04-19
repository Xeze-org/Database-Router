const std = @import("std");
const string = []const u8;
// In a full implementation, we would import the generated pb files:
// const pb = @import("pb/dbrouter.pb.zig");

/// Configuration options for the Xeze Database Router client
pub const ConnectOptions = struct {
    host: string,
    insecure: bool = false,
    
    cert_path: ?string = null,
    key_path: ?string = null,
    ca_path: ?string = null,

    cert_data: ?string = null,
    key_data: ?string = null,
    ca_data: ?string = null,
};

/// The unified gRPC Client wrapping all backend services.
pub const Client = struct {
    allocator: std.mem.Allocator,
    host: string,

    // NOTE: This represents the high-level API wrapper logic for connecting via C gRPC Core
    // or a pure Zig gRPC transport. Given the experimental state of Zig gRPC, this struct
    // acts primarily as the abstraction boundary mimicking Go and Python SDKs.

    pub fn init(allocator: std.mem.Allocator, opts: ConnectOptions) !*Client {
        if (!opts.insecure and opts.cert_data == null and opts.cert_path == null) {
            return error.MissingCredentials;
        }

        const client = try allocator.create(Client);
        client.* = .{
            .allocator = allocator,
            .host = try allocator.dupe(u8, opts.host),
        };
        
        // Transport setup goes here
        return client;
    }

    pub fn deinit(self: *Client) void {
        self.allocator.free(self.host);
        self.allocator.destroy(self);
    }
};

test "client initialization" {
    const testing = std.testing;
    
    var client = try Client.init(testing.allocator, .{
        .host = "localhost:50051",
        .insecure = true,
    });
    defer client.deinit();

    try testing.expectEqualStrings("localhost:50051", client.host);
}
