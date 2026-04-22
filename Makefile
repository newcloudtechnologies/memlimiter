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
ALLOCATOR_TEST_BIN := ./test/allocator/allocator

PYTHON_BIN ?= python3.14
PYTHON_REQUIREMENTS_FILE := requirements.txt
PYTHON_VENV_DIR ?= .venv
PYTHON_VENV_BIN := $(PYTHON_VENV_DIR)/bin
PYTHON_VENV_PYTHON := $(PYTHON_VENV_BIN)/python
PYTHON_ANALYZE_DIR := ./test/allocator/analyze
PYTHON_ANALYZE_SCRIPT := $(PYTHON_ANALYZE_DIR)/compare.py

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

.PHONY: python-venv
python-venv:
## Create venv with PYTHON_BIN and update pip.
	$(PYTHON_BIN) -m venv $(PYTHON_VENV_DIR)
	$(PYTHON_VENV_PYTHON) -m pip install --upgrade pip

.PHONY: python-deps
python-deps: python-venv
## Install Python dependencies from requirements file.
	$(PYTHON_VENV_PYTHON) -m pip install -r $(PYTHON_REQUIREMENTS_FILE)

.PHONY: python-check
python-check: python-deps
## Validate Python dependency graph and script syntax.
	$(PYTHON_VENV_PYTHON) -m pip check
	$(PYTHON_VENV_PYTHON) -m py_compile $(PYTHON_ANALYZE_DIR)/*.py

.PHONY: docker-check
docker-check:
## Ensure Docker daemon is reachable.
	@docker info >/dev/null 2>&1 || { \
		echo "error: Docker is required. Install Docker and make sure the daemon is running, then retry."; \
		exit 1; \
	}

.PHONY: generate
generate:
## Regenerate protobuf stubs for allocator schema.
## Requires protoc to be installed: https://grpc.io/docs/protoc-installation/
	@command -v protoc >/dev/null || { \
		echo "error: protoc is required. Install it first: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	}
	cd test/allocator/schema && ./generate.sh

.PHONY: allocator-build
allocator-build:
## Build allocator binary.
	go build -o $(ALLOCATOR_TEST_BIN) ./test/allocator

.PHONY: build
build: allocator-build
## Build all Go packages.
	go build ./...

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

.PHONY: lint-prepare
lint-prepare: install-lint
## Verify linter configuration.
	$(GOLANGCI_LINT) config verify

.PHONY: lint
lint: lint-prepare
## Analyze code locally.
## Use project-local golangci-lint from ./bin.
	$(GOLANGCI_LINT) run ./...

.PHONY: fix
fix: lint-prepare
## Apply automatic source fixes.
## Run go mod tidy, and golangci-lint --fix.
	go mod tidy
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: allocator-analyze
allocator-analyze: allocator-build docker-check python-check
## Run allocator docker benchmark and render plots.
## Requires Docker daemon and test allocator binary.
	$(PYTHON_VENV_PYTHON) $(PYTHON_ANALYZE_SCRIPT)

.PHONY: clean
clean:
## Remove generated build, test, and Python cache artifacts.
	rm -f $(UNIT_COVERAGE_FILE) $(INTEGRATION_COVERAGE_FILE) $(OVERALL_COVERAGE_FILE) $(COVERAGE_SUMMARY_FILE) $(INTEGRATION_TEST_BIN) $(ALLOCATOR_TEST_BIN)
	rm -rf "$(PYTHON_ANALYZE_DIR)/__pycache__"
