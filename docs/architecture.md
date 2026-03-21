# Architecture

Gradient Linux ships a thin Ubuntu 24.04 base with Docker Engine and the `concave`
binary. `concave` owns the workspace, host checks, Compose rendering, suite lifecycle,
GPU inspection, and rollback metadata.

## Design Goals

- keep the host clean: no host Python package installation
- keep user data outside containers in `~/gradient/`
- make suite operations repeatable and reversible
- keep output consistent through a single UI layer
- make Docker and GPU interactions testable through command seams

## Runtime Model

1. `concave doctor` inspects Docker, network reachability, workspace state, and GPU
   support.
2. `concave workspace init` creates the canonical `~/gradient/` tree.
3. The suite registry defines containers, ports, mounts, and GPU requirements.
4. `internal/docker/compose.go` reads a suite template from `templates/`, substitutes
   `{{WORKSPACE_ROOT}}` and `{{COMPOSE_NETWORK}}`, and writes the rendered Compose file
   into `~/gradient/compose/`.
5. Docker image pulls and `docker compose` lifecycle actions are executed through the
   `internal/docker/` package.
6. Installed suite state and image version history are recorded in
   `~/gradient/config/state.json` and `~/gradient/config/versions.json`.
7. Rollback and update operations manipulate Compose output and config metadata only;
   they never modify user datasets, notebooks, or model files.

## Documentation Model

This repository keeps documentation in two places only:

- `docs/` for system-wide and suite-level prose
- inline godoc in Go source

Suite docs live in `docs/suites/*.md`. Each suite doc includes the relevant
container-level details for that suite. There is no standalone `services/` tree.

## Ownership Resolution

This implementation treats `cmd/lab.go` as suite-owned behavior and
`internal/system/ports.go` as infra-owned behavior to resolve the overlap called out in
`AGENTS.md`.
