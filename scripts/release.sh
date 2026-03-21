#!/bin/bash
set -euo pipefail

mkdir -p dist

VERSION=$(git describe --tags --always)

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o dist/concave-linux-amd64 .

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o dist/concave-linux-arm64 .

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o dist/concave-tui-linux-amd64 ./cmd/concave-tui/

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags="-s -w -X main.Version=${VERSION}" \
  -o dist/concave-tui-linux-arm64 ./cmd/concave-tui/
