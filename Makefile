SHELL := /bin/bash

.ONESHELL:
.SHELLFLAGS := -euo pipefail -c

# By default, show help.
.DEFAULT_GOAL := help

LOCAL_BIN := $(CURDIR)/bin
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v2.11.4
GOLANGCI_LINT_INSTALL_METHOD := build

UNIT_TEST_PACKAGES := $(shell go list ./...)
UNIT_COVERAGE_FILE := coverage.unit.out
INTEGRATION_COVERAGE_FILE := coverage.integration.out
OVERALL_COVERAGE_FILE := coverage.overall.out
COVERAGE_SUMMARY_FILE := coverage.out
INTEGRATION_TEST_BIN := ./test/integration/integration-test

.PHONY: help
help:
## Show help for available targets.
	@awk -f scripts/make_help.awk $(MAKEFILE_LIST)

.PHONY: install-lint
install-lint:
## Install pinned golangci-lint version into ./bin.
## Use scripts/install_golangci_lint.sh with selected install method.
	@mkdir -p "$(LOCAL_BIN)"
	@sh -s -- "$(LOCAL_BIN)" "$(GOLANGCI_LINT_VERSION)" "$(GOLANGCI_LINT_INSTALL_METHOD)" < scripts/install_golangci_lint.sh

.PHONY: generate
generate:
## Regenerate protobuf stubs for allocator schema.
## Requires protoc to be installed: https://grpc.io/docs/protoc-installation/
	@command -v protoc >/dev/null || { \
		echo "error: protoc is required; install it first: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	}
	cd test/allocator/schema && ./generate.sh

.PHONY: build
build:
## Build all Go packages.
## Also build allocator demo binary.
	go build ./...
	go build ./test/allocator

.PHONY: unit-test
unit-test:
## Run unit tests with coverage for all packages.
	go test -v -count=1 -cover $(UNIT_TEST_PACKAGES) -coverprofile=$(UNIT_COVERAGE_FILE) -coverpkg ./...

.PHONY: integration-test
integration-test:
## Build and run integration test binary.
## Write integration coverage to a separate profile.
	go test -c ./test/integration -o $(INTEGRATION_TEST_BIN) -coverpkg ./...
	$(INTEGRATION_TEST_BIN) -test.v -test.coverprofile=$(INTEGRATION_COVERAGE_FILE)

.PHONY: test
test: unit-test integration-test
## Merge unit and integration coverage reports.
## Produce a human-readable coverage summary.
	./scripts/merge_coverage.sh \
		"$(UNIT_COVERAGE_FILE)" \
		"$(INTEGRATION_COVERAGE_FILE)" \
		"$(OVERALL_COVERAGE_FILE)" \
		"$(COVERAGE_SUMMARY_FILE)"

.PHONY: lint
lint: install-lint
## Analyze code locally.
## Use project-local golangci-lint from ./bin.
## Verify linter config before checks.
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run ./...

.PHONY: fix
fix: install-lint
## Apply automatic source fixes.
## Run go mod tidy, and golangci-lint --fix.
	go mod tidy
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: clean
clean:
## Remove generated test and coverage artifacts.
## Keep workspace clean between test runs.
	rm -f $(UNIT_COVERAGE_FILE) $(INTEGRATION_COVERAGE_FILE) $(OVERALL_COVERAGE_FILE) $(COVERAGE_SUMMARY_FILE) $(INTEGRATION_TEST_BIN)
