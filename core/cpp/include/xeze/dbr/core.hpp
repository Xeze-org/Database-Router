#pragma once

#include "xeze/dbr/client.hpp"
#include <string>
#include <memory>
#include <map>
#include <any>

namespace xeze::dbr {

class CoreClient {
public:
    explicit CoreClient(std::string app_namespace);
    ~CoreClient() = default;

    void init_workspace();

    // PostgreSQL API
    std::string pg_insert(const std::string& table, const std::map<std::string, std::any>& data);

    // MongoDB API
    std::string mongo_insert(const std::string& collection, const std::map<std::string, std::any>& doc);

    // Redis API
    void redis_set(const std::string& key, const std::string& value, int ttl);
    std::string redis_get(const std::string& key);

private:
    void connect_via_vault();
    std::string get_env(const std::string& key, const std::string& fallback);
    
    // Abstracting std::any values into protobuf dynamic Value structures
    google::protobuf::Value to_proto_value(const std::any& val);

    std::string app_namespace_;
    std::string pg_db_;
    std::string mongo_db_;
    std::string redis_prefix_;

    std::unique_ptr<Client> dbr_;
};

} // namespace xeze::dbr
