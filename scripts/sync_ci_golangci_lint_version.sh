#!/usr/bin/env bash
#
# Sync golangci-lint version in CI workflow with Makefile pin.
# Usage:
#   scripts/sync_ci_golangci_lint_version.sh [makefile] [workflow]

set -euo pipefail

if [[ "$#" -gt 2 ]]; then
	echo "usage: $0 [makefile] [workflow]" >&2
	exit 2
fi

makefile_path="${1:-Makefile}"
workflow_path="${2:-.github/workflows/CI.yml}"

if [[ ! -f "${makefile_path}" ]]; then
	echo "makefile not found: ${makefile_path}" >&2
	exit 1
fi

if [[ ! -f "${workflow_path}" ]]; then
	echo "workflow file not found: ${workflow_path}" >&2
	exit 1
fi

# Read pinned local linter version from Makefile.
desired_version="$(
	sed -nE 's/^[[:space:]]*GOLANGCI_LINT_VERSION[[:space:]]*[:?+]?=[[:space:]]*(v[0-9]+\.[0-9]+\.[0-9]+).*/\1/p' "${makefile_path}" | head -n1
)"

if [[ -z "${desired_version}" ]]; then
	echo "failed to read GOLANGCI_LINT_VERSION from ${makefile_path}" >&2
	exit 1
fi

# Read current CI linter version for status output.
current_version="$(
	awk '
		/^[[:space:]]*uses:[[:space:]]*golangci\/golangci-lint-action@/ { in_step=1; next }
		in_step && /^[[:space:]]*version:[[:space:]]*/ {
			sub(/^[[:space:]]*version:[[:space:]]*/, "", $0)
			print $0
			exit
		}
		in_step && /^[[:space:]]*-[[:space:]]*name:[[:space:]]*/ { in_step=0 }
	' "${workflow_path}"
)"

if [[ -z "${current_version}" ]]; then
	echo "failed to read golangci-lint version from ${workflow_path}" >&2
	exit 1
fi

tmp_file="$(mktemp)"
trap 'rm -f "${tmp_file}"' EXIT

if ! awk -v desired_version="${desired_version}" '
	/^[[:space:]]*uses:[[:space:]]*golangci\/golangci-lint-action@/ { in_step=1 }
	in_step && /^[[:space:]]*version:[[:space:]]*/ {
		sub(/version:[[:space:]]*.*/, "version: " desired_version)
		updated=1
		in_step=0
	}
	{ print }
	END {
		if (updated != 1) {
			exit 3
		}
	}
' "${workflow_path}" > "${tmp_file}"; then
	echo "failed to update golangci-lint version in ${workflow_path}" >&2
	exit 1
fi

if cmp -s "${tmp_file}" "${workflow_path}"; then
	echo "golangci-lint CI version already synced (${current_version})"
	exit 0
fi

mv "${tmp_file}" "${workflow_path}"
trap - EXIT

echo "updated ${workflow_path}: golangci-lint version ${current_version} -> ${desired_version}"
