# AGENTS.override.md — concave
# This file overrides root AGENTS.md for anything inside concave/ only.
# Per Codex discovery rules, this file is merged AFTER root AGENTS.md,
# so anything here wins for work done inside this directory.

## concave-specific overrides

### Language Target
Go 1.25-compatible source. CI and goreleaser pinned to Go 1.26.1.

### Build Command
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.Version=$(git describe --tags --always)" \
  -o concave .
```

### Test Commands
```bash
go test ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Lint
```bash
golangci-lint run ./...
go vet ./...
staticcheck ./...
govulncheck ./...
```

### Single Permitted External Dependency
`github.com/spf13/cobra v1.8.0`
All other dependencies require PM Agent approval documented in PR description.

### File Ownership Enforcement
No agent modifies a file owned by another agent without PM Agent sequencing the edit.
See root AGENTS.md §4 for the full ownership table.

### Template Directory Rule
`templates/` contains exactly four files. No subdirectories.
No README. No other files. Any PR violating this is a Code Reviewer blocker.

### Documentation Location Rule
All documentation lives in `docs/` or inline godoc.
Suite-level docs live in `docs/suites/<suite>.md` only.
No `services/` directory. No `docs/suites/<suite>/` subdirectory trees.
