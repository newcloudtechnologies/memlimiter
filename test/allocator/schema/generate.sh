#!/usr/bin/env bash

#
# Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
# Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
# License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
#

set -euo pipefail
set -x

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/../../.." && pwd)"
TOOLS_BIN="$(mktemp -d "${TMPDIR:-/tmp}/memlimiter-proto-tools.XXXXXX")"
trap 'rm -rf "${TOOLS_BIN}"' EXIT

PROTOBUF_VERSION="$(cd "${REPO_ROOT}" && go list -m -f '{{.Version}}' google.golang.org/protobuf)"
PROTOC_GEN_GO_GRPC_VERSION="$(cd "${REPO_ROOT}" && go list -m -f '{{.Version}}' google.golang.org/grpc/cmd/protoc-gen-go-grpc)"

GOBIN="${TOOLS_BIN}" go install "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOBUF_VERSION}"
GOBIN="${TOOLS_BIN}" go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}"

PATH="${TOOLS_BIN}:${PATH}" protoc \
	-I/usr/include \
	-I. \
	--go_out=. \
	--go_opt=paths=source_relative \
	--go-grpc_out=. \
	--go-grpc_opt=paths=source_relative \
	allocator.proto
