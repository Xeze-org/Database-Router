#include "xeze/dbr/client.hpp"
#include <fstream>
#include <sstream>
#include <stdexcept>

namespace xeze::dbr {

namespace {
std::string load_file(const std::string& path) {
    std::ifstream file(path);
    if (!file.is_open()) {
        throw std::runtime_error("Failed to open file: " + path);
    }
    std::stringstream buffer;
    buffer << file.rdbuf();
    return buffer.str();
}
} // namespace

Client::Client(const Options& opts) {
    std::shared_ptr<grpc::ChannelCredentials> creds;

    if (opts.insecure) {
        creds = grpc::InsecureChannelCredentials();
    } else {
        grpc::SslCredentialsOptions ssl_opts;

        if (!opts.cert_data.empty() && !opts.key_data.empty()) {
            ssl_opts.pem_cert_chain = opts.cert_data;
            ssl_opts.pem_private_key = opts.key_data;
        } else if (!opts.cert_file.empty() && !opts.key_file.empty()) {
            ssl_opts.pem_cert_chain = load_file(opts.cert_file);
            ssl_opts.pem_private_key = load_file(opts.key_file);
        } else {
            throw std::invalid_argument("Either cert_data/key_data or cert_file/key_file must be provided for secure connections.");
        }

        if (!opts.ca_data.empty()) {
            ssl_opts.pem_root_certs = opts.ca_data;
        } else if (!opts.ca_file.empty()) {
            ssl_opts.pem_root_certs = load_file(opts.ca_file);
        }

        creds = grpc::SslCredentials(ssl_opts);
    }

    // Set up channel arguments for modern settings
    grpc::ChannelArguments args;
    args.SetMaxReceiveMessageSize(100 * 1024 * 1024); // 100MB
    args.SetMaxSendMessageSize(100 * 1024 * 1024);

    channel_ = grpc::CreateCustomChannel(opts.host, creds, args);

    health = dbrouter::HealthService::NewStub(channel_);
    postgres = dbrouter::PostgresService::NewStub(channel_);
    mongo = dbrouter::MongoService::NewStub(channel_);
    redis = dbrouter::RedisService::NewStub(channel_);
}

} // namespace xeze::dbr
