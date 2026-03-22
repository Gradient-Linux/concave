# Gradient Linux — concave

`concave` is the control-plane interface for Gradient Linux, an Ubuntu 24.04 LTS
distribution built for machine learning engineers, data scientists, and MLOps teams.
The host stays thin: Ubuntu, Docker Engine, and the `concave` binary. Suites,
models, notebooks, tracking, orchestration, and observability stay inside containers.

## What This Repo Contains

- a statically compiled Go CLI (`concave`)
- an authenticated local control-plane server (`concave serve`)
- Docker Compose templates for Boosting, Neural, Flow, and Forge
- workspace lifecycle management for `~/gradient/`
- suite install, start, stop, update, rollback, and status flows
- Unix-group-based role resolution for CLI, TUI, and web clients
- JWT-backed API sessions for `concave serve`
- GPU detection and NVIDIA driver guidance
- system and suite documentation

## Architecture

- `cmd/` contains the Cobra command surface
- `internal/ui/` owns terminal output, spinners, and prompts
- `internal/system/` checks host prerequisites, browser launch, and shared port logic
- `internal/workspace/` manages the fixed `~/gradient/` layout
- `internal/docker/` renders Compose files, validates them, and wraps Docker operations
- `internal/suite/` defines suite metadata and lifecycle helpers
- `internal/config/` persists `state.json` and `versions.json`
- `internal/gpu/` detects GPU state and drives NVIDIA-specific checks
- `internal/auth/` owns Unix role resolution, PAM auth, JWT issuance, and permission checks
- `internal/api/` exposes the authenticated `concave serve` HTTP and WebSocket control plane
- `templates/` is a flat directory of the canonical Compose YAML templates
- `docs/` holds system documentation and suite reference material

## Documentation Layout

The repo keeps contributor-facing documentation in two places:

- system-wide documentation lives in [docs](docs)
- suite-level documentation lives in [docs/suites](docs/suites)
- inline godoc lives in the Go source

There is no `services/` documentation tree. Service-level details are part of each suite
document under `docs/suites/*.md`.

## Workspace Layout

```text
~/gradient/
  data/
  notebooks/
  models/
  outputs/
  mlruns/
  dags/
  compose/
  config/
  backups/
```

## Quick Start

Ubuntu 24.04 package install:

```bash
curl -fsSL https://packages.gradientlinux.io/install.sh | sudo bash
concave setup
```

Manual local build:

```bash
go build -o concave .
./concave doctor
./concave workspace init
```

`concave --verbose` enables structured debug logging on stderr without changing stdout.
`scripts/build.sh` builds the static binary and generates shell completions into
`scripts/completions/`.

Phase 3 adds multi-user and frontend-facing control surfaces:

```bash
concave whoami
concave serve --addr 127.0.0.1:7777
```

`concave serve` is intended to run under the packaged `concave-serve.service`
systemd unit as `gradient-svc`. Authentication and authorization are derived from
Unix group membership:

- `gradient-viewer`
- `gradient-developer`
- `gradient-operator`
- `gradient-admin`

## Suite Reference

- [Boosting](docs/suites/boosting.md): CPU-first experimentation, JupyterLab, MLflow
- [Neural](docs/suites/neural.md): GPU-oriented training, inference, notebooks
- [Flow](docs/suites/flow.md): tracking, orchestration, storage, dashboards, serving
- [Forge](docs/suites/forge.md): user-selected composition of components from other suites

See [docs/suite-guide.md](docs/suite-guide.md) for the high-level suite map and
[docs/concave-reference.md](docs/concave-reference.md) for command coverage.
See [docs/system-admin.md](docs/system-admin.md) for auth groups, `concave serve`,
and packaged service behavior.

## Companion Projects

- `concave-tui` is maintained as a separate repository so the infrastructure CLI can
  stay headless-safe and free of Bubble Tea dependencies.

## Contributing

Contributor expectations, repository conventions, and pull request rules live in
[CONTRIBUTING.md](CONTRIBUTING.md). Maintainers may use additional private automation or
internal workflows, but the public contribution contract is defined here in the repo.

## License

This project is released under the [MIT License](LICENSE).
