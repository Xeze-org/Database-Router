cmake_minimum_required(VERSION 3.24)
project(xeze-dbr-core CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED True)

include(FetchContent)

# 1. Fetch JSON parser (nlohmann_json)
FetchContent_Declare(json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(json)

# 2. Fetch CPR (C++ Requests / libcurl wrapper for easy HTTP)
FetchContent_Declare(cpr
    GIT_REPOSITORY https://github.com/libcpr/cpr.git
    GIT_TAG 1.10.5
)
set(CPR_USE_SYSTEM_CURL OFF CACHE BOOL "Use built-in libcurl")
set(CPR_BUILD_TESTS OFF CACHE BOOL "Disable cpr tests")
FetchContent_MakeAvailable(cpr)

# 3. Add the Local SDK dependency
add_subdirectory(${CMAKE_SOURCE_DIR}/../../sdk/cpp ${CMAKE_BINARY_DIR}/sdk)

# 4. Build Core Library
add_library(xeze_dbr_core
    src/core.cpp
)

target_include_directories(xeze_dbr_core PUBLIC
    $<BUILD_INTERFACE:${CMAKE_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

target_link_libraries(xeze_dbr_core PUBLIC
    xeze_dbr
    nlohmann_json::nlohmann_json
    cpr::cpr
)
