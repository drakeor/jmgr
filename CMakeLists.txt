cmake_minimum_required(VERSION 3.15)
include(FetchContent)

# Project definition
project(jmgr
  LANGUAGES CXX
  VERSION 1.0.0
)

# Set the output directory
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

# Use ZeroMQ, find it's package
find_package(ZeroMQ REQUIRED)

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


# Build the server manager tool
add_executable(jmgr
  src/main_app.cpp
)

# Build the server manager service
add_executable(jmgr-service
  src/main_daemon.cpp
)

# Build the splash screen
add_executable(jmgr-splash
  src/main_splash.cpp
  src/tiv/image_view.cpp
  src/tiv/tiv_lib.cpp

  src/metrics/uptime.cpp
  src/metrics/kernel_info.cpp
  src/metrics/cpu_info.cpp
  src/metrics/os_info.cpp
  src/metrics/memory_info.cpp
  src/metrics/process_info.cpp
  src/metrics/disk_usage.cpp
)

target_include_directories(jmgr PRIVATE src)
target_include_directories(jmgr-splash PRIVATE src)
 
target_link_libraries(jmgr
  PRIVATE ftxui::screen
  PRIVATE ftxui::dom
  PRIVATE ftxui::component
#  grpc++
  zmq
)

target_link_libraries(jmgr-service
#  grpc++
  zmq
)

target_link_libraries(jmgr-splash
  PRIVATE ftxui::screen
  PRIVATE ftxui::dom
  PRIVATE ftxui::component
#  grpc++
  zmq
)