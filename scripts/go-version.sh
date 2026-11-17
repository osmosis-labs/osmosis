#!/usr/bin/env bash
# Prints the exact Go version every build of this repo uses (for example 1.26.10).
#
# go.mod is the single source of truth: the `toolchain` line when present, else the `go` line.
# The Makefile, the Docker build workflows and `actions/setup-go` all derive the version from it.
#
# Usage: scripts/go-version.sh [path/to/go.mod]
set -euo pipefail

gomod="${1:-$(dirname "$0")/../go.mod}"

version="$(awk '
  $1 == "toolchain" && $2 ~ /^go[0-9]/ { sub(/^go/, "", $2); toolchain = $2 }
  $1 == "go"        && $2 ~ /^[0-9]/   { gover = $2 }
  END { print (toolchain != "" ? toolchain : gover) }
' "$gomod")"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "go-version.sh: could not read an exact Go version (major.minor.patch) from $gomod, got '$version'" >&2
  exit 1
fi

echo "$version"
