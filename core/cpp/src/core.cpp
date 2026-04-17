#include "xeze/dbr/core.hpp"
#include <cpr/cpr.h>
#include <nlohmann/json.hpp>
#include <cstdlib>
#include <stdexcept>
#include <iostream>

using json = nlohmann::json;

namespace xeze::dbr {

CoreClient::CoreClient(std::string app_namespace)
    : app_namespace_(std::move(app_namespace)) {
    
    if (app_namespace_.empty()) {
        throw std::invalid_argument("app_namespace is required");
    }

    pg_db_ = app_namespace_ + "_pg";
    mongo_db_ = app_namespace_ + "_mongo";
    redis_prefix_ = app_namespace_ + ":";

    connect_via_vault();
}

std::string CoreClient::get_env(const std::string& key, const std::string& fallback) {
    const char* val = std::getenv(key.c_str());
    if (val == nullptr) return fallback;
    return std::string(val);
}

void CoreClient::connect_via_vault() {
    auto vault_addr = get_env("VAULT_ADDR", "http://127.0.0.1:8200");
    auto vault_token = get_env("VAULT_TOKEN", "dev-root-token");
    auto host = get_env("DB_ROUTER_HOST", "db.0.xeze.org:443");

    cpr::Response r = cpr::Get(
        cpr::Url{vault_addr + "/v1/secret/data/dbrouter/certs"},
        cpr::Header{{"X-Vault-Token", vault_token}}
    );

    if (r.status_code != 200) {
        throw std::runtime_error("Failed to read from Vault. Status: " + std::to_string(r.status_code));
    }

    auto parsed = json::parse(r.text);
    std::string cert_pem = parsed["data"]["data"]["client_cert"];
    std::string key_pem = parsed["data"]["data"]["client_key"];

    if (cert_pem.empty() || key_pem.empty()) {
        throw std::runtime_error("Invalid or missing certificates in Vault response.");
    }

    Options opts;
    opts.host = host;
    opts.cert_data = cert_pem;
    opts.key_data = key_pem;

    dbr_ = std::make_unique<Client>(opts);
}

void CoreClient::init_workspace() {
    dbrouter::CreateDatabaseRequest req;
    req.set_name(pg_db_);
    dbrouter::CreateDatabaseResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = dbr_->postgres->CreateDatabase(&ctx, req, &resp);
    if (!status.ok()) {
        if (status.error_message().find("already exists") != std::string::npos) {
            return; // ignore
        }
        std::cerr << "[WARN] Workspace check failed: " << status.error_message() << std::endl;
        return;
    }
    std::cout << "[OK] Provisioned workspace: " << pg_db_ << std::endl;
}

google::protobuf::Value CoreClient::to_proto_value(const std::any& val) {
    google::protobuf::Value p_val;
    
    if (val.type() == typeid(std::string)) {
        p_val.set_string_value(std::any_cast<std::string>(val));
    } else if (val.type() == typeid(const char*)) {
        p_val.set_string_value(std::any_cast<const char*>(val));
    } else if (val.type() == typeid(int)) {
        p_val.set_number_value(std::any_cast<int>(val));
    } else if (val.type() == typeid(double)) {
        p_val.set_number_value(std::any_cast<double>(val));
    } else if (val.type() == typeid(bool)) {
        p_val.set_bool_value(std::any_cast<bool>(val));
    } else {
        // Fallback or explicit null mapping can be complex in C++.
        p_val.set_null_value(google::protobuf::NULL_VALUE);
    }
    return p_val;
}

std::string CoreClient::pg_insert(const std::string& table, const std::map<std::string, std::any>& data) {
    dbrouter::InsertDataRequest req;
    req.set_database(pg_db_);
    req.set_table(table);

    auto mutable_data = req.mutable_data();
    for (const auto& [k, v] : data) {
        (*mutable_data)[k] = to_proto_value(v);
    }

    dbrouter::InsertDataResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = dbr_->postgres->InsertData(&ctx, req, &resp);
    if (!status.ok()) {
        throw std::runtime_error("PG Insert failed: " + status.error_message());
    }
    return resp.inserted_id();
}

std::string CoreClient::mongo_insert(const std::string& collection, const std::map<std::string, std::any>& doc) {
    dbrouter::InsertDocumentRequest req;
    req.set_database(mongo_db_);
    req.set_collection(collection);

    auto struct_doc = req.mutable_document();
    auto fields = struct_doc->mutable_fields();
    for (const auto& [k, v] : doc) {
        (*fields)[k] = to_proto_value(v);
    }

    dbrouter::InsertDocumentResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = dbr_->mongo->InsertDocument(&ctx, req, &resp);
    if (!status.ok()) {
        throw std::runtime_error("Mongo Insert failed: " + status.error_message());
    }
    return resp.inserted_id();
}

void CoreClient::redis_set(const std::string& key, const std::string& value, int ttl) {
    dbrouter::SetValueRequest req;
    req.set_key(redis_prefix_ + key);
    req.set_value(value);
    req.set_ttl(ttl);

    dbrouter::SetValueResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = dbr_->redis->SetValue(&ctx, req, &resp);
    if (!status.ok()) {
        throw std::runtime_error("Redis Set failed: " + status.error_message());
    }
}

std::string CoreClient::redis_get(const std::string& key) {
    dbrouter::GetValueRequest req;
    req.set_key(redis_prefix_ + key);

    dbrouter::GetValueResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = dbr_->redis->GetValue(&ctx, req, &resp);
    if (!status.ok()) {
        if (status.error_message().find("not found") != std::string::npos) {
            return "";
        }
        throw std::runtime_error("Redis Get failed: " + status.error_message());
    }
    return resp.value();
}

} // namespace xeze::dbr
