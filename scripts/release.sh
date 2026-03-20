#!/bin/bash
set -euo pipefail

mkdir -p dist

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=$(git describe --tags --always)" \
  -o dist/concave-linux-amd64 .

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags="-s -w -X main.Version=$(git describe --tags --always)" \
  -o dist/concave-linux-arm64 .
