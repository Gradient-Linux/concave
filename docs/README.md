# concave docs

This directory contains the operator and contributor reference material for
`concave`.

## Core docs

- [architecture.md](architecture.md): repository structure, runtime boundaries, and
  the package-level split between CLI, server, Docker, workspace, auth, and suite
  management.
- [api.md](api.md): the `concave serve` control plane, authentication model, route
  groups, job model, and terminal endpoints.
- [concave-reference.md](concave-reference.md): command reference for the CLI.
- [system-admin.md](system-admin.md): Unix groups, packaged service behavior,
  `gradient-svc`, sudoers scope, and machine administration notes.
- [gpu-setup.md](gpu-setup.md): GPU detection and driver/toolkit guidance.
- [suite-guide.md](suite-guide.md): suite selection guide and high-level use cases.

## Suite reference

- [suites/boosting.md](suites/boosting.md)
- [suites/neural.md](suites/neural.md)
- [suites/flow.md](suites/flow.md)
- [suites/forge.md](suites/forge.md)

## How to use this tree

- Start with [architecture.md](architecture.md) if you need to understand the codebase.
- Read [system-admin.md](system-admin.md) if you are deploying `concave` on a real
  machine.
- Read [api.md](api.md) if you are integrating `concave-web`, `concave-tui`, or any
  other client against `concave serve`.
