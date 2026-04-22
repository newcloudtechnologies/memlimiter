#!/usr/bin/env bash

# Merge unit and integration Go coverage profiles into one report.
# Usage:
#   scripts/merge_coverage.sh <unit-coverage> <integration-coverage> <overall-coverage> <summary-output>

set -euo pipefail

if [[ "$#" -ne 4 ]]; then
	echo "usage: $0 <unit-coverage> <integration-coverage> <overall-coverage> <summary-output>" >&2
	exit 2
fi

# Input/output files.
unit_coverage_file="$1"
integration_coverage_file="$2"
overall_coverage_file="$3"
coverage_summary_file="$4"

# Copy unit coverage and append integration coverage body (skip header line).
cp "${unit_coverage_file}" "${overall_coverage_file}"
tail --lines=+2 "${integration_coverage_file}" >> "${overall_coverage_file}"

# Exclude binaries/entrypoints that are intentionally out of test coverage scope.
sed -i '/test\/allocator\/app/d' "${overall_coverage_file}"
sed -i '/test\/allocator\/main.go/d' "${overall_coverage_file}"

# Produce a human-readable coverage summary.
go tool cover -func="${overall_coverage_file}" -o="${coverage_summary_file}"
