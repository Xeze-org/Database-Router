# xeze-dbr

C++23 gRPC client for the [Xeze Database Router](https://github.com/Xeze-org/Database-Router) — a unified interface for **PostgreSQL**, **MongoDB**, and **Redis** over **mTLS**.

## Install

This library is built with CMake and strictly enforces C++23 standards. You can easily pull it into your existing CMake infrastructure utilizing `FetchContent`.

```cmake
# In your CMakeLists.txt
include(FetchContent)

FetchContent_Declare(xeze_dbr
    GIT_REPOSITORY https://code.xeze.org/xeze/Database-Router.git
    GIT_TAG main
    SOURCE_SUBDIR sdk/cpp
)
FetchContent_MakeAvailable(xeze_dbr)

target_link_libraries(your_target PUBLIC xeze_dbr)
```

## Quick Start

```cpp
#include "xeze/dbr/client.hpp"
#include <iostream>

int main() {
    xeze::dbr::Options opts;
    opts.host = "db.0.xeze.org:443";
    opts.cert_file = "client.crt";
    opts.key_file = "client.key";

    // Establish mTLS secure connection to the Router
    xeze::dbr::Client client(opts);

    // Run Health Check
    dbrouter::HealthCheckRequest req;
    dbrouter::HealthCheckResponse resp;
    grpc::ClientContext ctx;

    grpc::Status status = client.health->Check(&ctx, req, &resp);
    if(status.ok()) {
        std::cout << "Router Healthy? " << resp.overall_healthy() << std::endl;
    }

    return 0;
}
```

## License

Apache 2.0
