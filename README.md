# Gradient Linux — concave

`concave` is the infrastructure control plane for Gradient Linux. It is responsible
for host checks, suite lifecycle, workspace management, GPU integration, and the
authenticated local API server used by the terminal and browser frontends.

The project is intentionally infrastructure-first:

- headless-safe CLI behavior
- Docker- and workspace-centric lifecycle control
- packaged systemd deployment for the local control plane
- Unix-group-backed authorization for machine operations
- a stable API for `concave-web` and `concave-tui`

## Responsibilities

This repository owns:

- the `concave` CLI
- the `concave serve` local control-plane server
- suite installation and lifecycle logic for Boosting, Neural, Flow, and Forge
- `~/gradient/` or service-root workspace layout management
- Docker Compose template rendering and validation
- GPU detection and driver guidance
- Unix-group role resolution, PAM login, JWT issuance, and permission checks
- packaged system integration: systemd unit, sudoers helper, shell completions, and release artifacts

This repository does not own:

- the Bubble Tea terminal UI
- the browser UI
- marketing or site content

Those live in `concave-tui` and `concave-web`.

## Repository layout

- `cmd/`: Cobra command surface, including `serve` and `whoami`
- `internal/api/`: authenticated HTTP and WebSocket control plane
- `internal/auth/`: Unix group roles, PAM auth, JWT sessions, permission checks
- `internal/config/`: workspace state, version manifests, setup-state persistence
- `internal/docker/`: Compose rendering, Docker process wrappers, retries, image helpers
- `internal/gpu/`: GPU detection, NVIDIA/AMD helpers
- `internal/suite/`: suite registry, Forge composition, installers, health checks
- `internal/system/`: locks, exit codes, logging, crash logs, privileged helpers, host checks
- `internal/ui/`: CLI output, prompts, and progress display
- `internal/workspace/`: fixed workspace layout, status, backup, and cleanup
- `templates/`: canonical Compose YAML templates
- `scripts/`: build, release, service, apt, and postinstall helpers
- `docs/`: operator, contributor, suite, and API documentation

## Runtime model

The normal deployment model is:

1. Ubuntu host
2. Docker Engine
3. `concave` installed as a system binary
4. `concave-serve.service` running as `gradient-svc`
5. frontends talking to `concave serve`

The control plane exposes machine state without forcing the host binary itself to
grow a UI dependency tree.

## Role model

Access is derived from Unix groups:

- `gradient-viewer`
- `gradient-developer`
- `gradient-operator`
- `gradient-admin`

The CLI resolves the current user's role directly from the host. Browser and TUI
clients authenticate through `concave serve`, which uses PAM for password checking
and JWTs for session continuity.

Useful commands:

```bash
concave whoami
concave serve --addr 127.0.0.1:7777
```

## Workspace layout

User-facing installs default to `~/gradient/`. Service deployments use the
configured service root, typically `/var/lib/gradient`.

```text
gradient/
  data/
  notebooks/
  models/
  outputs/
  mlruns/
  dags/
  compose/
  config/
  backups/
  logs/
```

## Quick start

Package install on an Ubuntu host:

```bash
curl -fsSL https://packages.gradientlinux.io/install.sh | sudo bash
concave doctor
concave setup
concave whoami
```

Manual local development:

```bash
go test ./...
go test -race ./...
go build -o concave .
./concave doctor
./concave workspace init
```

Verbose mode:

```bash
concave --verbose status
```

This keeps normal command output on stdout and writes structured diagnostics to stderr.

## Core commands

- `concave doctor`
- `concave workspace init|status|backup|clean`
- `concave install|remove|start|stop|restart|update|rollback <suite>`
- `concave logs <suite>`
- `concave lab`
- `concave whoami`
- `concave serve`

See [docs/concave-reference.md](docs/concave-reference.md) for the full command
surface.

## Suite map

- [Boosting](docs/suites/boosting.md): notebooks, experiments, MLflow
- [Neural](docs/suites/neural.md): GPU-oriented training, inference, lab workflows
- [Flow](docs/suites/flow.md): orchestration, dashboards, storage, serving
- [Forge](docs/suites/forge.md): user-selected composition across suite components

## Documentation

Start with [docs/README.md](docs/README.md).

Important docs:

- [docs/architecture.md](docs/architecture.md)
- [docs/api.md](docs/api.md)
- [docs/system-admin.md](docs/system-admin.md)
- [docs/concave-reference.md](docs/concave-reference.md)
- [docs/suite-guide.md](docs/suite-guide.md)
- [docs/gpu-setup.md](docs/gpu-setup.md)

## Companion repos

- `concave-web`: browser control plane and proxy
- `concave-tui`: terminal UI client

## Contributing

Contributor-facing rules live in [CONTRIBUTING.md](CONTRIBUTING.md). Public docs,
README updates, and behavior changes should land together.

## License

This project is released under the [MIT License](LICENSE).
