#!/bin/bash
set -euo pipefail

VERSION="${VERSION:-$(git describe --tags --always)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}"

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -ldflags="${LDFLAGS}" \
  -o concave .
echo "Built: concave ($(du -sh concave | cut -f1))"

mkdir -p scripts/completions
./concave completion bash > scripts/completions/concave.bash
./concave completion zsh > scripts/completions/concave.zsh
./concave completion fish > scripts/completions/concave.fish
echo "Generated: shell completions"
