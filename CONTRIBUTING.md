# Contributing to Gradient Linux / concave

`concave` is the control-plane CLI for Gradient Linux. This document is the public
source of truth for contributor-facing expectations in this repository: architecture
rules, documentation layout, code conventions, testing gates, and pull request flow.

Maintainers may use their own internal tooling, automation, or agentic workflows while
building and reviewing the project. Those internal workflows are not part of the public
contribution contract. If you contribute with your own local tooling or agents, the
result still needs to follow this document.

## Scope

- Active repository: `github.com/Gradient-Linux/concave`
- Go module: `github.com/Gradient-Linux/concave`
- Target platform: Ubuntu 24.04 LTS
- Primary deliverable: a static `concave` Go binary

Gradient Linux keeps the host thin. Docker Engine and the `concave` binary live on the
host. AI frameworks, notebooks, model services, tracking, and orchestration all run in
containers. Do not add host Python installation paths to this project.

## Repository Layout

```text
concave/
  CONTRIBUTING.md
  README.md
  CHANGELOG.md
  main.go
  go.mod
  cmd/
  internal/
  templates/
  scripts/
  tests/
    integration/
    benchmarks/
  docs/
    architecture.md
    concave-reference.md
    gpu-setup.md
    suite-guide.md
    suites/
      boosting.md
      neural.md
      flow.md
      forge.md
```

## Documentation Layout

There are exactly two documentation locations in this repo:

- `docs/`
- inline godoc in Go source

Rules:

- System-wide behavior belongs in `docs/`.
- Suite-level prose belongs in `docs/suites/*.md`, one file per suite.
- Service-level details belong inside the relevant suite doc, not in a separate tree.
- There is no `services/` directory.
- There is no `docs/suites/<suite>/` directory tree.
- There are no README files inside `templates/`.
- `templates/` is a flat directory containing only the four Compose YAML files.

If your change alters user-facing behavior, command flow, suite topology, container
ports, environment variables, or workspace mounts, update the matching document in the
same pull request.

## Local Setup

Requirements:

- Ubuntu 24.04 LTS preferred
- Go 1.25.0 or newer locally
- Docker Engine
- `golangci-lint` for local lint checks
- optional: `goreleaser` for local release artifact validation
- optional: `syft` for local SPDX SBOM generation during release validation

Clone and build:

```bash
git clone git@github.com:Gradient-Linux/concave.git
cd concave
go build -o concave .
go test ./...
go test -race ./...
go vet ./...
bash scripts/build.sh
```

Optional local lint:

```bash
golangci-lint run ./...
```

## Branching and Pull Requests

Human contributors branch from `main` using one of these prefixes:

- `feat/`
- `fix/`
- `docs/`
- `refactor/`
- `test/`

Examples:

- `feat/boosting-log-follow`
- `fix/compose-rollback-cleanup`
- `docs/update-suite-guide`

Workflow:

1. Branch from `main`.
2. Make the smallest coherent change possible.
3. Run the required checks locally.
4. Update docs in the same branch when behavior changes.
5. Open a pull request targeting `dev`, never `main`.

Maintainers handle review, additional automation, and the final `dev` to `main` merge.

## Commit Format

Use Conventional Commits:

```text
<type>(<scope>): <short description>
```

Common types:

- `feat`
- `fix`
- `docs`
- `refactor`
- `test`
- `chore`
- `perf`

Common scopes:

- `core`
- `suite`
- `gpu`
- `infra`
- `templates`
- `docs`

Examples:

- `feat(suite): add boosting install rollback cleanup`
- `fix(infra): delete invalid compose files on validation failure`
- `docs(docs): align suite guide with current install flow`

## Code Rules

### General

- Use Go 1.25-compatible code. CI runs on Go 1.26.1.
- Direct dependencies are intentionally minimal and require maintainer approval.
- Approved direct dependencies today are:
  - `github.com/spf13/cobra v1.8.0`
- New dependencies require explicit maintainer approval.
- Keep functions small and easy to test.
- Wrap errors with context.
- Leave the system clean on failure.

### Output and UX

- All command output in `cmd/` goes through `internal/ui/`.
- Do not use `fmt.Println` or `log.Printf` in `cmd/`.
- User-facing errors must explain recovery when practical.

### Paths, images, and ports

- Do not hardcode image tags outside `internal/suite/registry.go`.
- Do not hardcode workspace paths outside `internal/workspace/init.go`.
- Do not hardcode port assignments outside suite definitions and the shared port logic.
- `~/gradient/` is the fixed workspace root.

### External commands

- Use `exec.Command` or `exec.CommandContext` with separate arguments.
- Never build shell commands through string interpolation.
- Code that shells out must stay testable through an injectable command seam.
- Never call real Docker or GPU binaries in unit tests.

### Privilege boundaries

- Only `internal/gpu/nvidia.go` and `cmd/driver_wizard.go` may invoke `sudo`.
- Do not write to `/etc`, `/usr`, or `/var` outside approved GPU/setup flows.
- Do not use `--privileged` or `--network host`.

### Data safety

- `remove` and `rollback` must never touch:
  - `~/gradient/data/`
  - `~/gradient/models/`
  - `~/gradient/notebooks/`
- Invalid generated Compose files must be deleted before returning an error.
- Mutating commands must remain safe to interrupt and safe to rerun.

## Testing and Quality Gates

Run these locally before opening a PR:

```bash
go test ./...
go test -race ./...
go vet ./...
CGO_ENABLED=0 go build -o concave .
```

Coverage gate:

- overall coverage must be at least 80%
- no package may fall below 60%

Integration tests:

- live under `tests/integration/`
- must skip unless `CONCAVE_INTEGRATION=1` is set
- are for real environment validation, not default CI execution

Example:

```bash
CONCAVE_INTEGRATION=1 go test ./tests/integration -v
```

GPU-related changes must include manual validation notes in the PR description.

## Documentation Expectations

Update docs when you change:

- command behavior
- suite topology
- ports
- environment variables
- workspace mounts
- GPU setup flow
- rollback/update behavior

Required doc targets:

- system behavior: `docs/*.md`
- suite behavior and service internals: `docs/suites/<suite>.md`
- exported Go symbols: godoc comments in source

## What Gets Rejected

These changes will be rejected:

- PRs opened directly to `main`
- new dependencies without approval
- `fmt.Println` or `log.Printf` in `cmd/`
- hardcoded image tags outside `internal/suite/registry.go`
- hardcoded workspace paths outside `internal/workspace/init.go`
- docs added outside `docs/` or inline godoc
- a `services/` directory or docs in `templates/`
- shell string interpolation in `exec.Command`
- `sudo` outside approved GPU/setup files
- changes that modify user data during remove or rollback
- undocumented behavior changes

## Internal Workflows

This repository may also be developed with private internal scaffolding that is not
checked into version control. That internal workflow is not the contributor contract.

For public contributions, `CONTRIBUTING.md` is the repo policy.

## Security Reports

Do not open public issues for security vulnerabilities. Report them privately to the
maintainers through the project security contact.
