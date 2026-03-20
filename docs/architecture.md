# Architecture

Gradient Linux ships a thin Ubuntu 24.04 base with Docker Engine and the `concave` CLI. The CLI owns the workspace, detects host capabilities, renders suite-specific Compose files, and manages container lifecycle actions.

## Design Goals

- Keep the host clean: no host Python package installation
- Keep user data outside containers in `~/gradient/`
- Make suite operations repeatable and reversible
- Keep command output human-readable and consistent

## Runtime Model

1. `concave` checks host prerequisites.
2. `concave` initializes `~/gradient/`.
3. A suite registry defines containers, ports, and mounts.
4. A Compose template is rendered into `~/gradient/compose/`.
5. Docker pulls images and runs the suite.
6. State and image versions are recorded under `~/gradient/config/`.

## Ownership Resolution

This implementation treats `cmd/lab.go` as suite-owned behavior and `internal/system/ports.go` as infra-owned behavior, which resolves the overlaps in `AGENTS.md`.
