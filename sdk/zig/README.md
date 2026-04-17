# xeze-dbr

Zig gRPC client for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router) — a unified interface for **PostgreSQL**, **MongoDB**, and **Redis** over **mTLS**.

> **⚠️ WARNING**: Zig's gRPC ecosystem is currently highly experimental. This library connects to Protobuf bindings and represents an emerging abstraction.

## Install

Add the following to your `build.zig.zon`:

```zig
.{
    .dependencies = .{
        .@"xeze-dbr" = .{
            .url = "https://code.xeze.org/xeze/Database-Router/archive/main.tar.gz",
            // Make sure to append the hash!
        },
    },
}
```

Then in your `build.zig`:

```zig
const sdk_dep = b.dependency("xeze-dbr", .{
    .target = target,
    .optimize = optimize,
});
exe.root_module.addImport("xeze-dbr", sdk_dep.module("xeze-dbr"));
```

## Quick Start

```zig
const std = @import("std");
const sdk = @import("xeze-dbr");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();

    // 1. Configure and connect
    var client = try sdk.Client.init(gpa.allocator(), .{
        .host = "db.0.xeze.org:443",
        .cert_path = "client.crt",
        .key_path = "client.key",
    });
    defer client.deinit();

    // The client is securely connected and ready!
}
```

## License

Apache 2.0
