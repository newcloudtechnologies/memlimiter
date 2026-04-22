# Make Workflows (Linux / WSL)

This file describes how to use project `Makefile` targets and what output to expect.

## Prerequisites

- Go toolchain available in `PATH`.
- Python `3.14` available as `python3.14` (or override with `PYTHON_BIN`).
- Docker daemon running for allocator analysis.
- `protoc` installed only if you run `make generate`.

## Quick Start

Run these once on a fresh checkout:

```bash
make install-lint
make python-check
```

Expected results:

- `bin/golangci-lint` is installed.
- `.venv` is created and dependencies are installed from `requirements.txt`.
- `pip check` prints `No broken requirements found.`

## Daily Development Flow

### 1) Build

```bash
make build
```

What it does:

- Builds allocator demo binary: `test/allocator/allocator`.
- Builds all Go packages in the repository.

### 2) Lint

```bash
make lint
```

What it does:

- Installs pinned `golangci-lint` if needed.
- Verifies linter config.
- Runs `golangci-lint run ./...`.

### 3) Auto-fix

```bash
make fix
```

What it does:

- Runs `go mod tidy`.
- Runs `golangci-lint` with `--fix`.

### 4) Tests

```bash
make unit-test
make integration-test
make test
```

Expected artifacts:

- `coverage.unit.out`
- `coverage.integration.out`
- `coverage.overall.out`
- `coverage.out` (human-readable summary)
- `test/integration/integration-test` (integration test binary)

## Allocator Analysis Flow

Run:

```bash
make allocator-analyze
```

This target runs:

1. `make allocator-build`
2. `make docker-check`
3. `make python-check`
4. Python benchmark/plot script: `test/allocator/analyze/compare.py`

Expected console signals:

- Lines like `>>> Start case: ...`
- Progress logs from allocator perf client.

Expected output directory:

- `/tmp/allocator/allocator_<HHMMSS>/`

Expected generated files:

- `control_params.png`
- `gogc_floor_hits.png`
- `memory_limits_overlay.png`
- `rss.png`
- Per-case directories with:
  - `server_config.json`
  - `perf_config.json`
  - `tracker.csv`

## Utility Targets

- `make help` - print all available targets.
- `make python-venv` - create `.venv` and upgrade `pip`.
- `make python-deps` - install dependencies from `requirements.txt`.
- `make python-check` - run `pip check` and Python syntax compile.
- `make docker-check` - fail fast if Docker daemon is unavailable.
- `make allocator-build` - build `test/allocator/allocator` only.
- `make generate` - regenerate protobuf files for allocator schema.
- `make lint-prepare` - install lint tools and verify lint config.
- `make sync-ci-lint-version` - copy `GOLANGCI_LINT_VERSION` from `Makefile` to `.github/workflows/CI.yml`.
- `make clean` - remove generated binaries, coverage files, and Python cache for analyzer scripts.

## Common Overrides

Use a different Python interpreter:

```bash
make python-check PYTHON_BIN=python3.13
```

Use a different virtual environment directory:

```bash
make python-check PYTHON_VENV_DIR=.venv-local
```
