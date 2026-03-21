# Gradient Linux — concave

`concave` is the control-plane CLI for Gradient Linux, an Ubuntu 24.04 LTS distribution
built for machine learning engineers, data scientists, and MLOps teams. The host stays
thin: Ubuntu, Docker Engine, and the `concave` binary. Suites, models, notebooks,
tracking, orchestration, and observability stay inside containers.

## What This Repo Contains

- a statically compiled Go CLI
- Docker Compose templates for Boosting, Neural, Flow, and Forge
- workspace lifecycle management for `~/gradient/`
- suite install, start, stop, update, rollback, and status flows
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

```bash
go build -o concave .
./concave doctor
./concave workspace init
```

## Suite Reference

- [Boosting](docs/suites/boosting.md): CPU-first experimentation, JupyterLab, MLflow
- [Neural](docs/suites/neural.md): GPU-oriented training, inference, notebooks
- [Flow](docs/suites/flow.md): tracking, orchestration, storage, dashboards, serving
- [Forge](docs/suites/forge.md): user-selected composition of components from other suites

See [docs/suite-guide.md](docs/suite-guide.md) for the high-level suite map and
[docs/concave-reference.md](docs/concave-reference.md) for command coverage.

## Contributing

Contributor expectations, repository conventions, and pull request rules live in
[CONTRIBUTING.md](CONTRIBUTING.md). Maintainers may use additional private automation or
internal workflows, but the public contribution contract is defined here in the repo.

## License

License terms are pending project publication.
