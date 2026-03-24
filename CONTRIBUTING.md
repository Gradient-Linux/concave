# Contributing to concave

Contributions are welcome for CLI behavior, API handlers, suite lifecycle, workspace tooling, GPU support, tests, and documentation. Keep changes focused on the control plane. Client-side presentation work belongs in `concave-tui` and `concave-web`.

## Before you start

Read these documents before changing behavior:

- [README.md](README.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/concave-reference.md](docs/concave-reference.md)
- [docs/suite-guide.md](docs/suite-guide.md)

## Development setup

Use Ubuntu 24.04 with Go 1.25 or newer and Docker Engine installed.

```bash
git clone <repo-url>
cd concave
CGO_ENABLED=0 go build -o concave .
go test ./...
go test -race ./...
./concave --help
```

If you want to exercise live suite flows, start from a machine with Docker available and a writable `~/gradient/` workspace.

## Making changes

### Branching

Use one of these branch prefixes:

- `feat/<slug>`
- `fix/<slug>`
- `docs/<slug>`

### Commit messages

Format commits as `<type>(<scope>): <summary>`.

Use these types:

- `feat`
- `fix`
- `refactor`
- `test`
- `docs`
- `chore`

Keep the summary under 72 characters.

Examples:

- `feat(gpu): add secure boot enrollment guidance`
- `fix(workspace): keep outputs cleanup scoped to workspace root`
- `docs(reference): document resolver status commands`

### Tests

- Add or update unit tests for any new function or behavior change.
- Run `go test ./...` before opening a pull request.
- Run `go test -race ./...` when you touch shared state, jobs, or long-lived goroutines.
- Integration tests belong in `tests/integration/` and should stay opt-in.

### Pull requests

- Keep pull requests focused on one logical change.
- Explain what changed, why it changed, and how you verified it.
- Update user-facing docs in the same pull request when command behavior changes.

## Code conventions

- All terminal output in `cmd/` must go through `internal/ui/printer.go`.
- Do not use `fmt.Println` or `log.Printf` in `cmd/`.
- Docker-facing functions should accept `context.Context` first.
- Keep direct dependencies tightly controlled. Any new dependency needs prior discussion in an issue.
- `internal/suite/registry.go` is the single source of truth for suite names, images, ports, and mounts.
- `cmd/` is the only layer that may call `os.Exit`.
- Return errors up the call stack and wrap them with context, for example `fmt.Errorf("docker pull %s: %w", image, err)`.

## What we don't accept

- Dependencies added without prior discussion in an issue.
- Code that writes outside `~/gradient/` without explicit user confirmation.
- Hardcoded image tags outside `internal/suite/registry.go`.
- Shell string interpolation with user-controlled input.

## License

By contributing, you agree that your contributions are licensed under the MIT License.
