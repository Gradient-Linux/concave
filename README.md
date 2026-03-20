# Gradient Linux — concave

`concave` is the control plane CLI for Gradient Linux, an Ubuntu 24.04 LTS distribution for machine learning engineers and MLOps teams. The host stays thin: Docker plus the `concave` binary. Suites, models, notebooks, and orchestration tooling live inside containers managed by the CLI.

## What is this?

This repository builds:

- `concave`, a statically compiled Go CLI
- Docker Compose templates for Boosting, Neural, Flow, and Forge suites
- Workspace management for `~/gradient/`
- GPU detection and driver setup logic
- Operational documentation for every suite and every service container

## Architecture

- `cmd/` contains Cobra commands
- `internal/ui/` standardizes all terminal output and prompts
- `internal/system/` checks host prerequisites and opens browsers
- `internal/workspace/` manages `~/gradient/`
- `internal/docker/` wraps Docker and Compose
- `internal/suite/` defines suites and lifecycle helpers
- `internal/config/` persists state and image versions
- `internal/gpu/` handles hardware detection and driver guidance
- `templates/` stores canonical Compose templates
- `docs/` holds system, suite, and service documentation

## Workspace Layout

```
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

## Suites

- Boosting: CPU-first ML workflow with JupyterLab and MLflow
- Neural: GPU training and inference suite
- Flow: MLOps stack with orchestration, tracking, storage, and observability
- Forge: Custom composition of services from the other suites

Suite docs live under [docs/suites](docs/suites). Service docs live under [docs/services](docs/services).

## Command Reference

The authoritative CLI reference lives in [docs/concave-reference.md](docs/concave-reference.md).

## GPU Setup

GPU setup and driver workflow notes live in [docs/gpu-setup.md](docs/gpu-setup.md).

## Contributing

This repo follows the engineering workflow described in [AGENTS.md](AGENTS.md). The codebase also keeps documentation as a first-class deliverable: when behavior changes, the matching suite and service docs change with it.

## License

License terms are pending project publication.
