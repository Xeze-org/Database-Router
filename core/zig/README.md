# xeze-dbr-core

Official unified database wrapper for the Xeze infrastructure. Provides a single, heavily-abstracted client for **PostgreSQL**, **MongoDB**, and **Redis** over mTLS via HashiCorp Vault.

> **⚠️ WARNING**: Zig's ecosystem is experimental. Use caution when running in production.

## Installation

Add the following to your `build.zig.zon`:

```zig
.{
    .dependencies = .{
        .@"xeze-dbr-core" = .{
            .url = "https://code.xeze.org/xeze/Database-Router/archive/main.tar.gz",
            // Make sure to append the hash!
        },
    },
}
```

And in `build.zig`:

```zig
const core_dep = b.dependency("xeze-dbr-core", .{
    .target = target,
    .optimize = optimize,
});
exe.root_module.addImport("xeze-dbr-core", core_dep.module("xeze-dbr-core"));
```

## Quick Start

```zig
const std = @import("std");
const xezecore = @import("xeze-dbr-core");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();

    // Connects via HashiCorp Vault automatically using std.http.Client
    var db = try xezecore.XezeCoreClient.init(gpa.allocator(), "xms");
    defer db.deinit();
    
    std.debug.print("Connected to isolated workspace: {s}\n", .{db.pg_db});
}
```

## Environment Variables

| Variable           | Default                    | Description                |
| ------------------ | -------------------------- | -------------------------- |
| `VAULT_ADDR`       | `http://127.0.0.1:8200`   | HashiCorp Vault address    |
| `VAULT_TOKEN`      | `dev-root-token`           | Vault authentication token |
| `DB_ROUTER_HOST`   | `db.0.xeze.org:443`       | Database Router gRPC host  |

## License

Apache 2.0
