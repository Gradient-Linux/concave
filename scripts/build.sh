#!/bin/bash
set -euo pipefail

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=$(git describe --tags --always)" \
  -o concave .
