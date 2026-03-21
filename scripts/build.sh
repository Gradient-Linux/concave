#!/bin/bash
set -euo pipefail

VERSION=$(git describe --tags --always)

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o concave .
echo "Built: concave ($(du -sh concave | cut -f1))"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o concave-tui ./cmd/concave-tui/
echo "Built: concave-tui ($(du -sh concave-tui | cut -f1))"
