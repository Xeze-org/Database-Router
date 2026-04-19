#pragma once

#include <string>
#include <memory>
#include <grpcpp/grpcpp.h>
#include "dbrouter.grpc.pb.h"

namespace xeze::dbr {

struct Options {
    std::string host;
    bool insecure = false;

    std::string cert_file;
    std::string key_file;
    std::string ca_file;

    std::string cert_data;
    std::string key_data;
    std::string ca_data;
};

class Client {
public:
    explicit Client(const Options& opts);
    ~Client() = default;

    // Health Stub
    std::unique_ptr<dbrouter::HealthService::Stub> health;

    // Database Stubs
    std::unique_ptr<dbrouter::PostgresService::Stub> postgres;
    std::unique_ptr<dbrouter::MongoService::Stub> mongo;
    std::unique_ptr<dbrouter::RedisService::Stub> redis;

private:
    std::shared_ptr<grpc::Channel> channel_;
};

} // namespace xeze::dbr
