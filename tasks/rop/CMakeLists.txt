cmake_minimum_required(VERSION 3.15)
project(rop C)

set(CMAKE_C_FLAGS "-static -no-pie -fno-stack-protector")

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_executable(${PROJECT_NAME} src/rop.c)
target_include_directories(${PROJECT_NAME} PRIVATE ${PROJECT_SOURCE_DIR}/include)

