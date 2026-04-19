const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    // Determine path to our SDK
    const sdk_dep = b.dependency("xeze-dbr", .{
        .target = target,
        .optimize = optimize,
    });
    const sdk_mod = sdk_dep.module("xeze-dbr");

    const module = b.addModule("xeze-dbr-core", .{
        .root_source_file = b.path("src/core.zig"),
        .target = target,
        .optimize = optimize,
    });
    
    module.addImport("xeze-dbr", sdk_mod);

    // Tests
    const tests = b.addTest(.{
        .root_source_file = b.path("src/core.zig"),
        .target = target,
        .optimize = optimize,
    });
    tests.addImport("xeze-dbr", sdk_mod);

    const run_tests = b.addRunArtifact(tests);
    const test_step = b.step("test", "Run library tests");
    test_step.dependOn(&run_tests.step);
}
