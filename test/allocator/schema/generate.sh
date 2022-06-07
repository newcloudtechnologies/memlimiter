#!/bin/bash

set -e
set -x

protoc  -I/usr/include \
        -I. \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        allocator.proto
