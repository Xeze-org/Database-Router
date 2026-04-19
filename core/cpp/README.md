# xeze-dbr-core

Official unified database wrapper for the Xeze infrastructure. Provides a single, heavily-abstracted client for **PostgreSQL**, **MongoDB**, and **Redis** over mTLS via HashiCorp Vault utilizing C++23.

## Installation

```cmake
# In your CMakeLists.txt
include(FetchContent)

FetchContent_Declare(xeze_core
    GIT_REPOSITORY https://code.xeze.org/xeze/Database-Router.git
    GIT_TAG main
    SOURCE_SUBDIR core/cpp
)
FetchContent_MakeAvailable(xeze_core)

# Target linking
target_link_libraries(your_target PUBLIC xeze_dbr_core)
```

## Quick Start

```cpp
#include "xeze/dbr/core.hpp"
#include <iostream>

int main() {
    // Automatically dials HashiCorp Vault & instantiates mTLS connectivity underneath!
    xeze::dbr::CoreClient db("xms");

    db.init_workspace();

    // Redis
    db.redis_set("cache:student:latest", "Ayush", 300);
    std::cout << "Cached: " << db.redis_get("cache:student:latest") << std::endl;

    // PostgreSQL (utilizing standard std::map<std::string, std::any> properties)
    std::string id = db.pg_insert("students", {
        {"name", std::string("Ayush")}, 
        {"grade", std::string("A")}
    });
    
    return 0;
}
```

## Architecture Requirements
- CMake 3.24+
- A modern compiler supporting `-std=c++23` (GCC 13+, Clang 16+, MSVC v143+)
- Automatic Dependencies included via CMake:
    - [libcurl wrappers (libcpr)](https://github.com/libcpr/cpr)
    - [nlohmann_json](https://github.com/nlohmann/json)
    - Google gRPC

## Environment Variables

| Variable           | Default                    | Description                |
| ------------------ | -------------------------- | -------------------------- |
| `VAULT_ADDR`       | `http://127.0.0.1:8200`   | HashiCorp Vault address    |
| `VAULT_TOKEN`      | `dev-root-token`           | Vault authentication token |
| `DB_ROUTER_HOST`   | `db.0.xeze.org:443`       | Database Router gRPC host  |

## License

Apache 2.0
