cmake_minimum_required(VERSION 3.24)
project(xeze-dbr CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED True)

# Set up FetchContent for gRPC and Protobuf
include(FetchContent)

# Fetch gRPC (this will automatically fetch Protobuf as a dependency)
FetchContent_Declare(
    gRPC
    GIT_REPOSITORY https://github.com/grpc/grpc
    GIT_TAG        v1.62.1
)
set(FETCHCONTENT_QUIET OFF)
set(gRPC_INSTALL OFF CACHE BOOL "Disable installation of gRPC")
set(gRPC_BUILD_TESTS OFF CACHE BOOL "Disable building gRPC tests")
FetchContent_MakeAvailable(gRPC)

# Locate the protobuf compiler and grpc_cpp_plugin
set(_PROTOBUF_PROTOC $<TARGET_FILE:protoc>)
set(_GRPC_CPP_PLUGIN_EXECUTABLE $<TARGET_FILE:grpc_cpp_plugin>)

# Paths
set(PROTO_FILE "${CMAKE_SOURCE_DIR}/../../proto/dbrouter.proto")
set(PROTO_OUT_DIR "${CMAKE_BINARY_DIR}/generated")
file(MAKE_DIRECTORY ${PROTO_OUT_DIR})

# Generate stubs
add_custom_command(
    OUTPUT "${PROTO_OUT_DIR}/dbrouter.pb.h"
           "${PROTO_OUT_DIR}/dbrouter.pb.cc"
           "${PROTO_OUT_DIR}/dbrouter.grpc.pb.h"
           "${PROTO_OUT_DIR}/dbrouter.grpc.pb.cc"
    COMMAND ${_PROTOBUF_PROTOC}
    ARGS --proto_path="${CMAKE_SOURCE_DIR}/../../proto"
         --cpp_out="${PROTO_OUT_DIR}"
         --grpc_out="${PROTO_OUT_DIR}"
         --plugin=protoc-gen-grpc="${_GRPC_CPP_PLUGIN_EXECUTABLE}"
         "${PROTO_FILE}"
    DEPENDS "${PROTO_FILE}" protoc grpc_cpp_plugin
)

# Create library
add_library(xeze_dbr
    src/client.cpp
    "${PROTO_OUT_DIR}/dbrouter.pb.cc"
    "${PROTO_OUT_DIR}/dbrouter.grpc.pb.cc"
)

target_include_directories(xeze_dbr PUBLIC
    $<BUILD_INTERFACE:${CMAKE_SOURCE_DIR}/include>
    $<BUILD_INTERFACE:${PROTO_OUT_DIR}>
    $<INSTALL_INTERFACE:include>
)

target_link_libraries(xeze_dbr PUBLIC
    gRPC::grpc++
    protobuf::libprotobuf
)
