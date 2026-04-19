const std = @import("std");
const protobuf = @import("protobuf");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    // Create the module representing our SDK
    const module = b.addModule("xeze-dbr", .{
        .root_source_file = b.path("src/client.zig"),
        .target = target,
        .optimize = optimize,
    });

    // Protobuf Generation Step
    const protobuf_dep = b.dependency("protobuf", .{
        .target = target,
        .optimize = optimize,
    });
    const protobuf_mod = protobuf_dep.module("protobuf");

    // We assume the dbrouter.proto file is at ../../proto/dbrouter.proto
    const gen_proto = b.addRunArtifact(protobuf_dep.artifact("protoc-gen-zig"));
    gen_proto.addArg("--zig_out=.");
    gen_proto.addArg("-I../../proto");
    gen_proto.addArg("../../proto/dbrouter.proto");

    module.addImport("protobuf", protobuf_mod);

    // Tests
    const tests = b.addTest(.{
        .root_source_file = b.path("src/client.zig"),
        .target = target,
        .optimize = optimize,
    });
    tests.addImport("protobuf", protobuf_mod);

    const run_tests = b.addRunArtifact(tests);
    const test_step = b.step("test", "Run library tests");
    test_step.dependOn(&run_tests.step);
}
