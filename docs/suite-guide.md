# Suite Guide

`concave` manages four suite targets:

- `boosting`
- `neural`
- `flow`
- `forge`

Each suite is defined in `internal/suite/registry.go`, rendered to `~/gradient/compose/<suite>.compose.yml`, and tracked under `~/gradient/config/`.

## Install a suite

Install writes the suite state, records image versions, and renders the Compose file. It does not start the containers.

```bash
concave install boosting
```

If the suite is already present, rerun with:

```bash
concave install boosting --force
```

## Start and stop

Start one suite:

```bash
concave start boosting
```

Start every installed suite:

```bash
concave start
```

Stop one suite:

```bash
concave stop boosting
```

Stop every installed suite:

```bash
concave stop
```

Restart one suite:

```bash
concave restart boosting
```

## Observe the current state

Use these commands while the suite is running:

```bash
concave status
concave list
concave logs boosting --service gradient-boost-lab
concave lab --suite boosting
```

`concave lab` opens JupyterLab for `boosting` or `neural`. If `--suite` is omitted, `concave` picks the first installed Jupyter-capable suite.

## Update and rollback

Update pulls the target images from the suite registry, records the previous image tags, rewrites the Compose file, and restarts the suite:

```bash
concave update boosting
```

Rollback swaps the current and previous image tags recorded in `~/gradient/config/versions.json`, rewrites the Compose file, and restarts the suite:

```bash
concave rollback boosting
```

The current build uses `versions.json` for rollback state. A content-addressed `gradient.lock` file is planned for a later release and is not shipped in this build.

Rollback and remove preserve user data:

- `~/gradient/data/`
- `~/gradient/notebooks/`
- `~/gradient/models/`

## Remove a suite

Removal tears down the Compose stack, deletes the suite state, and keeps user data in place:

```bash
concave remove boosting
```

The command asks for confirmation before it proceeds.

## Example workflow: from zero to JupyterLab

```bash
concave workspace init
concave install boosting
concave start boosting
concave lab --suite boosting
concave status
```

That flow creates the workspace, installs the CPU-first suite, starts it, opens JupyterLab, and then prints the suite status table.

## Per-suite details

Use the suite-specific documents for images, ports, and mounts:

- [suites/boosting.md](suites/boosting.md)
- [suites/neural.md](suites/neural.md)
- [suites/flow.md](suites/flow.md)
- [suites/forge.md](suites/forge.md)
