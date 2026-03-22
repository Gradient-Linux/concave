#!/bin/bash
set -euo pipefail

bash scripts/build.sh

if ! command -v goreleaser >/dev/null 2>&1; then
  echo "ERROR: goreleaser is required. Install it from https://goreleaser.com/install/"
  exit 1
fi

if [ "$#" -eq 0 ]; then
  exec goreleaser release --snapshot --clean
fi

exec goreleaser "$@"
