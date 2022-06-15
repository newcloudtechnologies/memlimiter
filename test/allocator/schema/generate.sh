#!/bin/bash

#
# Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
# Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
# License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
#

set -e
set -x

protoc  -I/usr/include \
        -I. \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        allocator.proto
