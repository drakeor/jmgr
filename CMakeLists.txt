cmake_minimum_required(VERSION 3.15)

# Project definition
project(jmgr
  LANGUAGES CXX
  VERSION 1.0.0
)

# Fetch gRPC
include(FetchContent)
set(ABSL_ENABLE_INSTALL ON)
FetchContent_Declare(
  gRPC
  GIT_REPOSITORY https://github.com/grpc/grpc
  GIT_TAG        v1.66.2  # e.g v1.28.0
)
set(FETCHCONTENT_QUIET OFF)
FetchContent_MakeAvailable(gRPC)

#  Fetch FTXUI 
set(FETCHCONTENT_UPDATES_DISCONNECTED TRUE)
FetchContent_Declare(ftxui
  GIT_REPOSITORY https://github.com/ArthurSonzogni/ftxui
  GIT_TAG v5.0.0
)
 
FetchContent_GetProperties(ftxui)
if(NOT ftxui_POPULATED)
  FetchContent_Populate(ftxui)
  add_subdirectory(${ftxui_SOURCE_DIR} ${ftxui_BINARY_DIR} EXCLUDE_FROM_ALL)
endif()

# Set the output directory
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

# Build the server manager tool
add_executable(jmgr
  src/main.cpp
)

add_executable(jmgr-service
  src/main.cpp
)

target_include_directories(jmgr PRIVATE src)
 
target_link_libraries(jmgr
  PRIVATE ftxui::screen
  PRIVATE ftxui::dom
  PRIVATE ftxui::component
  grpc++
)

target_link_libraries(jmgr-service
  grpc++
)