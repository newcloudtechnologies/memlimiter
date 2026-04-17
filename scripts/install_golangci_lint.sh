#!/usr/bin/env sh
# Purpose: install/update golangci-lint into the specified directory
# if the current version is missing or differs from the desired one.
# Script is provided via stdin:
# sh -s -- <install_dir> <version> < scripts/install_golangci_lint.sh

# Stop on errors and undefined variables.
# /bin/sh has no pipefail, so use only -e and -u.
set -eu

# Script arguments:
# $1 - destination directory for the golangci-lint binary.
# $2 - desired golangci-lint version (for example, v1.64.8 or v2.7.2).
# $3 - installation method: curl (default) or build.
# In some CI runners installation via curl may be disallowed, so use build.
# Validate that enough arguments were provided.
if [ "$#" -lt 2 ]; then
  echo "Usage: sh -s -- <install_dir> <version> [curl|build]" >&2
  exit 2
fi

# Save arguments into descriptive variable names.
LOCAL_BIN="$1"
DESIRED="$2"
INSTALL_METHOD="${3:-curl}"
BINARY="$LOCAL_BIN/golangci-lint"

case "$INSTALL_METHOD" in
  curl|download)
    INSTALL_METHOD="curl"
    ;;
  build|go)
    INSTALL_METHOD="build"
    ;;
  *)
    echo "Unknown install method '$INSTALL_METHOD'. Use 'curl' or 'build'." >&2
    exit 2
    ;;
esac

# Validate required tools for the chosen install method.
if [ "$INSTALL_METHOD" = "curl" ]; then
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required to install golangci-lint via curl" >&2
    exit 1
  fi
else
  if ! command -v go >/dev/null 2>&1; then
    echo "go toolchain is required to build golangci-lint from source" >&2
    exit 1
  fi
fi

# Create install directory if needed.
mkdir -p "$LOCAL_BIN"

# extract_ver extracts a version like vX.Y.Z from arbitrary text.
extract_ver() {
  # 1) "has version v?X.Y.Z".
  ver=$(sed -n 's/.*has version[[:space:]]\{1,\}v\{0,1\}\([0-9]\+\.[0-9]\+\.[0-9]\+\).*/v\1/p' | head -n1)
  if [ -n "$ver" ]; then printf '%s\n' "$ver"; return; fi

  # 2) "... version v?X.Y.Z ..." (with spaces around "version").
  ver=$(sed -n 's/.*[[:space:]]version[[:space:]]\{1,\}v\{0,1\}\([0-9]\+\.[0-9]\+\.[0-9]\+\)[^0-9.].*/v\1/p' | head -n1)
  if [ -n "$ver" ]; then printf '%s\n' "$ver"; return; fi

  # 3) Fallback: first standalone X.Y.Z token bounded by non-digit/non-dot.
  #    This avoids matching things like "go1.25.1".
  ver=$(sed -n 's/.*[^0-9.]\([0-9]\+\.[0-9]\+\.[0-9]\+\)[^0-9.].*/v\1/p' | head -n1)
  if [ -n "$ver" ]; then printf '%s\n' "$ver"; return; fi

  # 4) If version is at end of line (no trailing delimiter).
  ver=$(sed -n 's/.*[^0-9.]\([0-9]\+\.[0-9]\+\.[0-9]\+\)$$/v\1/p' | head -n1)
  [ -n "$ver" ] && printf '%s\n' "$ver"
}

# Detect currently installed version if the binary already exists.
installed=""

# Check whether executable binary exists.
if [ -x "$BINARY" ]; then
  # Run binary to get version output.
  out="$("$BINARY" --version 2>&1 || "$BINARY" version 2>&1 || true)"
  
  # Extract version from output.
  installed=$(printf '%s\n' "$out" | extract_ver || true)
fi

# Compare versions and install if needed.
if [ "$installed" != "$DESIRED" ]; then
  # Print action with previous version (or "none").
  echo "Installing golangci-lint $DESIRED into $LOCAL_BIN via $INSTALL_METHOD (was: ${installed:-none})"
  
  if [ "$INSTALL_METHOD" = "curl" ]; then
    # Download and execute official install script:
    # -s --: pass script arguments.
    # -b "$LOCAL_BIN": install directory.
    # "$DESIRED": version.
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
      | sh -s -- -b "$LOCAL_BIN" "$DESIRED"
  else
    # Use local Go toolchain to install the requested version.
    # For v2 major, module path must include /v2.
    case "$DESIRED" in
      v2.*|2.*)
        MODULE="github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$DESIRED"
        ;;
      *)
        MODULE="github.com/golangci/golangci-lint/cmd/golangci-lint@$DESIRED"
        ;;
    esac

    env GOBIN="$LOCAL_BIN" GO111MODULE=on GOWORK=off \
      go install "$MODULE"
  fi
else
  # Report that the current version is already up-to-date.
  echo "golangci-lint $DESIRED already installed in $LOCAL_BIN"
fi
